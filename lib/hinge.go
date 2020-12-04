// functions to compute a hinge-tree decomposition of a hypergraph, and methods to use it to speed up the computation
// of a GHD

package lib

import (
	"bytes"
	"log"
	"reflect"
)

type hingeEdge struct {
	h Hingetree
	e Edge
}

type Algorithm_h interface {
	Name() string
	FindDecomp() Decomp
	FindDecompGraph(G Graph) Decomp
	SetWidth(K int)
}

type Hingetree struct {
	hinge    Graph
	decomp   Decomp
	minimal  bool
	children []hingeEdge
}

func (h Hingetree) GetLargestGraph() Graph {
	maxEdges := h.hinge.Edges.Len()
	maxGraph := h.hinge

	for _, c := range h.children {
		tmpGraph := c.h.GetLargestGraph()

		if tmpGraph.Edges.Len() > maxEdges {
			maxEdges = tmpGraph.Edges.Len()
			maxGraph = tmpGraph
		}
	}

	return maxGraph
}

func (h Hingetree) stringIdent(i int) string {
	var buffer bytes.Buffer

	if reflect.DeepEqual(h.decomp, Decomp{}) {
		buffer.WriteString("\n" + indent(i) + h.hinge.String() + "\n")
	} else {
		buffer.WriteString("\n" + indent(i) + h.decomp.String() + "\n")
	}

	if len(h.children) > 0 {
		buffer.WriteString(indent(i) + "Children:\n" + indent(i) + "[")
		for _, c := range h.children {
			buffer.WriteString("\n" + indent(i+1) + "sepEdge: " + c.e.String())
			buffer.WriteString(c.h.stringIdent(i + 1))
		}
		buffer.WriteString(indent(i) + "]\n")
	}

	return buffer.String()
}

func (h Hingetree) String() string {
	return h.stringIdent(0)
}

func GetHingeTree(g Graph) Hingetree {

	initialTree := Hingetree{hinge: g}

	isUsed := make(map[int]bool)

	for _, e := range g.Edges.Slice() {
		isUsed[e.Name] = false
	}

	//fmt.Println("Initial Hingetree \n")
	//fmt.Println(initialTree.String())

	output := initialTree.expandHingeTree(isUsed, -1)

	//fmt.Println("Proudced Hingetree \n")
	//fmt.Println(output.String())

	return output
}

// Following Gyssens et alii, 1994
func (h Hingetree) expandHingeTree(isUsed map[int]bool, parentE int) Hingetree {

	// keep expanding the current node until no more new children can be generated
	for !h.minimal {

		var e *Edge

		{ // maybe a separate function for this?
			//select next unused edge
			for _, i := range h.hinge.Edges.Slice() {
				if isUsed[i.Name] {
					continue
				} else {
					e = &i
					isUsed[i.Name] = true // set selected edge to used
					break
				}
			}
		}

		if e == nil {
			h.minimal = true

			//fmt.Println("Setting hinge ", h.hinge, " to minimal")
			continue
		}

		if parentE != -1 {
			//fmt.Println("Next unused edge", *e, "with parent Edge: ", m[parentE])

		} else {

			//fmt.Println("Next unused edge", *e, " at root")
		}

		sepEdge := NewEdges([]Edge{*e})
		hinges, gamma, _ := h.hinge.GetComponents(sepEdge)

		//fmt.Printf("Hinges of Sep: %v\n", hinges)

		// Skip reordering step if only single component
		if len(hinges) <= 1 {
			//fmt.Println("skipping sepEdge since no progress")
			continue
		}

		// hinges into Hingetrees, preserving order
		var htrees []Hingetree

		for i := range hinges {
			edges := hinges[i].Edges.Slice()
			extendedHinge := Graph{Edges: NewEdges(append(edges, *e))}
			htrees = append(htrees, Hingetree{hinge: extendedHinge})
		}

		// then assign the children to each Hingetree
		for _, hingedge := range h.children {
			i := gamma[hingedge.e.Name]
			htrees[i].children = append(htrees[i].children, hingedge)
		}

		// finally add Hingetrees above to newchildren
		if parentE == -1 {
			h = htrees[0]
		} else {
			h = htrees[gamma[parentE]]

			// found := false

			// for _, e := range h.hinge.Edges.Slice() {
			//  if e.Name == parentE {
			//      found = true
			//  }
			// }

			// if !found {
			//  log.Panic(m[parentE], "does not occur in", htrees[gamma[parentE]].hinge)
			// }

		}

		for i := range htrees {
			if (parentE == -1 && i == 0) || (parentE != -1 && i == gamma[parentE]) {
				continue
			}
			h.children = append(h.children, hingeEdge{e: *e, h: htrees[i]})
		}

		//fmt.Println("Current Hingetree \n")
		//fmt.Println(h.String())

	}

	//fmt.Println("going over children")

	// recursively repeat procedure over all children of h

	for i := range h.children {
		h.children[i].h = h.children[i].h.expandHingeTree(isUsed, h.children[i].e.Name)
	}

	return h
}

func (n Node) containsSubset(subset []int) bool {
	// every node contains itself
	if Subset(subset, n.Bag) {
		return true
	}
	// Check recursively if contained in children
	for _, child := range n.Children {
		if child.containsSubset(subset) {
			return true
		}
	}

	return false
}

func (n Node) parentSubset(subset []int) Node {
	// Check recursively if subset covered in children
	for i := range n.Children {
		if Subset(subset, n.Children[i].Bag) {
			return n
		} else if n.Children[i].containsSubset(subset) {
			return n.Children[i].parentSubset(subset)
		}

	}

	return n
}

// reroot G at node covering edge, producing an isomorphic graph
func (n Node) RerootEdge(edge []int) Node {

	if !n.containsSubset(edge) {
		log.Panicf("Can't reRoot: no node covering %+v in node %+v!\n", PrintVertices(edge), n)
	}
	if Subset(edge, n.Bag) {
		return n
	}
	p := n.parentSubset(edge)
	p = n.Reroot(p)
	var childIndex int

	// remove node containing edge from children of parent
	for i := range p.Children {
		if Subset(edge, p.Children[i].Bag) {
			childIndex = i
			break
		}
	}

	child := p.Children[childIndex]
	if childIndex == len(p.Children) {
		p.Children = p.Children[:childIndex] // slice out child
	} else {
		p.Children = append(p.Children[:childIndex], p.Children[childIndex+1:]...) // slice out child
	}

	child.Children = append(child.Children, p) // add p to children of c

	return child
}

func (h Hingetree) DecompHinge(alg Algorithm_h, g Graph) Decomp {

	h.decomp = alg.FindDecompGraph(h.hinge)

	if reflect.DeepEqual(h.decomp, Decomp{}) {
		return Decomp{}
	}

	// go recursively over children

	for i := range h.children {
		out := h.children[i].h.DecompHinge(alg, g)
		if reflect.DeepEqual(out, Decomp{}) { // reject if subtree cannot be merged to GHD
			return Decomp{}
		}
		//reroot child and parent to a connecting node:
		out.Root = out.Root.RerootEdge(h.children[i].e.Vertices)
		h.decomp.Root = h.decomp.Root.RerootEdge(h.children[i].e.Vertices)

		if Subset(out.Root.Bag, h.decomp.Root.Bag) {

			h.decomp.Root.Children = append(h.decomp.Root.Children, out.Root.Children...)
		} else {

			h.decomp.Root.Children = append(h.decomp.Root.Children, out.Root)
		}

	}

	h.decomp.Graph = g
	return h.decomp
}
