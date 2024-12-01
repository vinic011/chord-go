package main

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type Message struct {
	Type       string
	Key        int
	Value      string
	SenderID   int
	ReceiverID int
	OriginID   int
	Address    *net.UDPAddr
	RequestID  int
}

type NodeInfo struct {
	ID      int
	Address *net.UDPAddr
}

type Node struct {
	ID          int
	Address     *net.UDPAddr
	FingerTable []*FingerEntry
	Successor   *NodeInfo
	Predecessor *NodeInfo
	Data        map[int]string
	Conn        *net.UDPConn

	mutex           sync.Mutex
	requestMutex    sync.Mutex
	pendingRequests map[int]chan Message
	nextRequestID   int
}

func NewNode(id int, port string) *Node {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port: %v", err)
	}

	node := &Node{
		ID:              id,
		Address:         addr,
		FingerTable:     make([]*FingerEntry, FTSize),
		Data:            make(map[int]string),
		Conn:            conn,
		pendingRequests: make(map[int]chan Message),
	}

	go node.listen()
	go node.Stabilize()
	go node.FixFingers()

	return node
}

func (node *Node) listen() {
	buf := make([]byte, 4096)
	for {
		n, addr, err := node.Conn.ReadFromUDP(buf)

		if err != nil {
			log.Printf("Error reading from UDP: %v %v", err, addr)
			continue
		}

		var msg Message
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		switch msg.Type {
		case "STORE":
			node.handleStore(msg)
		case "RETRIEVE":
			node.handleRetrieve(msg)
		case "FIND_SUCCESSOR":
			node.handleFindSuccessor(msg)
		case "FIND_SUCCESSOR_REPLY":
			node.handleFindSuccessorReply(msg)
		case "NOTIFY":
			node.handleNotify(msg)
		case "GET_PREDECESSOR":
			node.handleGetPredecessor(msg)
		case "GET_PREDECESSOR_REPLY":
			node.handleGetPredecessorReply(msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (node *Node) sendMessage(msg Message, addr *net.UDPAddr) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	_, err = node.Conn.WriteToUDP(data, addr)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// closestPrecedingNode finds the closest preceding node for the given key.
func (node *Node) closestPrecedingNode(key int) *NodeInfo {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for i := FTSize - 1; i >= 0; i-- {
		finger := node.FingerTable[i]
		if finger != nil && IsInInterval(finger.Node.ID, node.ID, key, false) {
			return finger.Node
		}
	}
	return NodeToNodeInfo(node)
}

func (node *Node) FixFingers() {
	var next int
	for {
		time.Sleep(500 * time.Millisecond)
		next = (next + 1) % FTSize
		if next == 0 {
			next = 1
		}
		start := (node.ID + int(math.Pow(2, float64(next-1)))) % RingSize
		node.FingerTable[next-1].Node = node.findSuccessor(start)
	}
}

func (node *Node) Stabilize() {
	for {
		time.Sleep(1 * time.Second)
		// Get successor's predecessor
		node.requestMutex.Lock()
		requestID := node.nextRequestID
		node.nextRequestID++
		responseChan := make(chan Message)
		node.pendingRequests[requestID] = responseChan
		node.requestMutex.Unlock()

		msg := Message{
			Type:      "GET_PREDECESSOR",
			SenderID:  node.ID,
			Address:   node.Address,
			RequestID: requestID,
		}
		node.sendMessage(msg, node.Successor.Address)

		// Wait for response
		select {
		case responseMsg := <-responseChan:
			node.requestMutex.Lock()
			delete(node.pendingRequests, requestID)
			node.requestMutex.Unlock()

			pred := &NodeInfo{
				ID:      responseMsg.SenderID,
				Address: responseMsg.Address,
			}
			if IsInInterval(pred.ID, node.ID, node.Successor.ID, false) {
				node.Successor = pred
			}
		case <-time.After(2 * time.Second):
			node.requestMutex.Lock()
			delete(node.pendingRequests, requestID)
			node.requestMutex.Unlock()
			log.Printf("Timeout waiting for GET_PREDECESSOR_REPLY")
		}

		notifyMsg := Message{
			Type:     "NOTIFY",
			SenderID: node.ID,
			Address:  node.Address,
		}
		node.sendMessage(notifyMsg, node.Successor.Address)
	}
}

func (node *Node) Join(existingNode *Node) {
	if existingNode != nil {
		node.Predecessor = nil
		node.Successor = existingNode.findSuccessor(node.ID)
	} else {
		node.Predecessor = NodeToNodeInfo(node)
		node.Successor = NodeToNodeInfo(node)
	}
	node.InitializeFingerTable()
}

func (node *Node) findSuccessor(id int) *NodeInfo {
	node.mutex.Lock()
	successor := node.Successor
	node.mutex.Unlock()

	if successor == nil {
		return NodeToNodeInfo(node)
	}

	if IsInInterval(id, node.ID, successor.ID, true) {
		return successor
	} else {
		next := node.closestPrecedingNode(id)
		if next.ID == node.ID {
			return successor
		}
		response := node.sendFindSuccessor(next, id)
		return response
	}
}

func (node *Node) sendFindSuccessor(target *NodeInfo, id int) *NodeInfo {
	node.requestMutex.Lock()
	requestID := node.nextRequestID
	node.nextRequestID++
	responseChan := make(chan Message)
	node.pendingRequests[requestID] = responseChan
	node.requestMutex.Unlock()

	msg := Message{
		Type:      "FIND_SUCCESSOR",
		Key:       id,
		SenderID:  node.ID,
		Address:   node.Address,
		RequestID: requestID,
	}

	node.sendMessage(msg, target.Address)

	// Wait for response
	select {
	case responseMsg := <-responseChan:
		node.requestMutex.Lock()
		delete(node.pendingRequests, requestID)
		node.requestMutex.Unlock()
		return &NodeInfo{
			ID:      responseMsg.SenderID,
			Address: responseMsg.Address,
		}
	case <-time.After(2 * time.Second):
		node.requestMutex.Lock()
		delete(node.pendingRequests, requestID)
		node.requestMutex.Unlock()
		log.Printf("Timeout waiting for FIND_SUCCESSOR_REPLY")
		return node.Successor
	}
}

func (node *Node) Store(key int, value string) {
	msg := Message{
		Type:     "STORE",
		Key:      key,
		Value:    value,
		SenderID: node.ID,
		Address:  node.Address,
	}
	node.handleStore(msg)
}

func (node *Node) Retrieve(key int) {
	msg := Message{
		Type:     "RETRIEVE",
		Key:      key,
		SenderID: node.ID,
		Address:  node.Address,
	}
	node.handleRetrieve(msg)
}
