//go:build asan
// +build asan

package main

/*
#cgo CFLAGS: -fsanitize=address
#cgo LDFLAGS: -fsanitize=address
#include <sanitizer/asan_interface.h>
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
import "unsafe"

func poisonMemory(ptr unsafe.Pointer, size int) {
	C.poison_memory(ptr, C.size_t(size))
}

func unpoisonMemory(ptr unsafe.Pointer, size int) {
	C.unpoison_memory(ptr, C.size_t(size))
}