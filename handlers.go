package main

import (
	"log"
	"net"
)

func (node *Node) handleStore(msg Message) {
	originalKey := msg.Key
	hashedKey := originalKey % RingSize
	if node.Predecessor == nil {
		node.mutex.Lock()
		node.Data[originalKey] = msg.Value
		node.mutex.Unlock()
		println("Node", node.ID, "stored key", originalKey, "with value '"+msg.Value+"'")
		return
	}
	if IsInInterval(hashedKey, node.Predecessor.ID, node.ID, true) {
		node.mutex.Lock()
		node.Data[originalKey] = msg.Value
		node.mutex.Unlock()
		println("Node", node.ID, "stored key", originalKey, "with value '"+msg.Value+"'")
	} else {
		next := node.closestPrecedingNode(hashedKey)
		node.sendMessage(msg, next.Address)
	}
}

func (node *Node) handleRetrieve(msg Message, addr *net.UDPAddr) {
	originalKey := msg.Key
	hashedKey := originalKey % RingSize
	if node.Predecessor == nil {
		node.mutex.Lock()
		value, exists := node.Data[originalKey]
		node.mutex.Unlock()
		if exists {
			reply := Message{
				Type:       "RETRIEVE_REPLY",
				Key:        originalKey,
				Value:      value,
				SenderID:   node.ID,
				Address:    node.Address,
				OriginID:   msg.OriginID,
				OriginAddr: msg.OriginAddr,
			}
			node.sendMessage(reply, msg.OriginAddr)
		} else {
			log.Printf("Key %d not found at Node %d", originalKey, node.ID)
		}
		return
	}
	if IsInInterval(hashedKey, node.Predecessor.ID, node.ID, true) {
		node.mutex.Lock()
		value, exists := node.Data[originalKey]
		node.mutex.Unlock()
		if exists {
			reply := Message{
				Type:       "RETRIEVE_REPLY",
				Key:        originalKey,
				Value:      value,
				SenderID:   node.ID,
				Address:    node.Address,
				OriginID:   msg.OriginID,
				OriginAddr: msg.OriginAddr,
			}
			node.sendMessage(reply, msg.OriginAddr)
		} else {
			log.Printf("Key %d not found at Node %d", originalKey, node.ID)
		}
	} else {
		next := node.closestPrecedingNode(hashedKey)
		msg.Address = addr
		node.sendMessage(msg, next.Address)
	}
}

func (node *Node) handleRetrieveReply(msg Message) {
	if msg.OriginID == node.ID {
		println("Node", node.ID, "received retrieve reply for key", msg.Key, ":'"+msg.Value+"'")
	}
}

func (node *Node) handleFindSuccessorMsg(msg Message) {
	originalKey := msg.Key
	hashedKey := originalKey % RingSize
	node.mutex.Lock()
	successor := node.Successor
	node.mutex.Unlock()
	if successor == nil {
		response := Message{
			Type:      "FIND_SUCCESSOR_REPLY",
			Key:       originalKey,
			SenderID:  node.ID,
			Address:   node.Address,
			RequestID: msg.RequestID,
		}
		node.sendMessage(response, msg.Address)
		return
	}
	if IsInInterval(hashedKey, node.ID, successor.ID, true) {
		response := Message{
			Type:      "FIND_SUCCESSOR_REPLY",
			Key:       originalKey,
			SenderID:  successor.ID,
			Address:   successor.Address,
			RequestID: msg.RequestID,
		}
		node.sendMessage(response, msg.Address)
	} else {
		n0 := node.closestPrecedingNode(hashedKey)
		if n0.ID == node.ID && !IsInInterval(hashedKey, node.ID, successor.ID, true) {
			response := Message{
				Type:      "FIND_SUCCESSOR_REPLY",
				Key:       originalKey,
				SenderID:  successor.ID,
				Address:   successor.Address,
				RequestID: msg.RequestID,
			}
			node.sendMessage(response, msg.Address)
			return
		}
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
	if node.Predecessor == nil || (node.Predecessor.ID == node.ID && node.Successor.ID == node.ID) {
		node.Predecessor = &NodeInfo{
			ID:      predID,
			Address: msg.Address,
		}
		if node.Successor.ID == node.ID {
			node.Successor = &NodeInfo{
				ID:      predID,
				Address: msg.Address,
			}
		}
		return
	}
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
