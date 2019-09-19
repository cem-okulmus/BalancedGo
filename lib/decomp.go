package lib

import (
	"fmt"
)

// A Decomp (short for Decomposition) consists of a labelled tree which
// subdivides a graph in a certain way
type Decomp struct {
	Graph         Graph
	Root          Node
	SkipRerooting bool //needed for BalDetK
}

func (d Decomp) String() string {
	return d.Root.String()
}

func (d *Decomp) RestoreSubedges() {
	newRoot := d.Root.RestoreEdges(d.Graph.Edges)

	d.Root = newRoot
}

func (d Decomp) connected(vert int) bool {
	conGraph := d.Root.getConGraph(true)
	var containingNodes = d.Root.allChildrenContaining(vert)
	var edgesContaining = FilterVertices(conGraph, containingNodes)

	// log.Printf("All nodes containing %s\n", m[vert])
	// for _, n := range containingNodes {
	// 	log.Printf("%v ,", m[n])
	// }

	// log.Printf("All edges")
	// for _, e := range edgesContaining {
	// 	log.Printf("%v ,", e)
	// }

	g, _, _ := Graph{Edges: edgesContaining}.GetComponents(Edges{}, []Special{})

	// if len(g) > 1 {
	// 	fmt.Printf("Components of Node %s\n", m[vert])
	// 	for _, c := range g {
	// 		fmt.Printf("%v \n\n", c)
	// 	}
	// }

	return len(g) == 1
}

func (d Decomp) Correct(g Graph) bool {
	output := true

	//must be a decomp of same graph
	if !d.Graph.equal(g) {
		if d.Graph.Edges.Len() > 0 {
			fmt.Println("Decomp of different graph")
		} else {
			fmt.Println("Empty Decomp")
		}
		output = false
	}

	//Every bag must be subset of the lambda label
	if !d.Root.bagSubsets() {
		fmt.Printf("Bags not subsets of edge labels")
		output = false
	}

	// Every edge has to be covered
	for _, e := range d.Graph.Edges.Slice() {
		if !d.Root.coversEdge(e) {
			fmt.Println("Edge ", e, " isn't covered")
			output = false
		}
	}

	//connectedness
	for _, i := range d.Graph.Edges.Vertices() {
		if !d.connected(i) {
			mutex.RLock()
			fmt.Printf("Vertex %v doesn't span connected subtree\n", m[i])
			mutex.RUnlock()
			output = false
		}

	}

	//special condition (optionally)

	if !d.Root.noSCViolation() {
		fmt.Println("SCV found!. Not a valid hypertree!")
	}

	return output
}

func (d Decomp) CheckWidth() int {
	var output = 0

	current := []Node{d.Root}

	// iterate over decomp in BFS
	for len(current) > 0 {
		children := []Node{}
		for _, n := range current {
			if n.Cover.Len() > output {
				output = n.Cover.Len()
			}

			for _, c := range n.Children {
				children = append(children, c) // build up the next level of the tree
			}
		}
		current = children
	}

	return output
}

// Takes the output of balKDecomp and ``blows it up'' to GHD
func (d *Decomp) Blowup() Decomp {
	var output Decomp
	output.Graph = d.Graph
	output.Root = d.Root
	current := []Node{output.Root}

	// iterate over decomp in BFS to add union
	for len(current) > 0 {
		children := []Node{}
		for _, n := range current {
			lambda := n.Cover
			nchildren := n.Children
			for _, c := range nchildren {
				// fmt.Println("Cover prior: ", c.Cover)
				c.Cover.Append(lambda.Slice()...)
				c.Cover = removeDuplicateEdges(c.Cover.Slice()) // merge lambda with direct ancestor

				// fmt.Println("Cover after: ", c.Cover)
				c.Bag = c.Cover.Vertices()     // fix the bag
				children = append(children, c) // build up the next level of the tree
			}
			n.Children = nchildren
		}
		current = children
	}

	// fmt.Println("GHD WIDTH: ", output.checkWidth())
	return output
}
