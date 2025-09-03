package main

import "fmt"

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
