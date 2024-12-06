package main

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type NodeInfo struct {
	ID      int
	Address *net.UDPAddr
}

type Node struct {
	ID              int
	Address         *net.UDPAddr
	FingerTable     []*FingerEntry
	Successor       *NodeInfo
	Predecessor     *NodeInfo
	Data            map[int]string
	Conn            *net.UDPConn
	mutex           sync.Mutex
	requestMutex    sync.Mutex
	pendingRequests map[int]chan Message
	nextRequestID   int
}

func NewNode(id int, port string) *Node {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatalf("Falha ao resolver endereço UDP: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Falha ao escutar na porta UDP: %v", err)
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
			log.Printf("Erro lendo do UDP: %v", err)
			continue
		}
		var msg Message
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			log.Printf("Erro ao deserializar mensagem: %v", err)
			continue
		}
		switch msg.Type {
		case "STORE":
			node.handleStore(msg)
		case "RETRIEVE":
			node.handleRetrieve(msg, addr)
		case "FIND_SUCCESSOR":
			node.handleFindSuccessorMsg(msg)
		case "FIND_SUCCESSOR_REPLY":
			node.handleFindSuccessorReply(msg)
		case "NOTIFY":
			node.handleNotify(msg)
		case "GET_PREDECESSOR":
			node.handleGetPredecessor(msg)
		case "GET_PREDECESSOR_REPLY":
			node.handleGetPredecessorReply(msg)
		case "RETRIEVE_REPLY":
			node.handleRetrieveReply(msg)
		}
	}
}

func (node *Node) findSuccessor(id int) *NodeInfo {
	id = id % RingSize
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
		return node.sendFindSuccessor(next, id)
	}
}

func (node *Node) sendFindSuccessor(target *NodeInfo, id int) *NodeInfo {
	id = id % RingSize
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
		log.Printf("Timeout esperando FIND_SUCCESSOR_REPLY para chave %d no Node %d", id, node.ID)
		return node.Successor
	}
}

func (node *Node) Store(key int, value string) {
	key = key % RingSize
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
	key = key % RingSize
	msg := Message{
		Type:       "RETRIEVE",
		Key:        key,
		SenderID:   node.ID,
		Address:    node.Address,
		OriginID:   node.ID,
		OriginAddr: node.Address,
	}
	node.handleRetrieve(msg, node.Address)
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

func (node *Node) Stabilize() {
	for {
		time.Sleep(1 * time.Second)
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
			log.Printf("Timeout esperando GET_PREDECESSOR_REPLY no node %d", node.ID)
		}
		notifyMsg := Message{
			Type:     "NOTIFY",
			SenderID: node.ID,
			Address:  node.Address,
		}
		node.sendMessage(notifyMsg, node.Successor.Address)
	}
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
