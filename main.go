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
	node0.Store(10, "Hello Chord")
	time.Sleep(2 * time.Second)
	ring.Nodes[2].Store(20, "World Chord")
	time.Sleep(2 * time.Second)
	ring.Nodes[2].Store(4, "Chord Network")
	time.Sleep(2 * time.Second)
	node0.Retrieve(20)
	time.Sleep(2 * time.Second)

	ring.Nodes[2].Store(17, "Ola 17")
	time.Sleep(2 * time.Second)

	ring.Print()
}
