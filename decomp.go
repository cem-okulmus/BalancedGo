package main

import (
	"fmt"
	"reflect"
)

type Decomp struct {
	graph Graph
	root  Node
}

func (d Decomp) String() string {
	return d.root.String()
}

func (d Decomp) connected(vert int) bool {
	var containingNodes = d.root.allChildrenContaining(vert, 0)
	var edgesContaining = d.root.getConGraph(vert, 0)

	g, _, _ := Graph{edges: edgesContaining}.getCompGeneral(containingNodes, []Edge{}, []Special{})

	return len(g) == 1
}

func (d Decomp) correct(g Graph) bool {

	//must be a decomp of same graph
	if !reflect.DeepEqual(d.graph, g) {
		return false
	}

	//Every bag must be subset of the lambda label
	if !d.root.bagSubsets() {
		fmt.Printf("Bags not subsets of edge labels")
		return false
	}

	// Every edge has to be covered
	for _, e := range d.graph.edges {
		if !d.root.coversEdge(e) {
			fmt.Printf("Edge %v isn't covered", e)
			return false
		}
	}

	//connectedness
	for _, i := range Vertices(d.graph.edges) {
		if !d.connected(i) {
			fmt.Printf("Node %v doesn't span connected subtree\n", m[i])
			return false
		}
	}

	return true
}
