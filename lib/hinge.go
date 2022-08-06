package lib

// hinge.go provides functions to compute a hinge-tree decomposition of a hypergraph,
// and methods to use it to speed up the computation of a GHD

import (
	"bytes"
	"log"
	"reflect"

	"github.com/cem-okulmus/disjoint"
)

type hingeEdge struct {
	h Hingetree
	e Edge
}

// AlgorithmH is strict generalisation on the Algorithm interface.
type AlgorithmH interface {
	Name() string
	FindDecomp() Decomp
	FindDecompGraph(G Graph) Decomp
	SetWidth(K int)
}

// A Hingetree is a tree with each node representing a subgraph
// and its respective decomposition and connected by exactly one edge,
// which is the sole intersection between the two graphs of its connecting nodes
type Hingetree struct {
	hinge    Graph
	decomp   Decomp
	minimal  bool
	children []hingeEdge
}

// GetLargestGraph returns the largest graph within the hinge tree
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

// GetHingeTree produces a hingetree for a given graph
func GetHingeTree(g Graph) Hingetree {
	initialTree := Hingetree{hinge: g}
	isUsed := make(map[int]bool)

	for _, e := range g.Edges.Slice() {
		isUsed[e.Name] = false
	}

	return initialTree.expandHingeTree(isUsed, -1)
}

// expandHingeTree computes a hintegree, following an algorithm from Gyssens et alii, 1994
func (h Hingetree) expandHingeTree(isUsed map[int]bool, parentE int) Hingetree {

	// keep expanding the current node until no more new children can be generated
	for !h.minimal {
		var e *Edge

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

		if e == nil {
			h.minimal = true
			continue
		}

		var Vertices = make(map[int]*disjoint.Element)
		sepEdge := NewEdges([]Edge{*e})
		hinges, gamma, _ := h.hinge.GetComponents(sepEdge, Vertices)

		// Skip reordering step if only single component
		if len(hinges) <= 1 {
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
		}

		for i := range htrees {
			if (parentE == -1 && i == 0) || (parentE != -1 && i == gamma[parentE]) {
				continue
			}
			h.children = append(h.children, hingeEdge{e: *e, h: htrees[i]})
		}
	}

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

// RerootEdge reroots G at node covering edge, producing an isomorphic graph
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

// DecompHinge computes a decomposition of the original input graph,
// using the hingetree to speed up the computation
func (h Hingetree) DecompHinge(alg AlgorithmH, g Graph) Decomp {
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
