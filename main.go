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
	"unsafe"
)

func main() {
	// Create byte arrays instead of strings
	var s = "Hello World"
	var ptr = unsafe.StringData(s)

	fmt.Printf("Original byte array: %v\n", ptr)

	// Get pointer to the byte array data

	// Poison the memory region
	C.poison_memory((unsafe.Pointer)(ptr), C.size_t(len(s)))

	// Try to access the poisoned memory - this should trigger ASAN
	fmt.Printf("Accessing poisoned memory: %v\n", s)

	// Unpoison the memory
	C.unpoison_memory((unsafe.Pointer)(ptr), C.size_t(len(s)))

	// Access should work now
	fmt.Printf("After unpoisoning: %v\n", s)
}
