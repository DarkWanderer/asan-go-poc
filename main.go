package main

/*
#include <sanitizer/asan_interface.h>
#include <stdlib.h>
#include <string.h>

// Helper function to allocate memory and copy string data
char* allocate_string(const char* data, size_t len) {
    char* ptr = (char*)malloc(len + 1);
    if (ptr != NULL) {
        memcpy(ptr, data, len);
        ptr[len] = '\0';
    }
    return ptr;
}

// Helper function to poison memory
void poison_memory(void* ptr, size_t size) {
    __asan_poison_memory_region(ptr, size);
}

// Helper function to unpoison memory
void unpoison_memory(void* ptr, size_t size) {
    __asan_unpoison_memory_region(ptr, size);
}

// Helper function to free memory
void free_memory(void* ptr) {
    free(ptr);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// PooledString represents a string that can be returned to the pool
type PooledString struct {
	data   unsafe.Pointer
	length int
	pool   *StringPool
}

// String returns the string value (unsafe if string has been returned to pool)
func (ps *PooledString) String() string {
	if ps.data == nil {
		return ""
	}
	// Convert C string to Go string
	return C.GoStringN((*C.char)(ps.data), C.int(ps.length))
}

// Return gives the string back to the pool and poisons the memory
func (ps *PooledString) Return() {
	if ps.data != nil && ps.pool != nil {
		ps.pool.returnString(ps)
	}
}

// StringPool manages a pool of strings with AddressSanitizer integration
type StringPool struct {
	available []unsafe.Pointer
	sizes     []int
	mutex     sync.Mutex
}

// NewStringPool creates a new string pool
func NewStringPool() *StringPool {
	return &StringPool{
		available: make([]unsafe.Pointer, 0),
		sizes:     make([]int, 0),
	}
}

// Rent gets a string from the pool or allocates a new one
func (sp *StringPool) Rent(data string) *PooledString {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	dataBytes := []byte(data)
	needed := len(dataBytes)

	// Try to reuse existing memory
	for i, ptr := range sp.available {
		if sp.sizes[i] >= needed {
			// Remove from available pool
			sp.available = append(sp.available[:i], sp.available[i+1:]...)
			size := sp.sizes[i]
			sp.sizes = append(sp.sizes[:i], sp.sizes[i+1:]...)

			// Unpoison the memory before use
			C.unpoison_memory(ptr, C.size_t(size+1))

			// Copy new data
			C.memcpy(ptr, unsafe.Pointer(&dataBytes[0]), C.size_t(needed))
			*(*C.char)(unsafe.Pointer(uintptr(ptr) + uintptr(needed))) = 0

			return &PooledString{
				data:   ptr,
				length: needed,
				pool:   sp,
			}
		}
	}

	// Allocate new memory
	cstr := C.CString(data)
	defer C.free(unsafe.Pointer(cstr))

	ptr := C.allocate_string(cstr, C.size_t(needed))
	if ptr == nil {
		return nil
	}

	return &PooledString{
		data:   unsafe.Pointer(ptr),
		length: needed,
		pool:   sp,
	}
}

// returnString returns a string to the pool and poisons its memory
func (sp *StringPool) returnString(ps *PooledString) {
	if ps.data == nil {
		return
	}

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Poison the memory to detect use-after-free
	size := ps.length
	C.poison_memory(ps.data, C.size_t(size+1))

	// Add to available pool
	sp.available = append(sp.available, ps.data)
	sp.sizes = append(sp.sizes, size)

	// Clear the PooledString to prevent further use
	ps.data = nil
	ps.length = 0
}

// Cleanup frees all memory in the pool
func (sp *StringPool) Cleanup() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, ptr := range sp.available {
		C.free_memory(ptr)
	}
	sp.available = nil
	sp.sizes = nil
}

// Demonstration functions

func demonstrateCorrectUsage() {
	fmt.Println("=== Demonstrating Correct Usage ===")

	pool := NewStringPool()
	defer pool.Cleanup()

	// Rent a string
	ps := pool.Rent("Hello, World!")
	fmt.Printf("Rented string: %s\n", ps.String())

	// Return it immediately
	ps.Return()
	fmt.Println("String returned to pool successfully")
}

func demonstrateUseAfterFree() {
	fmt.Println("\n=== Demonstrating Use-After-Free Bug ===")

	pool := NewStringPool()
	defer pool.Cleanup()

	// Rent a string
	ps := pool.Rent("This will cause a bug!")
	fmt.Printf("Rented string: %s\n", ps.String())

	// Keep a reference to the string value (simulating retaining pointer)
	buggyRef := ps

	// Return the string to pool (this will poison the memory)
	ps.Return()
	fmt.Println("String returned to pool")

	// Force garbage collection to ensure memory operations are complete
	runtime.GC()

	fmt.Println("Attempting to access returned string...")
	fmt.Println("This should trigger AddressSanitizer error:")

	// This access should be detected by AddressSanitizer as use-after-free
	// The memory has been poisoned, so ASAN should report an error
	_ = buggyRef.String() // This line should trigger ASAN error
}

func demonstratePoolReuse() {
	fmt.Println("\n=== Demonstrating Pool Reuse ===")

	pool := NewStringPool()
	defer pool.Cleanup()

	// Rent and return a string
	ps1 := pool.Rent("First string")
	fmt.Printf("First rental: %s\n", ps1.String())
	ps1.Return()

	// Rent another string (should reuse memory)
	ps2 := pool.Rent("Second")
	fmt.Printf("Second rental (reused memory): %s\n", ps2.String())
	ps2.Return()

	fmt.Println("Pool reuse demonstrated successfully")
}

func main() {
	fmt.Println("AddressSanitizer String Pool Proof of Concept")
	fmt.Println("=============================================")
	fmt.Println("Note: Run with 'go run -race' and AddressSanitizer enabled")
	fmt.Println("Expected: ASAN should detect use-after-free in the second demonstration")

	demonstrateCorrectUsage()
	demonstratePoolReuse()

	// This demonstration should trigger AddressSanitizer
	demonstrateUseAfterFree()

	fmt.Println("\nDemo completed - check for AddressSanitizer reports above")
}
