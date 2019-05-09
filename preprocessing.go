// Functions that simplify graphs, and allow for the transformation decompositions of simplfied graphs to decompositions of their original graphs

package main

import (
	"math/big"
	"reflect"
)

type OW struct {
	earmark Edge
	parent  Edge
}

func (e Edge) OWcheck(l []Edge) (bool, Edge) {
	var parent Edge
	var intersect []int

	for _, o := range l {
		temp_intersect := inter(o.vertices, e.vertices)
		if len(intersect) == 0 && len(temp_intersect) != 0 {
			intersect = temp_intersect
			parent = o
		} else {
			if !reflect.DeepEqual(temp_intersect, intersect) {
				return false, e
			}
		}
	}

	return true, parent
}

func (g Graph) OWremoval() (Graph, []OW) {
	var newedges Edges
	var earmarks []OW

	for _, e := range g.edges {
		check, p := e.OWcheck(g.edges)
		if check {
			earmarks = append(earmarks, OW{earmark: e, parent: p})
		} else {
			newedges = append(newedges, e)
		}
	}

	return Graph{edges: newedges}, earmarks
}

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
func (g Graph) typeCollapse() (Graph, map[int][]int) {

	substituteMap := make(map[int]int)    // to keep track of which vertices to collapse
	restorationMap := make(map[int][]int) // used to restore "full" edges from simplified one

	// identify vertices to replace
	encountered := make(map[string]int)

	for _, v := range g.Vertices() {
		typeString := g.getType(v).String()

		if _, ok := encountered[typeString]; ok {
			// already seen this type before
			substituteMap[v] = encountered[typeString]
			restorationMap[encountered[typeString]] = append(restorationMap[encountered[typeString]], v)
		} else {
			// Record thie type as a new element
			encountered[typeString] = v
		}
	}

	newEdges := g.edges

	for _, e := range newEdges {
		for _, v := range e.vertices {
			v, _ = substituteMap[v]
		}
		e.vertices = removeDuplicates(e.vertices)
	}

	return Graph{edges: newEdges}, restorationMap
}
