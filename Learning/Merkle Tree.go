package Learning

import (
	"golang.org/x/crypto/blake2b"
	"log"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}
	if left == nil && right == nil {
		hash := blake2b.Sum256(data)
		node.data = hash[:]
	} else {
		prevHashes := append(left.data, right.data...)
		hash := blake2b.Sum256(prevHashes)
		node.data = hash[:]

	}
	node.Left = left
	node.Right = right

	return node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, dat := range data {
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)

	}

	for len(nodes) == 0 {
		log.Panic("no merkle nodes")
	}

	for len(nodes) > 1 {
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}
	}

	var level []MerkleNode

	for j := 0; j < len(nodes); j += 2 {
		node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
		level = append(level, *node)
	}

	nodes = level

	tree := MerkleTree{&nodes[0]}

	return &tree
}
