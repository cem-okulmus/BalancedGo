// Functions that simplify graphs, and transform decompositions of the simplified graph back to decomposition of original graph

package lib

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
	return fmt.Sprintf("(%v ⊆ %v)", m[e.subedge.Name], m[e.parent.Name])
}

type vertOp struct {
	vertex int
	edge   Edge
}

func (_ vertOp) isGYÖ() {}

func (v vertOp) String() string {
	return fmt.Sprintf("(%v ∈ %v)", m[v.vertex], m[v.edge.Name])
}

//TODO fix this to run in linear time
// Performs one part of GYÖ reduct
func (g Graph) removeEdges() (Graph, []GYÖReduct) {
	var output Edges
	var removed Edges
	var ops []GYÖReduct

OUTER:
	for _, e1 := range g.Edges.Slice {
		for _, e2 := range g.Edges.Slice {
			if e1.Name != e2.Name && !e2.containedIn(removed.Slice) && Subset(e1.Vertices, e2.Vertices) {
				ops = append(ops, edgeOp{subedge: e1, parent: e2})
				removed.append(e1)
				continue OUTER
			}
		}

		output.append(e1)
	}

	return Graph{Edges: output}, ops
}

func (g Graph) removeVertices() (Graph, []GYÖReduct) {
	var ops []GYÖReduct
	var edges Edges

	for _, e1 := range g.Edges.Slice {
		var vertices []int
		//	fmt.Println("Working on edge ", e1)
	INNER:
		for _, v := range e1.Vertices {
			//		fmt.Printf("Degree of %v is %v\n", m[v], getDegree(g.Edges, v))
			if getDegree(g.Edges, v) == 1 {
				ops = append(ops, vertOp{vertex: v, edge: e1})
				continue INNER
			}
			vertices = append(vertices, v)
		}
		if len(vertices) > 0 {
			edges.append(Edge{Name: e1.Name, Vertices: vertices})
		}

	}

	return Graph{Edges: edges}, ops
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
		// for _, e := range g2.Edges {
		// 	fmt.Printf("%v %v\n", e, Edge{Vertices: e.Vertices})
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
	if e.parent.containedIn(n.Cover.Slice) {
		n.Children = append(n.Children, Node{Bag: e.subedge.Vertices, Cover: Edges{Slice: []Edge{e.subedge}}})
		return n, true // Won't work without deep copy
	}

	for _, child := range n.Children {
		res, b := child.restoreEdgeOp(e)
		if b {
			child = res // updating this element!
			return Node{Bag: n.Bag, Cover: n.Cover, Children: n.Children}, true
		}
	}

	return Node{}, false
}

func (n Node) restoreVertex(v vertOp) (Node, bool) {
	if Subset(v.edge.Vertices, n.Bag) {
		return Node{Bag: append(n.Bag, v.vertex), Cover: n.Cover, Children: n.Children}, true
	}

	for _, child := range n.Children {
		res, b := child.restoreVertex(v)
		if b {
			child = res // updating this element!
			return Node{Bag: n.Bag, Cover: n.Cover, Children: n.Children}, true
		}
	}

	return Node{}, false

}

/*
Type Collapse
*/

func (g Graph) getType(vertex int) *big.Int {
	output := new(big.Int)

	for i := range g.Edges.Slice {
		if Mem(g.Edges.Slice[i].Vertices, vertex) {
			output.SetBit(output, i, 1)
		}
	}
	return output
}

// Possible optimization: When computing the distances, use the matrix to speed up type detection
func (g Graph) TypeCollapse() (Graph, map[int][]int, int) {
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

	for _, e := range g.Edges.Slice {
		var vertices []int
		for _, v := range e.Vertices {
			vertices = append(vertices, substituteMap[v])
		}
		newEdges.append(Edge{Name: e.Name, Vertices: RemoveDuplicates(vertices)})
	}

	return Graph{Edges: newEdges}, restorationMap, count
}
