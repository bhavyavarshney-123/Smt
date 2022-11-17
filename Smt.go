package smt

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"hash"
	"log"
)

//constant Defining the right side of the tree as 1.
const (
	right = 1
)

//DefaultVal is the empty slice of Byte
var DefaultVal []byte

//SparseMerkleTree is the struct defining sparse Merkle Tree.
type SparseMerkleTree struct {
	st            SmtHasher
	values, nodes MapDb
	root          *SparseMerkleNode
}

type SparseMerkleNode struct {
	Left  *SparseMerkleNode
	Right *SparseMerkleNode
	data  []byte
}

//NewSparseMerkleNode for creating a new merkle Node
func NewSparseMerkleNode(left, right *SparseMerkleNode, data []byte) *SparseMerkleNode {
	node := &SparseMerkleNode{}
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

//Option is a function that configures Smt.
type Option func(tree *SparseMerkleTree)

//NewSparseMerkleTree creates a new Sparse Merkle on an empty MapDb.
func NewSparseMerkleTree(nodes, values MapDb, hasher hash.Hash, opts ...Option) *SparseMerkleTree {
	smt := SparseMerkleTree{
		st:     *newSmtHasher(hasher),
		nodes:  nodes,
		values: values,
	}

	for _, opt := range opts {
		opt(&smt)
	}

	smt.SetRoot(nil)

	return &smt
}

// GetRoot gets the root of the tree.
func (smt *SparseMerkleTree) GetRoot() *SparseMerkleNode {
	return smt.root
}

// SetRoot sets the root of the tree.
func (smt *SparseMerkleTree) SetRoot(root *SparseMerkleNode) *SparseMerkleTree {
	if smt.root == nil {
		smt.root = root
	} else {
		fmt.Println("root already exists")
	}

	return smt
}

//cache function insert that insert the data into sparse merkle Tree
func insert(left, right *SparseMerkleNode, data []byte) *SparseMerkleNode {

	node := &SparseMerkleNode{}
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

func NewMerkleTree(data [][]byte) *SparseMerkleTree {
	var nodes []SparseMerkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, dat := range data {
		node := insert(nil, nil, dat)
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

	var level []SparseMerkleNode

	for j := 0; j < len(nodes); j += 2 {
		node := insert(&nodes[j], &nodes[j+1], nil)
		level = append(level, *node)
	}

	nodes = level

	tree := SparseMerkleTree{root: &nodes[0]}

	return &tree
}

//Depth for the dept of the Sparse Merkle Tree
func (smt *SparseMerkleTree) depth() int {
	return smt.st.pathSize() * 8
}

// Get gets the value of a key from the tree.
func (smt *SparseMerkleTree) Get(key []byte) ([]byte, error) {
	// Get tree's root
	root := smt.GetRoot()

	if root == nil {
		// The tree is empty, return the default value.
		return DefaultVal, nil
	}

	path := smt.st.path(key)
	value, err := smt.values.Get(path)

	if err != nil {
		var invalidKeyError *InvalidKey

		if errors.As(err, &invalidKeyError) {
			// If key isn't found, return default value
			return DefaultVal, nil
		} else {
			// Otherwise percolate up any other error
			return nil, err
		}
	}
	return value, nil
}

//Check is a Bool function that returns true if the value at given key is non-default and false otherwise.
func (smt *SparseMerkleTree) Check(key []byte) (bool, error) {
	val, err := smt.Get(key)
	return !bytes.Equal(DefaultVal, val), err
}

// Update sets a new value for a key in the tree, and sets and returns the new root of the tree.
func (smt *SparseMerkleTree) Update(key []byte, value []byte) ([]byte, error) {
	getroot := GobEncode(*smt.GetRoot())
	newRoot, err := smt.RootUpdate(key, value, getroot)
	if err != nil {
		return nil, err
	}
	var root *SparseMerkleNode
	setroot := GobDecode(newRoot, root)
	smt.SetRoot(setroot)
	return newRoot, nil
}

// Delete deletes a value from tree. It returns the new default root of the tree.
func (smt *SparseMerkleTree) Delete(key []byte) ([]byte, error) {
	return smt.Update(key, DefaultVal)
}

//RootUpdate set and return new value for the key in the tree at a specific root.
func (smt *SparseMerkleTree) RootUpdate(key, value, root []byte) ([]byte, error) {
	path := smt.st.path(key)
	sideNodes, pathNodes, OldLeafValue, _, err := smt.sideNodesForRoot(path, root, false)
	if err != nil {
		return nil, err
	}

	var NewRoot []byte
	if bytes.Equal(value, DefaultVal) {
		// Delete operation.
		NewRoot, err = smt.DeleteNode(path, OldLeafValue, pathNodes, sideNodes)
		if err != nil {
			// This key is already empty; return the old root.
			return root, nil
		}
		if err := smt.values.Delete(path); err != nil {
			return nil, err
		}

	} else {
		// Insert or update operation.
		NewRoot, err = smt.UpdateNodes(path, value, OldLeafValue, pathNodes, sideNodes)
	}
	return NewRoot, err
}

// DeleteRoot deletes a value from tree at a specific root. It returns the new default root of the tree.
func (smt *SparseMerkleTree) DeleteRoot(key, root []byte) ([]byte, error) {
	return smt.RootUpdate(key, DefaultVal, root)
}

//DeleteNode deletes a value from the tree at a specific Node.It returns the new Node.
func (smt *SparseMerkleTree) DeleteNode(path, OldLeafValue []byte, sideNodes, pathNodes [][]byte) ([]byte, error) {
	if bytes.Equal(pathNodes[0], smt.st.EmptyPlace()) {

		//This key is already empty so return an error.
		return nil, fmt.Errorf("key is already empty")
	}

	ActualPath, _ := smt.st.parseLeaf(OldLeafValue)
	if !bytes.Equal(path, ActualPath) {

		//Both the keys are not similar then the different key was found at its place therefore return an error.
		return nil, fmt.Errorf("key not found")

	}

	//All nodes above the deleted node are now orphaned
	for _, Node := range pathNodes {
		if err := smt.nodes.Delete(Node); err != nil {
			return nil, err
		}
	}

	var CurrentNodeHash, CurrentNodeData []byte
	nonZeroValueReached := false
	for i, sideNode := range sideNodes {
		if CurrentNodeData == nil {
			sideNodeValue, err := smt.nodes.Get(sideNode)
			if err != nil {
				return nil, err
			}

			if smt.st.isLeaf(sideNodeValue) {
				//This is the leaf sibling that needs to be bubbled up the tree.
				CurrentNodeHash = sideNode
				CurrentNodeData = sideNode
				continue
			} else {
				//This is the node sibling that needs to be left in its place.
				CurrentNodeHash = smt.st.EmptyPlace()
				nonZeroValueReached = true

			}
		}
		if !nonZeroValueReached && bytes.Equal(sideNode, smt.st.zeroValue) {
			// We found another placeholder sibling node, keep going up the
			// tree until we find the first sibling that is not a placeholder.
			continue
		} else if !nonZeroValueReached {
			// We found the first sibling node that is not a placeholder, it is
			// time to insert our leaf sibling node here.
			nonZeroValueReached = true
		}

		if getBitFromMSB(path, len(sideNodes)-1-i) == right {
			CurrentNodeHash, CurrentNodeData = smt.st.digestNode(sideNode, CurrentNodeData)
		} else {
			CurrentNodeHash, CurrentNodeData = smt.st.digestNode(CurrentNodeData, sideNode)
		}
		if err := smt.nodes.Set(CurrentNodeHash, CurrentNodeData); err != nil {
			return nil, err
		}
		CurrentNodeData = CurrentNodeHash
	}

	if CurrentNodeHash == nil {
		// The tree is empty; return placeholder value as root.
		CurrentNodeHash = smt.st.EmptyPlace()
	}
	return CurrentNodeHash, nil
}

//UpdateNodes updates a value from the tree at a specific Node.It returns the new Node.
func (smt *SparseMerkleTree) UpdateNodes(path, OldLeafValue []byte, value []byte, sideNodes, pathNodes [][]byte) ([]byte, error) {
	valueHash := smt.st.digest(value)
	currentNodeHash, currentNodeData := smt.st.digestLeaf(path, valueHash)
	if err := smt.nodes.Set(currentNodeHash, currentNodeData); err != nil {
		return nil, err
	}
	currentNodeData = currentNodeHash

	// If the leaf node that sibling nodes lead to has a different actual path
	// than the leaf node being updated, we need to create an intermediate node
	// with this leaf node and the new leaf node as children.

	// First, get the number of bits that the paths of the two leaf nodes share in common as a prefix.
	var commonPrefixCount int
	var oldValueHash []byte
	if bytes.Equal(pathNodes[0], smt.st.EmptyPlace()) {
		commonPrefixCount = smt.depth()
	} else {
		var actualPath []byte
		actualPath, oldValueHash = smt.st.parseLeaf(OldLeafValue)
		commonPrefixCount = countCommonPrefix(path, actualPath)
	}
	if commonPrefixCount != smt.depth() {
		if getBitFromMSB(path, commonPrefixCount) == right {
			currentNodeHash, currentNodeData = smt.st.digestNode(pathNodes[0], currentNodeData)
		} else {
			currentNodeHash, currentNodeData = smt.st.digestNode(currentNodeData, pathNodes[0])
		}

		err := smt.nodes.Set(currentNodeHash, currentNodeData)
		if err != nil {
			return nil, err
		}

		currentNodeData = currentNodeHash
	} else if oldValueHash != nil {
		// Short-circuit if the same value is being set
		if bytes.Equal(oldValueHash, valueHash) {
			return GobEncode(*smt.root), nil
		}
		// If an old leaf exists, remove it
		if err := smt.nodes.Delete(pathNodes[0]); err != nil {
			return nil, err
		}
		if err := smt.values.Delete(path); err != nil {
			return nil, err
		}
	}
	// All remaining path nodes are orphaned
	for i := 1; i < len(pathNodes); i++ {
		if err := smt.nodes.Delete(pathNodes[i]); err != nil {
			return nil, err
		}
	}

	// The offset from the bottom of the tree to the start of the side nodes.
	// Note: i-offsetOfSideNodes is the index into sideNodes[]
	offsetOfSideNodes := smt.depth() - len(sideNodes)

	for i := 0; i < smt.depth(); i++ {
		var sideNode []byte

		if i-offsetOfSideNodes < 0 || sideNodes[i-offsetOfSideNodes] == nil {
			if commonPrefixCount != smt.depth() && commonPrefixCount > smt.depth()-1-i {
				// If there are no sideNodes at this height, but the number of
				// bits that the paths of the two leaf nodes share in common is
				// greater than this depth, then we need to build up the tree
				// to this depth with placeholder values at siblings.
				sideNode = smt.st.EmptyPlace()
			} else {
				continue
			}
		} else {
			sideNode = sideNodes[i-offsetOfSideNodes]
		}

		if getBitFromMSB(path, smt.depth()-1-i) == right {
			currentNodeHash, currentNodeData = smt.st.digestNode(sideNode, currentNodeData)
		} else {
			currentNodeHash, currentNodeData = smt.st.digestNode(currentNodeData, sideNode)
		}
		err := smt.nodes.Set(currentNodeHash, currentNodeData)
		if err != nil {
			return nil, err
		}
		currentNodeData = currentNodeHash
	}
	if err := smt.values.Set(path, value); err != nil {
		return nil, err
	}

	return currentNodeHash, nil
}

// Get all the sibling nodes (sideNodes) for a given path from a given root.
// Returns an array of sibling nodes, the leaf hash found at that path, the leaf data, and the sibling data.
// If the leaf is a EmptyPlace, the leaf data is nil.
func (smt *SparseMerkleTree) sideNodesForRoot(path []byte, root []byte, getSiblingData bool) ([][]byte, [][]byte, []byte, []byte, error) {
	// Side nodes for the path. Nodes are inserted in reverse order, then the slice is reversed at the end.
	sideNodes := make([][]byte, 0, smt.depth())
	pathNodes := make([][]byte, 0, smt.depth()+1)
	pathNodes = append(pathNodes, root)

	if bytes.Equal(root, smt.st.EmptyPlace()) {
		// If the root is a EmptyPlace, there are no sideNodes to return.
		// Let the "actual path" be the input path.
		return sideNodes, pathNodes, nil, nil, nil
	}

	currentData, err := smt.nodes.Get(root)
	if err != nil {
		return nil, nil, nil, nil, err
	} else if smt.st.isLeaf(currentData) {
		// If the root is a leaf, there are also no sideNodes to return.
		return sideNodes, pathNodes, currentData, nil, nil
	}

	var nodeHash []byte
	var sideNode []byte
	var siblingData []byte
	for i := 0; i < smt.depth(); i++ {
		leftNode, rightNode := smt.st.parseNode(currentData)

		// Get sideNode depending on whether the path bit is on or off.
		if getBitFromMSB(path, i) == right {
			sideNode = leftNode
			nodeHash = rightNode
		} else {
			sideNode = rightNode
			nodeHash = leftNode
		}
		sideNodes = append(sideNodes, sideNode)
		pathNodes = append(pathNodes, nodeHash)

		if bytes.Equal(nodeHash, smt.st.EmptyPlace()) {
			// If the node is a EmptyPlace, we've reached the end.
			currentData = nil
			break
		}

		currentData, err = smt.nodes.Get(nodeHash)
		if err != nil {
			return nil, nil, nil, nil, err
		} else if smt.st.isLeaf(currentData) {
			// If the node is a leaf, we've reached the end.
			break
		}
	}

	if getSiblingData {
		siblingData, err = smt.nodes.Get(sideNode)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}
	return reverseByteSlices(sideNodes), reverseByteSlices(pathNodes), currentData, siblingData, nil
}
