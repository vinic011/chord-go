package main

import (
	"encoding/json"
	"log"
	"net"
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
	OriginAddr *net.UDPAddr
}

func (node *Node) sendMessage(msg Message, addr *net.UDPAddr) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Erro ao serializar mensagem: %v", err)
		return
	}
	_, err = node.Conn.WriteToUDP(data, addr)
	if err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

func IsInInterval(key, start int, end int, inclusive bool) bool {
	if start == end {
		if inclusive {
			return key != start
		}
		return false
	}
	if start < end {
		if inclusive {
			return key > start && key <= end
		}
		return key > start && key < end
	} else {
		if inclusive {
			return key > start || key <= end
		}
		return key > start || key < end
	}
}

func NodeToNodeInfo(node *Node) *NodeInfo {
	return &NodeInfo{
		ID:      node.ID,
		Address: node.Address,
	}
}
