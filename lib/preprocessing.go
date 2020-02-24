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
	mutex.RLock()
	defer mutex.RUnlock()
	return fmt.Sprintf("(%v ⊆ %v)", m[e.subedge.Name], m[e.parent.Name])
}

type vertOp struct {
	vertex int
	edge   Edge
}

func (_ vertOp) isGYÖ() {}

func (v vertOp) String() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return fmt.Sprintf("(%v ∈ %v)", m[v.vertex], m[v.edge.Name])
}

//TODO fix this to run in linear time
// Performs one part of GYÖ reduct
func (g Graph) removeEdges() (Graph, []GYÖReduct) {
	var output []Edge
	var removed []Edge
	var ops []GYÖReduct

OUTER:
	for _, e1 := range g.Edges.Slice() {
		for _, e2 := range g.Edges.Slice() {
			if e1.Name != e2.Name && !e2.containedIn(removed) && Subset(e1.Vertices, e2.Vertices) {
				ops = append(ops, edgeOp{subedge: e1, parent: e2})
				removed = append(removed, e1)
				continue OUTER
			}
		}

		output = append(output, e1)
	}

	return Graph{Edges: NewEdges(output)}, ops
}

func (g Graph) removeVertices() (Graph, []GYÖReduct) {
	var ops []GYÖReduct
	var edges []Edge

	for _, e1 := range g.Edges.Slice() {
		var vertices []int
		//	fmt.Println("Working on edge ", e1)
	INNER:
		for _, v := range e1.Vertices {
			//		fmt.Printf("Degree of %v is %v\n", m[v], getDegree(g.Edges, v))
			if getDegree(g.Edges, v) == 1 {
				ops = append(ops, vertOp{vertex: v, edge: Edge{Name: e1.Name, Vertices: vertices}})
				continue INNER
			}
			vertices = append(vertices, v)
		}
		if len(vertices) > 0 {
			edges = append(edges, Edge{Name: e1.Name, Vertices: vertices})
		}

	}

	return Graph{Edges: NewEdges(edges)}, ops
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
		// for _, e := range g2.Edges.Slice() {
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
	if Subset(e.parent.Vertices, n.Bag) {
		n.Children = append(n.Children, Node{Bag: e.subedge.Vertices, Cover: Edges{slice: []Edge{e.subedge}}})
		return n, true // Won't work without deep copy
	}

	for i := range n.Children {
		res, b := n.Children[i].restoreEdgeOp(e)
		if b {
			n.Children[i] = res // updating this element!
			return n, true      // Won't work without deep copy
		}
	}

	return n, false
}

func (n Node) restoreVertex(v vertOp) (Node, bool) {
	if len(n.Bag) == 0 && n.Cover.Len() == 0 && len(n.Children) == 0 {
		edge := Edge{Name: v.edge.Name, Vertices: []int{v.vertex}}
		return Node{Bag: []int{v.vertex}, Cover: NewEdges([]Edge{edge})}, true
	}

	if v.edge.containedIn(n.Cover.Slice()) {

		nuCover := []Edge{}

		for _, e := range n.Cover.Slice() {
			if e.Name == v.edge.Name {
				edge := Edge{Name: e.Name, Vertices: append(e.Vertices, v.vertex)}
				nuCover = append(nuCover, edge)
			} else {
				nuCover = append(nuCover, e)
			}
		}

		return Node{Bag: append(n.Bag, v.vertex), Cover: NewEdges(nuCover), Children: n.Children}, true
	}

	for i := range n.Children {
		res, b := n.Children[i].restoreVertex(v)
		if b {
			n.Children[i] = res // updating this element!
			return n, true
		}
	}

	return n, false
}

func (n Node) RestoreGYÖ(reducts []GYÖReduct) (Node, bool) {

	output := n
	result := true

	for _, r := range reducts {
		switch v := r.(type) {
		case vertOp:
			output, result = output.restoreVertex(v)
		case edgeOp:
			output, result = output.restoreEdgeOp(v)
		}
		if !result {
			return output, false
		}
	}

	return output, true
}

/*
Type Collapse
*/

func (g Graph) getType(vertex int) *big.Int {
	output := new(big.Int)

	for i := range g.Edges.Slice() {
		if Mem(g.Edges.Slice()[i].Vertices, vertex) {
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

	var newEdges []Edge

	for _, e := range g.Edges.Slice() {
		var vertices []int
		for _, v := range e.Vertices {
			vertices = append(vertices, substituteMap[v])
		}
		newEdges = append(newEdges, Edge{Name: e.Name, Vertices: RemoveDuplicates(vertices)})
	}

	return Graph{Edges: NewEdges(newEdges)}, restorationMap, count
}

func (e Edges) addVertex(target int, oldVertices []int) Edges {
	edges := e.Slice()

	for i := range edges {
		if Mem(edges[i].Vertices, target) {
			edges[i].Vertices = append(edges[i].Vertices, oldVertices...)
		}
	}

	return NewEdges(edges)
}

func (n Node) addVertices(target int, oldVertices []int) (Node, bool) {
	found := false
	if Mem(n.Bag, target) {
		n.Bag = append(n.Bag, oldVertices...)
		n.Cover = n.Cover.addVertex(target, oldVertices)
		found = true
		found = true
	}

	for i := range n.Children {
		res, b := n.Children[i].addVertices(target, oldVertices)
		if b {
			n.Children[i] = res // replace child with updated res
			found = true
		}
	}

	if !found {
		return Node{}, false
	} else {
		return n, true
	}

}

func (n Node) RestoreTypes(restoreMap map[int][]int) (Node, bool) {

	output := n

	for k, v := range restoreMap {
		res, b := output.addVertices(k, v)
		if !b {
			return Node{}, false
		}
		output = res
	}

	return output, true
}
