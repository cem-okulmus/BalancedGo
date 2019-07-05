package main

import (
	"bytes"
	"log"
	"reflect"
)

// A Node is the root of a labelled tree, where the labels are the bag
// and the (edge) cover
type Node struct {
	bag      []int
	cover    []Edge
	children []Node
}

func (n Node) printBag() string {
	var buffer bytes.Buffer
	for i, v := range n.bag {
		buffer.WriteString(m[v])
		if i != len(n.bag)-1 {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
}

func indent(i int) string {
	output := ""

	for j := 0; j < i; j++ {
		output = output + "\t"
	}

	return output
}

func (n Node) stringIdent(i int) string {
	var buffer bytes.Buffer

	buffer.WriteString("\n" + indent(i) + "Bag: {")
	buffer.WriteString(n.printBag())

	buffer.WriteString("}\n" + indent(i) + "Cover: {")
	for i, e := range n.cover {
		buffer.WriteString(e.String())
		if i != len(n.cover)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}\n")
	if len(n.children) > 0 {
		buffer.WriteString(indent(i) + "Children:\n" + indent(i) + "[")
		for _, c := range n.children {
			buffer.WriteString(c.stringIdent(i + 1))
		}
		buffer.WriteString(indent(i) + "]\n")
	}

	return buffer.String()
}

func (n Node) String() string {
	return n.stringIdent(0)
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

func (n Node) bagSubsets() bool {
	if !subset(n.bag, Vertices(n.cover)) {
		return false
	}

	for _, c := range n.children {
		if !c.bagSubsets() {
			return false
		}
	}

	return true
}

// Think about how to make the contains check faster than linear
func (n Node) getConGraph(num int) Edges {
	var output Edges

	output.append(Edge{vertices: []int{num + encode + 1, num + encode + 1}}) // add loop (needed )

	for i, _ := range n.children {
		output.append(Edge{vertices: []int{num + encode + 1, (num + i + encode + 2)}}) //using breadth-first ordering to number nodes
	}

	for i, c := range n.children {
		output = append(output, c.getConGraph((num + 1 + i))...)
	}

	return output
}

func (n Node) allChildrenContaining(vert, num int) []int {
	var output []int
	//m[num+encode+1] = strconv.Itoa(num)

	if Contains(n.cover, vert) {
		output = append(output, num+encode+1)
	}

	for i, c := range n.children {
		output = append(output, c.allChildrenContaining(vert, (num+i+1))...)
	}

	return output
}

func (n Node) coversEdge(e Edge) bool {
	// edge contained in current node
	if subset(e.vertices, Vertices(n.cover)) {
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
	if !Contains(o.cover, i) {
		return o
	}
	if !(reflect.DeepEqual(o, n.parent(o))) && Contains(n.parent(o).cover, i) {
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

	return Node{bag: child.bag, cover: child.cover, children: newchildren}
}

// recurisvely collect all vertices from the bag of this node, and the bags of all its children
func (n Node) Vertices() []int {
	var output []int
	output = append(output, n.bag...)

	for _, c := range n.children {
		output = append(output, c.Vertices()...)
	}

	return output
}

//tests special condition violation on one node
func (n Node) specialCondition() bool {
	hiddenVertices := diff(Vertices(n.cover), n.bag)
	verticesRooted := n.Vertices()

	for _, v := range hiddenVertices {
		if mem(verticesRooted, v) {
			log.Println("Vertex ", v, " violates special condition")
			return false
		}
	}

	return true
}

//test special condition recursively on entire subtree rooted at node
func (n Node) noSCViolation() bool {
	if !n.specialCondition() {
		return false
	}

	for _, c := range n.children {
		if !c.specialCondition() {
			return false
		}
	}

	return true
}
