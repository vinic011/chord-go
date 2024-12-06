package main

import (
	"time"
)

func main() {
	ring := NewRing()
	for i := 1; i < RingSize; i++ {
		ring.AddNode(i)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(10 * time.Second)
	node0 := ring.Nodes[0]
	node0.Store(6, "Hello Chord")
	time.Sleep(2 * time.Second)
	ring.Nodes[2].Store(1, "World Chord")
	time.Sleep(2 * time.Second)
	ring.Nodes[2].Store(4, "Chord Network")
	time.Sleep(2 * time.Second)
	node0.Retrieve(4)
	time.Sleep(2 * time.Second)

	ring.Nodes[2].Store(10, "Chord Network 10")
	time.Sleep(2 * time.Second)

	ring.Nodes[2].Store(18, "Chord Network 20")
	time.Sleep(2 * time.Second)

	ring.Nodes[2].Store(5, "Ola 17")
	time.Sleep(2 * time.Second)

	ring.Print()
}
