package lib

import (
	"fmt"
	"log"
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

func (d *Decomp) RestoreSubedges() {
	newRoot := d.Root.RestoreEdges(d.Graph.Edges)

	d.Root = newRoot
}

// func (d Decomp) connected(vert int) bool {
// 	conGraph := d.Root.getConGraph(true)
// 	var containingNodes = d.Root.allChildrenContaining(vert)
// 	var edgesContaining = FilterVerticesStrict(conGraph, containingNodes)

// 	// log.Printf("All nodes containing %s\n", m[vert])
// 	// for _, n := range containingNodes {
// 	// 	log.Printf("%v ,", m[n])
// 	// }

// 	// log.Printf("All edges")
// 	// for _, e := range edgesContaining.Slice() {
// 	// 	log.Printf("%v ,", e)
// 	// }

// 	g, _, _ := Graph{Edges: edgesContaining}.GetComponents(Edges{}, []Special{})

// 	// if len(g) > 1 {
// 	// 	fmt.Printf("Components of Node %s\n", m[vert])
// 	// 	for _, c := range g {
// 	// 		fmt.Printf("%v \n\n", c)
// 	// 	}
// 	// }

// 	return len(g) == 1
// }

func (d Decomp) Correct(g Graph) bool {

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
		for i := range current {
			lambda := current[i].Cover
			nchildren := current[i].Children
			for j := range nchildren {
				var nuCover []Edge
				nuCover = append(nchildren[j].Cover.Slice(), lambda.Slice()...) // merge lambda with direct ancestor
				nchildren[j].Cover = removeDuplicateEdges(nuCover)

				nchildren[j].Bag = nchildren[j].Cover.Vertices() // fix the bag
				children = append(children, nchildren[j])        // build up the next level of the tree
			}
			current[i].Children = nchildren
		}
		current = children
	}

	return output
}

type SceneValue struct {
	Sep        Edges
	Perm       bool // one-time cached if false
	WoundingUp bool // created during wounding up
}

type Scene struct {
	Sub []int
	Val SceneValue
}

func (s SceneValue) String() string {
	var added string
	if s.WoundingUp {
		added = fmt.Sprint("WoundingUp")

	}
	return fmt.Sprint("Sep ", s.Sep, "Perm: ", s.Perm) + added
}

func (n Node) woundingDown(input Graph) []Scene {

	// fmt.Println("\n\n\nCurrent subhypergraph: ", PrintVertices(input.Vertices()))
	// fmt.Println("Current node:\n Bag: ", PrintVertices(n.Bag), "\n Cover:", n.Cover)

	var output []Scene

	if !Subset(n.Bag, n.Cover.Vertices()) {
		// start wounding up procedure
		outputChild, _ := n.woundingUp(input.Edges.Slice())

		output = append(output, outputChild...)
		return output
	}

	sep := n.Cover.IntersectWith(n.Bag)
	comps, _, _ := input.GetComponents(sep, []Special{})

	if len(n.Children) != len(comps) {
		// start wounding up procedure
		outputChild, _ := n.woundingUp(input.Edges.Slice())

		output = append(output, outputChild...)
		return output
	}

	// fmt.Println("\nSep from decomp: ", PrintVertices(sep.Vertices()))

	output = append(output, Scene{Sub: input.Vertices(), Val: SceneValue{Sep: sep, Perm: !n.containsMarked()}})

	// fmt.Println("\nCurrent components: ")
	// for i, c := range comps {

	// 	fmt.Println(i, ".) ", PrintVertices(c.Vertices()), "\n ")
	// }

OUTER:
	for _, u := range n.Children {

		// fmt.Println("\nVertices of Child : ", PrintVertices(u.Bag))

	INNER:
		for _, c := range comps {
			if len(Inter(Diff(c.Vertices(), sep.Vertices()), u.Bag)) == 0 { // check if this node belongs to this subgraph
				continue INNER
			}

			outputChild := u.woundingDown(c)
			output = append(output, outputChild...)

			continue OUTER
		}

		log.Panicln("\nCouldn't find matching subgraph!")

	}

	return output

}

func (n Node) woundingUp(edges []Edge) ([]Scene, []int) {

	var output []Scene
	var coveredVertices []int
	var coveredBelow []int

	for _, c := range n.Children {
		outputChild, coveredChild := c.woundingUp(edges)

		output = append(output, outputChild...)
		coveredBelow = append(coveredBelow, coveredChild...)
	}

	// if n.belowMarked(root) {
	//
	// }
	coveredSlice := []Edge{}
	var coveredEdges Edges

	if !n.Star { // skip if

		for i := range edges {
			if Subset(edges[i].Vertices, n.Bag) {
				coveredSlice = append(coveredSlice, edges[i])
			}
		}

		coveredEdges = NewEdges(coveredSlice)

	}
	coveredVertices = RemoveDuplicates(append(coveredEdges.Vertices(), coveredBelow...))

	if !n.Star {

		sep := n.Cover.IntersectWith(n.Bag)

		output = append(output, Scene{Sub: coveredVertices, Val: SceneValue{Sep: sep, Perm: true, WoundingUp: true}})
	}

	return output, coveredVertices

}

func (d Decomp) SceneCreation(input Graph) map[uint32]SceneValue {

	var output map[uint32]SceneValue
	output = make(map[uint32]SceneValue)

	// start_wd := time.Now()

	scenes := d.Root.woundingDown(input)

	// msec_wd := time.Now().Sub(start_wd).Seconds() * float64(time.Second/time.Millisecond)
	// fmt.Println("Wounding Down", msec_wd, " ms")

	// start_wu := time.Now()

	// scenes = append(scenes, d.Root.woundingUp(d.Root)...)

	// msec_wu := time.Now().Sub(start_wu).Seconds() * float64(time.Second/time.Millisecond)
	// fmt.Println("Wounding Up", msec_wu, " ms")

	// fmt.Println("Found scenes, ", len(scenes))
	for _, s := range scenes {
		output[IntHash(s.Sub)] = s.Val
	}

	return output
}
