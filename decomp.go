package main

import (
	"fmt"
	"reflect"
)

// A Decomp (short for Decomposition) consists of a labelled tree which
// subdivides a graph in a certain way
type Decomp struct {
	graph Graph
	root  Node
}

func (d Decomp) String() string {
	return d.root.String()
}

func (d Decomp) connected(vert int) bool {
	var containingNodes = d.root.allChildrenContaining(vert, 0)
	var edgesContaining = filterVertices(d.root.getConGraph(0), containingNodes)

	// log.Printf("All nodes containing %s\n", m[vert])
	// for _, n := range containingNodes {
	// 	log.Printf("%v ,", m[n])
	// }

	// log.Printf("All edges")
	// for _, e := range edgesContaining {
	// 	log.Printf("%v ,", e)
	// }

	g, _, _ := Graph{edges: edgesContaining}.getComponents([]Edge{}, []Special{})

	// log.Printf("Components of Node %s\n", m[vert])
	// for _, c := range g {
	// 	log.Printf("%v \n\n", c)
	// }

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

	//special condition (optionally)

	if !d.root.noSCViolation() {
		fmt.Println("SCV found!. Not a valid hypertree!")
	}

	return true
}
