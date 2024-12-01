package main

import "time"

func main() {
	ring := NewRing()

	// Add more nodes
	for i := 1; i < RingSize; i++ {
		ring.AddNode(i)
		time.Sleep(500 * time.Millisecond) // Wait for stabilization
	}

	// Allow time for stabilization
	time.Sleep(5 * time.Second)

	// Simulate storing and retrieving data
	node := ring.Nodes[0]
	node.Store(10, "Hello Chord")
	time.Sleep(1 * time.Second) // Wait for the message to propagate

	ring.Nodes[2].Store(20, "World Chord")
	time.Sleep(1 * time.Second)
	ring.Nodes[2].Store(4, "Chord Network")

	node.Retrieve(20)
	time.Sleep(1 * time.Second) // Wait for the response

	ring.Print()
}
