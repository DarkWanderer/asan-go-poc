//go:build !asan
// +build !asan

package main

import (
	"fmt"
	"unsafe"
)

func poisonMemory(ptr unsafe.Pointer, size int) {
	fmt.Printf("No-op: would poison memory at %p, size: %d\n", ptr, size)
}

func unpoisonMemory(ptr unsafe.Pointer, size int) {
	fmt.Printf("No-op: would unpoison memory at %p, size: %d\n", ptr, size)
}