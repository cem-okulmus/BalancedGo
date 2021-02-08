package lib

import (
	"fmt"
	"reflect"
)

// A Decomp (short for Decomposition) consists of a labelled tree which
// subdivides a graph in a certain way
type Decomp struct {
	Graph         Graph
	Root          Node
	SkipRerooting bool //needed for BalDetK
	UpConnecting  bool //needed for DivideK
}

func (d Decomp) String() string {
	return d.Root.String()
}

// RestoreSubedges replaces any ad-hoc subedge with actual edges occurring in the graph
func (d *Decomp) RestoreSubedges() {
	if reflect.DeepEqual(*d, Decomp{}) { // don't change the empty decomp
		return
	}

	newRoot := d.Root.restoreEdges(d.Graph.Edges)

	d.Root = newRoot
}

// Correct checks if a decomp full fills the properties of a GHD when given a hypergraph g as input.
// It also checks for the special condition of HDs, though it merely prints a warning if it is not satisfied,
// the output is not affected by this additional check.
func (d Decomp) Correct(g Graph) bool {

	if reflect.DeepEqual(d, Decomp{}) { // empty Decomp is always false
		return false
	}

	//must be a decomp of same graph
	if !d.Graph.equal(g) {
		if d.Graph.Edges.Len() > 0 {
			fmt.Println("Decomp of different graph")
		} else {
			fmt.Println("Empty Decomp")
		}
		return false
	}

	//Every bag must be subset of the lambda label
	if !d.Root.bagSubsets() {
		fmt.Printf("Bags not subsets of edge labels")
		return false
	}

	// Every edge has to be covered
	for _, e := range d.Graph.Edges.Slice() {
		if !d.Root.coversEdge(e) {
			fmt.Println("Edge ", e, " isn't covered")
			return false
		}
	}

	//connectedness
	for _, i := range d.Graph.Edges.Vertices() {

		nodeCheck, _ := d.Root.connected(i, false)
		if !nodeCheck {
			mutex.RLock()
			fmt.Printf("Vertex %v doesn't span connected subtree\n", m[i])
			mutex.RUnlock()
			return false
		}
		// if d.connected(i) != nodeCheck {
		// 	log.Panicln("Node based connectedness check not working!")
		// }

	}

	//special condition (optionally)

	if !d.Root.noSCViolation() {
		fmt.Println("SCV found!. Not a valid hypertree decomposition!")
	}

	return true
}

// CheckWidth returns the size of the largest bag of any node in a decomp
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
