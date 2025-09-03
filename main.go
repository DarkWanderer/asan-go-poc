package main

/*
#cgo CFLAGS: -fsanitize=address
#cgo LDFLAGS: -fsanitize=address
#include <sanitizer/asan_interface.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Helper function to poison memory
void poison_memory(void* ptr, size_t size) {
    printf("Poisoning memory at %p, size: %zu\n", ptr, size);
    __asan_poison_memory_region(ptr, size);
}

// Helper function to unpoison memory
void unpoison_memory(void* ptr, size_t size) {
    printf("Unpoisoning memory at %p, size: %zu\n", ptr, size);
    __asan_unpoison_memory_region(ptr, size);
}

*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

const BufferSize = 1024

// BufferPool manages a pool of 1024-byte buffers with poisoning/unpoisoning
type BufferPool struct {
	pool chan *[BufferSize]byte
	mu   sync.Mutex
}

// NewBufferPool creates a new buffer pool with the specified capacity
func NewBufferPool(capacity int) *BufferPool {
	return &BufferPool{
		pool: make(chan *[BufferSize]byte, capacity),
	}
}

// Rent gets a buffer from the pool, unpoisoning it for use
func (bp *BufferPool) Rent() *[BufferSize]byte {
	select {
	case buf := <-bp.pool:
		// Unpoison the buffer before returning it
		C.unpoison_memory(unsafe.Pointer(buf), C.size_t(BufferSize))
		fmt.Printf("Rented buffer at %p (unpoisoned)\n", buf)
		return buf
	default:
		// No buffers available, create a new one
		buf := &[BufferSize]byte{}
		fmt.Printf("Created new buffer at %p\n", buf)
		return buf
	}
}

// Return puts a buffer back into the pool, poisoning it to detect use-after-free
func (bp *BufferPool) Return(buf *[BufferSize]byte) {
	if buf == nil {
		return
	}
	
	// Poison the buffer before putting it back in the pool
	C.poison_memory(unsafe.Pointer(buf), C.size_t(BufferSize))
	fmt.Printf("Returned buffer at %p (poisoned)\n", buf)
	
	select {
	case bp.pool <- buf:
		// Successfully returned to pool
	default:
		// Pool is full, let the buffer be garbage collected
		fmt.Printf("Pool full, buffer at %p will be garbage collected\n", buf)
	}
}

func main() {
	fmt.Println("=== ASan Buffer Pool Demonstration ===")
	
	// Create a buffer pool with capacity for 2 buffers
	pool := NewBufferPool(2)
	
	// Rent a buffer from the pool
	fmt.Println("\n1. Renting a buffer...")
	buffer := pool.Rent()
	
	// Write some data to the buffer
	fmt.Println("\n2. Writing data to buffer...")
	testData := "Hello, ASan!"
	copy(buffer[:], testData)
	fmt.Printf("Written: %s\n", string(buffer[:len(testData)]))
	
	// Save a pointer to the buffer for later (this simulates keeping a dangling pointer)
	fmt.Println("\n3. Saving pointer to buffer...")
	savedPtr := buffer
	fmt.Printf("Saved pointer: %p\n", savedPtr)
	
	// Return the buffer to the pool (this will poison it)
	fmt.Println("\n4. Returning buffer to pool (poisoning)...")
	pool.Return(buffer)
	
	// Try to read from the saved pointer - this should trigger ASan violation
	fmt.Println("\n5. Attempting to read from saved pointer (should trigger ASan)...")
	fmt.Printf("Trying to read: %s\n", string(savedPtr[:len(testData)]))
	
	fmt.Println("\nDemo completed!")
}
