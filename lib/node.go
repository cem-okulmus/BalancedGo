package lib

import (
	"bytes"
	"log"
	"reflect"
	"strconv"
)

// A Node is the root of a labelled tree, where the labels are the bag
// and the (edge) cover
type Node struct {
	num           int
	Up            []int
	Low           []int
	Bag           []int
	Cover         Edges
	Children      []Node
	LowConnecting bool //used by Divide to determine reorder points
	Star          bool // used to indicate nodes which need to be updated
}

func (n Node) printUp() string {
	mutex.RLock()
	defer mutex.RUnlock()
	var buffer bytes.Buffer
	for i, v := range n.Up {
		buffer.WriteString(m[v])
		if i != len(n.Up)-1 {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
}
func (n Node) printLow() string {
	mutex.RLock()
	defer mutex.RUnlock()
	var buffer bytes.Buffer
	for i, v := range n.Low {
		buffer.WriteString(m[v])
		if i != len(n.Low)-1 {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
}

func (n Node) printBag() string {
	mutex.RLock()
	defer mutex.RUnlock()
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

	if n.LowConnecting {
		buffer.WriteString("\n" + indent(i) + "LowConn: {" + strconv.FormatBool(n.LowConnecting) + "} Bag: {" + n.printBag())
	} else {
		buffer.WriteString("\n" + indent(i) + "Bag: {" + n.printBag() + "}")
	}
	if len(n.Up) > 0 || len(n.Low) > 0 {
		buffer.WriteString("Up:{ " + n.printUp() + "} Low:{" + n.printLow() + "}")
	}
	if n.Star {
		buffer.WriteString(" 🧙")
	}
	buffer.WriteString("\n" + indent(i) + "Cover: {")
	for i, e := range n.Cover.Slice() {
		buffer.WriteString(e.String())
		if i != n.Cover.Len()-1 {
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
	if !Subset(n.Bag, n.Cover.Vertices()) {
		return false
	}

	for _, c := range n.Children {
		if !c.bagSubsets() {
			return false
		}
	}

	return true
}

func (n *Node) getNumber() {
	if n.num == 0 {

		temp := "num : " + strconv.Itoa(n.num) + " node: " + n.Cover.String()
		mutex.Lock()
		n.num = encode
		m[n.num] = temp
		encode++
		mutex.Unlock()

	}
}

// Think about how to make the contains check faster than linear
func (n *Node) getConGraph(withLoops bool) Edges {
	var output []Edge

	n.getNumber()
	if withLoops { // loops needed for connectivty check
		output = append(output, Edge{Vertices: []int{n.num, n.num}})
	}

	for i, _ := range n.Children {
		n.Children[i].getNumber()
		output = append(output, Edge{Vertices: []int{n.num, n.Children[i].num}}) //using breadth-first ordering to number nodes
	}

	for _, c := range n.Children {
		edgesChild := c.getConGraph(withLoops)
		output = append(output, edgesChild.Slice()...)
	}

	return NewEdges(output)
}

func (n *Node) allChildrenContaining(vert int) []int {
	var output []int

	if Mem(n.Bag, vert) {
		output = append(output, n.num)
	}

	for _, c := range n.Children {
		output = append(output, c.allChildrenContaining(vert)...)
	}

	return output
}

func (n Node) coversEdge(e Edge) bool {
	// edge contained in current node
	if Subset(e.Vertices, n.Bag) {
		return true
	}

	// Check recursively if contained in children
	for i := range n.Children {
		if n.Children[i].coversEdge(e) {
			return true
		}
	}

	return false
}

func (n Node) parent(o Node) Node {
	// Check recursively if contained in children
	for i := range n.Children {
		if reflect.DeepEqual(n.Children[i], o) {
			return n
		} else if n.Children[i].contains(o) {
			return n.Children[i].parent(o)
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
	hiddenVertices := Diff(n.Cover.Vertices(), n.Bag)
	verticesRooted := n.Vertices()

	for _, v := range hiddenVertices {
		if Mem(verticesRooted, v) {
			mutex.RLock()
			log.Println("Vertex ", m[v], " violates special condition")
			mutex.RUnlock()
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

	for i := range n.Children {
		if !n.Children[i].noSCViolation() {
			return false
		}
	}

	return true
}

func (n *Node) RestoreEdges(edges Edges) Node {
	var nuCover []Edge

OUTER:
	for _, e2 := range n.Cover.Slice() {
		if e2.Name != 0 {
			nuCover = append(nuCover, e2)
			continue
		}
		for _, e := range edges.Slice() {
			if Subset(e2.Vertices, e.Vertices) && e.Name != 0 {
				nuCover = append(nuCover, e)
				continue OUTER
			}
		}
	}

	var nuChildern []Node

	for i := range n.Children {
		nuChildern = append(nuChildern, n.Children[i].RestoreEdges(edges))
	}

	return Node{Bag: n.Bag, Cover: NewEdges(nuCover), Children: nuChildern}
}

func (n *Node) CombineNodes(subtree Node) *Node {

	if n.LowConnecting {
		n.LowConnecting = false
		n.Children = append(n.Children, subtree)
		return n
	}

	for i := range n.Children {
		result := n.Children[i].CombineNodes(subtree)
		if !reflect.DeepEqual(*result, Node{}) {
			n.Children[i] = *result
			return n
		}
	}

	return &Node{}
}
