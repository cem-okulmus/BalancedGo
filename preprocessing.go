// Functions that simplify graphs, and transform decompositions of the simplified graph back to decomposition of original graph

package main

import (
	"fmt"
	"math/big"
)

// A GYÖReduct (that's short for GYÖ (Graham - Yu - Özsoyoğlu) Reduction )
// consists of a list of operations that simplify a graph by 1) removing isolated
// vertices or 2) removing edges fully contained in other edges (and applying
// these two operations iteratively, until convergance)
type GYÖReduct interface {
	isGYÖ()
}

type edgeOp struct {
	subedge Edge
	parent  Edge
}

func (_ edgeOp) isGYÖ() {}

func (e edgeOp) String() string {
	return fmt.Sprintf("(%v ⊆ %v)", m[e.subedge.name], m[e.parent.name])
}

type vertOp struct {
	vertex int
	edge   Edge
}

func (_ vertOp) isGYÖ() {}

func (v vertOp) String() string {
	return fmt.Sprintf("(%v ∈ %v)", m[v.vertex], m[v.edge.name])
}

// Performs one part of GYÖ reduct
func (g Graph) removeEdges() (Graph, []GYÖReduct) {
	var output Edges
	var removed Edges
	var ops []GYÖReduct

OUTER:
	for _, e1 := range g.edges {
		for _, e2 := range g.edges {
			if e1.name != e2.name && !e2.containedIn(removed) && subset(e1.vertices, e2.vertices) {
				ops = append(ops, edgeOp{subedge: e1, parent: e2})
				removed.append(e1)
				continue OUTER
			}
		}

		output.append(e1)
	}

	return Graph{edges: output}, ops
}

func (g Graph) removeVertices() (Graph, []GYÖReduct) {
	var ops []GYÖReduct
	var edges Edges

	for _, e1 := range g.edges {
		var vertices []int
		//	fmt.Println("Working on edge ", e1)
	INNER:
		for _, v := range e1.vertices {
			//		fmt.Printf("Degree of %v is %v\n", m[v], getDegree(g.edges, v))
			if getDegree(g.edges, v) == 1 {
				ops = append(ops, vertOp{vertex: v, edge: e1})
				continue INNER
			}
			vertices = append(vertices, v)
		}
		if len(vertices) > 0 {
			edges.append(Edge{name: e1.name, vertices: vertices})
		}

	}

	return Graph{edges: edges}, ops
}

func (g Graph) GYÖReduct() (Graph, []GYÖReduct) {
	var ops []GYÖReduct

	for {
		//Perform edge removal
		g1, ops1 := g.removeEdges()
		// fmt.Println("After Edge Removal:")
		// fmt.Println(g1)

		ops = append(ops, ops1...)

		//Perform vertex removal
		g2, ops2 := g1.removeVertices()
		// fmt.Println("After Vertex Removal:")
		// for _, e := range g2.edges {
		// 	fmt.Printf("%v %v\n", e, Edge{vertices: e.vertices})
		// }

		ops = append(ops, ops2...)

		//Check if something changed
		if len(ops2)+len(ops1) == 0 {
			break
		}
		g = g2
	}

	//reverse order of ops
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}

	return g, ops
}

func (n Node) restoreEdgeOp(e edgeOp) (Node, bool) {
	if e.parent.containedIn(n.cover) {
		n.children = append(n.children, Node{bag: e.subedge.vertices, cover: []Edge{e.subedge}})
		return n, true // Won't work without deep copy
	}

	for _, child := range n.children {
		res, b := child.restoreEdgeOp(e)
		if b {
			child = res // updating this element!
			return Node{bag: n.bag, cover: n.cover, children: n.children}, true
		}
	}

	return Node{}, false
}

func (n Node) restoreVertex(v vertOp) (Node, bool) {
	if subset(v.edge.vertices, n.bag) {
		return Node{bag: append(n.bag, v.vertex), cover: n.cover, children: n.children}, true
	}

	for _, child := range n.children {
		res, b := child.restoreVertex(v)
		if b {
			child = res // updating this element!
			return Node{bag: n.bag, cover: n.cover, children: n.children}, true
		}
	}

	return Node{}, false

}

/*
Type Collapse
*/

func (g Graph) getType(vertex int) *big.Int {
	output := new(big.Int)

	for i := range g.edges {
		if mem(g.edges[i].vertices, vertex) {
			output.SetBit(output, i, 1)
		}
	}
	return output
}

// Possible optimization: When computing the distances, use the matrix to speed up type detection
func (g Graph) typeCollapse() (Graph, map[int][]int, int) {
	count := 0

	substituteMap := make(map[int]int)    // to keep track of which vertices to collapse
	restorationMap := make(map[int][]int) // used to restore "full" edges from simplified one

	// identify vertices to replace
	encountered := make(map[string]int)

	for _, v := range g.Vertices() {
		typeString := g.getType(v).String()
		// fmt.Println("Type of ", m[v], "is ", typeString)

		if _, ok := encountered[typeString]; ok {
			// already seen this type before
			// fmt.Println("Seen type of ", m[v], "before!")
			count++
			substituteMap[v] = encountered[typeString]
			restorationMap[encountered[typeString]] = append(restorationMap[encountered[typeString]], v)
		} else {
			// Record thie type as a new element
			encountered[typeString] = v
			substituteMap[v] = v
		}
	}

	var newEdges Edges

	for _, e := range g.edges {
		var vertices []int
		for _, v := range e.vertices {
			vertices = append(vertices, substituteMap[v])
		}
		newEdges.append(Edge{name: e.name, vertices: removeDuplicates(vertices)})
	}

	return Graph{edges: newEdges}, restorationMap, count
}
