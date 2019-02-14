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

func (d Decomp) correct(g Graph) bool {

	//must be a decomp of same graph
	if !reflect.DeepEqual(d.graph, g) {
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
		if !d.root.connected(i) {
			fmt.Printf("Node %v doesn't span connected subtree\n", d.graph.m[i])
			return false
		}
	}

	return true
}
