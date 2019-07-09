package lib

import (
	"fmt"
	"reflect"
)

// A Decomp (short for Decomposition) consists of a labelled tree which
// subdivides a graph in a certain way
type Decomp struct {
	Graph Graph
	Root  Node
}

func (d Decomp) String() string {
	return d.Root.String()
}

func (d *Decomp) RestoreSubedges() {
	newRoot := d.Root.RestoreEdges(d.Graph.Edges)

	d.Root = newRoot
}

func (d Decomp) connected(vert int) bool {
	var containingNodes = d.Root.allChildrenContaining(vert, 0)
	var edgesContaining = FilterVertices(d.Root.getConGraph(0), containingNodes)

	// log.Printf("All nodes containing %s\n", m[vert])
	// for _, n := range containingNodes {
	// 	log.Printf("%v ,", m[n])
	// }

	// log.Printf("All edges")
	// for _, e := range edgesContaining {
	// 	log.Printf("%v ,", e)
	// }

	g, _, _ := Graph{Edges: edgesContaining}.GetComponents([]Edge{}, []Special{})

	// log.Printf("Components of Node %s\n", m[vert])
	// for _, c := range g {
	// 	log.Printf("%v \n\n", c)
	// }

	return len(g) == 1
}

func (d Decomp) Correct(g Graph) bool {

	//must be a decomp of same graph
	if !reflect.DeepEqual(d.Graph.Edges, g.Edges) {
		return false
	}

	//Every bag must be subset of the lambda label
	if !d.Root.bagSubsets() {
		fmt.Printf("Bags not subsets of edge labels")
		return false
	}

	// Every edge has to be covered
	for _, e := range d.Graph.Edges {
		if !d.Root.coversEdge(e) {
			fmt.Printf("Edge %v isn't covered", e)
			return false
		}
	}

	//connectedness
	for _, i := range Vertices(d.Graph.Edges) {
		if !d.connected(i) {
			fmt.Printf("Node %v doesn't span connected subtree\n", m[i])
			return false
		}
	}

	//special condition (optionally)

	if !d.Root.noSCViolation() {
		fmt.Println("SCV found!. Not a valid hypertree!")
	}

	return true
}

func (d Decomp) CheckWidth() int {
	var output = 0

	current := []Node{d.Root}

	// iterate over decomp in BFS
	for len(current) > 0 {
		children := []Node{}
		for _, n := range current {
			if len(n.Cover) > output {
				output = len(n.Cover)
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
				c.Cover = removeDuplicateEdges(append(c.Cover, lambda...)) // merge lambda with direct ancestor

				// fmt.Println("Cover after: ", c.Cover)
				c.Bag = Vertices(c.Cover)      // fix the bag
				children = append(children, c) // build up the next level of the tree
			}
			n.Children = nchildren
		}
		current = children
	}

	// fmt.Println("GHD WIDTH: ", output.checkWidth())
	return output
}
