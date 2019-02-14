package main

import (
	"bytes"
	"log"
	"reflect"
)

type Node struct {
	lambda   []Edge
	children []Node
}

func indent(i int) string {
	output := ""

	for j := 0; j < i; j++ {
		output = output + "\t"
	}

	return output
}
func (n Node) StringIdent(i int) string {
	var buffer bytes.Buffer

	buffer.WriteString("\n" + indent(i) + "Lambda: {")
	for i, e := range n.lambda {
		buffer.WriteString(e.String())
		if i != len(n.lambda)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}\n")
	if len(n.children) > 0 {
		buffer.WriteString(indent(i) + "Children:\n" + indent(i) + "[")
		for _, c := range n.children {
			buffer.WriteString(c.StringIdent(i + 1))
		}
		buffer.WriteString(indent(i) + "]\n")
	}

	return buffer.String()
}

func (n Node) String() string {
	return n.StringIdent(0)
}

func (n Node) contains(o Node) bool {

	// every node contains itself
	if reflect.DeepEqual(n, o) {
		return true
	}

	// Check recursively if contained in children
	for _, child := range n.children {
		if child.contains(o) {
			return true
		}
	}

	return false
}

func (n Node) connected(i int) bool {
	var containingNodes = n.allChildrenContaining(i)

	for _, k := range containingNodes {
		for _, l := range containingNodes {
			if !reflect.DeepEqual(n.ancestorOnI(k, i), n.ancestorOnI(l, i)) {
				return false
			}

		}
	}

	return true
}

func (n Node) allChildrenContaining(i int) []Node {
	var output []Node

	if Contains(n.lambda, i) {
		output = append(output, n)
	}

	for _, c := range n.children {
		output = append(output, c.allChildrenContaining(i)...)
	}

	return output
}

func (n Node) coversEdge(e Edge) bool {
	// edge contained in current node
	if subset(e.nodes, Vertices(n.lambda)) {
		return true
	}

	// Check recursively if contained in children
	for _, child := range n.children {
		if child.coversEdge(e) {
			return true
		}
	}

	return false
}

func (n Node) ancestorOnI(o Node, i int) Node {
	if !Contains(o.lambda, i) {
		return o
	}
	if !(reflect.DeepEqual(o, n.parent(o))) && Contains(n.parent(o).lambda, i) {
		return n.ancestorOnI(n.parent(o), i)
	}

	return o
}

func (n Node) parent(o Node) Node {
	// Check recursively if contained in children
	for _, child := range n.children {
		if reflect.DeepEqual(child, o) {
			return n
		} else if child.contains(o) {
			return child.parent(o)
		}

	}

	return o
}

// reroot G at child, producing an isomorphic graph
func (n Node) reroot(child Node) Node {

	if !n.contains(child) {
		log.Panicf("Can't reroot: no child %+v in node %+v!\n", child, n)
	}
	if reflect.DeepEqual(n, child) {
		return child
	}
	p := n.parent(child)
	p = n.reroot(p)

	// remove child from children of parent
	var newparentchildren []Node
	for _, c := range p.children {
		if reflect.DeepEqual(c, child) {
			continue
		}
		newparentchildren = append(newparentchildren, c)
	}
	p.children = newparentchildren
	newchildren := append(child.children, p)

	return Node{lambda: child.lambda, children: newchildren}
}
