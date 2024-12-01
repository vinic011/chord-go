package main

import (
	"fmt"
	"strconv"
)

type Ring struct {
	Nodes []*Node
}

func NewRing() *Ring {
	ring := &Ring{
		Nodes: make([]*Node, 0),
	}

	firstNode := NewNode(0, strconv.Itoa(10000))
	firstNode.Join(nil)
	ring.Nodes = append(ring.Nodes, firstNode)

	return ring
}

func (ring *Ring) AddNode(id int) {
	port := strconv.Itoa(10000 + id)
	newNode := NewNode(id, port)
	if len(ring.Nodes) > 0 {
		newNode.Join(ring.Nodes[0])
	} else {
		newNode.Join(nil)
	}
	ring.Nodes = append(ring.Nodes, newNode)
}

func (ring *Ring) Print() {
	for _, node := range ring.Nodes {
		node.mutex.Lock()
		successorID := -1
		predecessorID := -1
		if node.Successor != nil {
			successorID = node.Successor.ID
		}
		if node.Predecessor != nil {
			predecessorID = node.Predecessor.ID
		}
		fmt.Printf("Node %d: Successor -> %d, Predecessor -> %d\n", node.ID, successorID, predecessorID)
		for key, value := range node.Data {
			fmt.Printf("  Key %d -> Value '%s'\n", key, value)
		}
		node.mutex.Unlock()
	}
}
