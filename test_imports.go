package main

import (
	"fmt"
	"time"
	
	"github.com/qdrant/go-client/qdrant"
)

func main() {
	fmt.Println("Testing qdrant import...")
	start := time.Now()
	
	fmt.Printf("Qdrant import took: %v\n", time.Since(start))
	// fmt.Printf(qdrant)
	
	// Test creating client (this might be slow)
	start2 := time.Now()
	_, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		fmt.Printf("Client creation failed: %v\n", err)
	}
	fmt.Printf("Qdrant client creation took: %v\n", time.Since(start2))
}
