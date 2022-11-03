package Learning

import "fmt"

//BinaryTree struct
type BinaryTree struct {
	root *BinaryNode
}

//BinaryNode struct defining the nodes of the binary Tree.
type BinaryNode struct {
	left  *BinaryNode
	right *BinaryNode
	data  byte
}

//insert function to insert root in the binary tree
func (t *BinaryTree) insert(data byte) *BinaryTree {
	if t.root == nil {
		t.root = &BinaryNode{data: data, right: nil, left: nil}
	} else {
		t.root.insert(data)
	}

	return t
}

//insert function to insert the left or right node in the binary tree.
func (i *BinaryNode) insert(data byte) {
	if i == nil {
		return
		//if data is less than the root data then insert data to left node.
	} else if data <= i.data {
		if i.left == nil {
			i.left = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			i.left.insert(data)
		}
		//otherwise to right node.
	} else {
		if i.right == nil {
			i.right = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			i.right.insert(data)
		}
	}
}

//printPreOrder print the binary tree in the Pre Order.
func printPreOrder(n *BinaryNode) {
	if n == nil {
		return
	} else {
		fmt.Println(n.data)
		printPreOrder(n.left)
		printPreOrder(n.right)
	}
}

//printInOrder print the binary tree in the In Order.
func printInOrder(n *BinaryNode) {
	if n == nil {
		return
	} else {
		printInOrder(n.left)
		fmt.Println(n.data)
		printInOrder(n.right)

	}
}

//printPostOrder print the binary tree in Post Order.
func printPostOrder(n *BinaryNode) {
	if n == nil {
		return
	} else {
		printPostOrder(n.left)
		printPostOrder(n.right)
		fmt.Println(n.data)
	}
}

func main() {
	tree := &BinaryTree{}
	tree.insert(10)
	tree.insert(20)
	tree.insert(30)
	tree.insert(40)
	fmt.Printf("Pre Order")
	printPreOrder(tree.root)
	fmt.Printf("In Order")
	printInOrder(tree.root)
	fmt.Printf("Post Order")
	printPostOrder(tree.root)

}
