package main

import "math"

type FingerEntry struct {
	Start int
	Node  *NodeInfo
}

func (node *Node) InitializeFingerTable() {
	for i := 0; i < FTSize; i++ {
		start := (node.ID + int(math.Pow(2, float64(i)))) % RingSize
		successor := node.findSuccessor(start)
		node.FingerTable[i] = &FingerEntry{
			Start: start,
			Node:  successor,
		}
	}
	node.Successor = node.FingerTable[0].Node
}
