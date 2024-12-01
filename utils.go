package main

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
