package main

import (
	"fmt"
	"log"
)

func (node *Node) handleStore(msg Message) {
	if node.Predecessor == nil {
		node.mutex.Lock()
		node.Data[msg.Key] = msg.Value
		node.mutex.Unlock()
		fmt.Printf("Node %d stored key %d with value '%s'\n", node.ID, msg.Key, msg.Value)
		return
	}

	if IsInInterval(msg.Key, node.Predecessor.ID, node.ID, true) {
		node.mutex.Lock()
		node.Data[msg.Key] = msg.Value
		node.mutex.Unlock()
		fmt.Printf("Node %d stored key %d with value '%s'\n", node.ID, msg.Key, msg.Value)
	} else {
		next := node.closestPrecedingNode(msg.Key)
		node.sendMessage(msg, next.Address)
	}
}

func (node *Node) handleRetrieve(msg Message) {
	if node.Predecessor == nil {
		node.mutex.Lock()
		value, exists := node.Data[msg.Key]
		node.mutex.Unlock()
		if exists {
			fmt.Printf("Node %d retrieved value for key %d: '%s'\n", node.ID, msg.Key, value)
		} else {
			log.Printf("Key %d not found at Node %d", msg.Key, node.ID)
		}
		return
	}

	if IsInInterval(msg.Key, node.Predecessor.ID, node.ID, true) {
		node.mutex.Lock()
		value, exists := node.Data[msg.Key]
		node.mutex.Unlock()
		if exists {
			fmt.Printf("Node %d retrieved value for key %d: '%s'\n", node.ID, msg.Key, value)
		} else {
			log.Printf("Key %d not found at Node %d", msg.Key, node.ID)
		}
	} else {
		next := node.closestPrecedingNode(msg.Key)
		node.sendMessage(msg, next.Address)
	}
}

func (node *Node) handleFindSuccessor(msg Message) {
	node.mutex.Lock()
	successor := node.Successor
	node.mutex.Unlock()

	if successor == nil {
		response := Message{
			Type:      "FIND_SUCCESSOR_REPLY",
			Key:       msg.Key,
			SenderID:  node.ID,
			Address:   node.Address,
			RequestID: msg.RequestID,
		}
		node.sendMessage(response, msg.Address)
		return
	}

	if IsInInterval(msg.Key, node.ID, successor.ID, true) {
		response := Message{
			Type:      "FIND_SUCCESSOR_REPLY",
			Key:       msg.Key,
			SenderID:  successor.ID,
			Address:   successor.Address,
			RequestID: msg.RequestID,
		}
		node.sendMessage(response, msg.Address)
	} else {
		n0 := node.closestPrecedingNode(msg.Key)
		node.sendMessage(msg, n0.Address)
	}
}

func (node *Node) handleFindSuccessorReply(msg Message) {
	node.requestMutex.Lock()
	responseChan, exists := node.pendingRequests[msg.RequestID]
	if exists {
		responseChan <- msg
	}
	node.requestMutex.Unlock()
}

func (node *Node) handleNotify(msg Message) {
	predID := msg.SenderID
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.Predecessor == nil || IsInInterval(predID, node.Predecessor.ID, node.ID, false) {
		node.Predecessor = &NodeInfo{
			ID:      predID,
			Address: msg.Address,
		}
	}
}

func (node *Node) handleGetPredecessor(msg Message) {
	node.mutex.Lock()
	var pred *NodeInfo
	if node.Predecessor != nil {
		pred = node.Predecessor
	}
	node.mutex.Unlock()

	var response Message
	if pred != nil {
		response = Message{
			Type:      "GET_PREDECESSOR_REPLY",
			SenderID:  pred.ID,
			Address:   pred.Address,
			RequestID: msg.RequestID,
		}
	} else {
		response = Message{
			Type:      "GET_PREDECESSOR_REPLY",
			SenderID:  node.ID,
			Address:   node.Address,
			RequestID: msg.RequestID,
		}
	}
	node.sendMessage(response, msg.Address)
}

func (node *Node) handleGetPredecessorReply(msg Message) {
	node.requestMutex.Lock()
	responseChan, exists := node.pendingRequests[msg.RequestID]
	if exists {
		responseChan <- msg
	}
	node.requestMutex.Unlock()
}
