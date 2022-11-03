package smt

import (
	"bytes"
	"golang.org/x/crypto/blake2b"
	"hash"
)

var leafPrefix = []byte{0}
var nodePrefix = []byte{1}

//SmtHasher struct used to hash the Sparse Merkle Tree
type SmtHasher struct {
	hasher    hash.Hash
	zeroValue []byte
}

//newSmtHasher for making a new Sparse Merkle Tree Hasher
func newSmtHasher(hasher hash.Hash) *SmtHasher {
	st := SmtHasher{hasher: hasher}
	st.zeroValue = make([]byte, st.pathSize())
	return &st
}

func (st *SmtHasher) digest(data []byte) []byte {
	st.hasher.Write(data)
	s := st.hasher.Sum(nil)
	sum := blake2b.Sum256(s)
	Sum := sum[:]
	st.hasher.Reset()
	return Sum
}

func (st *SmtHasher) path(key []byte) []byte {
	return st.digest(key)
}

func (st *SmtHasher) digestLeaf(path []byte, leafData []byte) ([]byte, []byte) {
	value := make([]byte, 0, len(leafPrefix)+len(path)+len(leafData))
	value = append(value, leafPrefix...)
	value = append(value, path...)
	value = append(value, leafData...)

	st.hasher.Write(value)
	s := st.hasher.Sum(nil)
	sum := blake2b.Sum256(s)
	Sum := sum[:]
	st.hasher.Reset()

	return Sum, value
}

func (st *SmtHasher) parseLeaf(data []byte) ([]byte, []byte) {
	return data[len(leafPrefix) : st.pathSize()+len(leafPrefix)], data[len(leafPrefix)+st.pathSize():]
}

func (st *SmtHasher) isLeaf(data []byte) bool {
	return bytes.Equal(data[:len(leafPrefix)], leafPrefix)
}

func (st *SmtHasher) digestNode(leftData []byte, rightData []byte) ([]byte, []byte) {
	value := make([]byte, 0, len(nodePrefix)+len(leftData)+len(rightData))
	value = append(value, nodePrefix...)
	value = append(value, leftData...)
	value = append(value, rightData...)

	st.hasher.Write(value)
	s := st.hasher.Sum(nil)
	sum := blake2b.Sum256(s)
	Sum := sum[:]
	st.hasher.Reset()

	return Sum, value
}

func (st *SmtHasher) parseNode(data []byte) ([]byte, []byte) {
	return data[len(nodePrefix) : st.pathSize()+len(nodePrefix)], data[len(nodePrefix)+st.pathSize():]
}

func (st *SmtHasher) pathSize() int {
	return st.hasher.Size()
}

func (st *SmtHasher) EmptyPlace() []byte {
	return st.zeroValue
}
