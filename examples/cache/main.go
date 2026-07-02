package main

import (
	"fmt"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/cache"
)

func main() {
	// 1. Create a new cache for strings
	c := cache.New[string, string]()

	// 2. Start the background cleanup goroutine
	c.StartCleanup(1 * time.Second)
	defer c.StopCleanup() // Ensure cleanup stops when main exits

	// 3. Set a value with a 2-second TTL
	fmt.Println("Setting 'session' to 'user123' with 2s TTL...")
	c.SetWithTTL("session", "user123", 2*time.Second)

	// 4. Retrieve the value immediately
	if val, ok := c.Get("session"); ok {
		fmt.Printf("Cache hit: session = %s\n", val)
	}

	// 5. Wait for the TTL to expire
	fmt.Println("Waiting 3 seconds for TTL to expire...")
	time.Sleep(3 * time.Second)

	// 6. Attempt to retrieve the expired value
	if _, ok := c.Get("session"); !ok {
		fmt.Println("Cache miss: session has expired and been cleaned up!")
	}
}
