package main

type FingerEntry struct {
	Start int
	Node  *NodeInfo
}

func (node *Node) InitializeFingerTable() {
	for i := 0; i < FTSize; i++ {
		start := (node.ID + (1 << i)) % RingSize
		successor := node.findSuccessor(start)
		node.FingerTable[i] = &FingerEntry{
			Start: start,
			Node:  successor,
		}
	}
	node.Successor = node.FingerTable[0].Node
}
func (node *Node) closestPrecedingNode(key int) *NodeInfo {
	key = key % RingSize
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for i := FTSize - 1; i >= 0; i-- {
		finger := node.FingerTable[i]
		if finger != nil && IsInInterval(finger.Node.ID, node.ID, key, false) && finger.Node.ID != node.ID {
			return finger.Node
		}
	}
	if node.Successor != nil && node.Successor.ID != node.ID {
		return node.Successor
	}
	return NodeToNodeInfo(node)
}
