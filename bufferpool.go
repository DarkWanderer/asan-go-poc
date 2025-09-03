package main

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
		unpoisonMemory(unsafe.Pointer(buf), BufferSize)
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
	poisonMemory(unsafe.Pointer(buf), BufferSize)
	fmt.Printf("Returned buffer at %p (poisoned)\n", buf)

	select {
	case bp.pool <- buf:
		// Successfully returned to pool
	default:
		// Pool is full, let the buffer be garbage collected
		fmt.Printf("Pool full, buffer at %p will be garbage collected\n", buf)
	}
}