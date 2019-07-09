package lib

import (
	"bytes"
	"log"
	"reflect"
)

// A Node is the root of a labelled tree, where the labels are the bag
// and the (edge) cover
type Node struct {
	Bag      []int
	Cover    []Edge
	Children []Node
}

func (n Node) printBag() string {
	var buffer bytes.Buffer
	for i, v := range n.Bag {
		buffer.WriteString(m[v])
		if i != len(n.Bag)-1 {
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
	for i, e := range n.Cover {
		buffer.WriteString(e.String())
		if i != len(n.Cover)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}\n")
	if len(n.Children) > 0 {
		buffer.WriteString(indent(i) + "Children:\n" + indent(i) + "[")
		for _, c := range n.Children {
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
	for _, child := range n.Children {
		if child.contains(o) {
			return true
		}
	}

	return false
}

func (n Node) bagSubsets() bool {
	if !Subset(n.Bag, Vertices(n.Cover)) {
		return false
	}

	for _, c := range n.Children {
		if !c.bagSubsets() {
			return false
		}
	}

	return true
}

// Think about how to make the contains check faster than linear
func (n Node) getConGraph(num int) Edges {
	var output Edges

	output.append(Edge{Vertices: []int{num + encode + 1, num + encode + 1}}) // add loop (needed )

	for i, _ := range n.Children {
		output.append(Edge{Vertices: []int{num + encode + 1, (num + i + encode + 2)}}) //using breadth-first ordering to number nodes
	}

	for i, c := range n.Children {
		output = append(output, c.getConGraph((num + 1 + i))...)
	}

	return output
}

func (n Node) allChildrenContaining(vert, num int) []int {
	var output []int
	//m[num+encode+1] = strconv.Itoa(num)

	if Contains(n.Cover, vert) {
		output = append(output, num+encode+1)
	}

	for i, c := range n.Children {
		output = append(output, c.allChildrenContaining(vert, (num+i+1))...)
	}

	return output
}

func (n Node) coversEdge(e Edge) bool {
	// edge contained in current node
	if Subset(e.Vertices, Vertices(n.Cover)) {
		return true
	}

	// Check recursively if contained in children
	for _, child := range n.Children {
		if child.coversEdge(e) {
			return true
		}
	}

	return false
}

func (n Node) ancestorOnI(o Node, i int) Node {
	if !Contains(o.Cover, i) {
		return o
	}
	if !(reflect.DeepEqual(o, n.parent(o))) && Contains(n.parent(o).Cover, i) {
		return n.ancestorOnI(n.parent(o), i)
	}

	return o
}

func (n Node) parent(o Node) Node {
	// Check recursively if contained in children
	for _, child := range n.Children {
		if reflect.DeepEqual(child, o) {
			return n
		} else if child.contains(o) {
			return child.parent(o)
		}

	}

	return o
}

// reroot G at child, producing an isomorphic graph
func (n Node) Reroot(child Node) Node {

	if !n.contains(child) {
		log.Panicf("Can't reRoot: no child %+v in node %+v!\n", child, n)
	}
	if reflect.DeepEqual(n, child) {
		return child
	}
	p := n.parent(child)
	p = n.Reroot(p)

	// remove child from children of parent
	var newparentchildren []Node
	for _, c := range p.Children {
		if reflect.DeepEqual(c, child) {
			continue
		}
		newparentchildren = append(newparentchildren, c)
	}
	p.Children = newparentchildren
	newchildren := append(child.Children, p)

	return Node{Bag: child.Bag, Cover: child.Cover, Children: newchildren}
}

// recurisvely collect all vertices from the bag of this node, and the bags of all its children
func (n Node) Vertices() []int {
	var output []int
	output = append(output, n.Bag...)

	for _, c := range n.Children {
		output = append(output, c.Vertices()...)
	}

	return output
}

//tests special condition violation on one node
func (n Node) specialCondition() bool {
	hiddenVertices := Diff(Vertices(n.Cover), n.Bag)
	verticesRooted := n.Vertices()

	for _, v := range hiddenVertices {
		if Mem(verticesRooted, v) {
			log.Println("Vertex ", m[v], " violates special condition")
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

	for _, c := range n.Children {
		if !c.specialCondition() {
			return false
		}
	}

	return true
}

func (n *Node) RestoreEdges(edges Edges) Node {
	var nuCover []Edge

	for _, e := range edges {
		for _, e2 := range n.Cover {
			if Subset(e2.Vertices, e.Vertices) {
				nuCover = append(nuCover, e)
			}
		}
	}

	var nuChildern []Node

	for _, c := range n.Children {

		nuChildern = append(nuChildern, c.RestoreEdges(edges))
	}

	return Node{Bag: n.Bag, Cover: nuCover, Children: nuChildern}
}