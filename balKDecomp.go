// Decomposition method based on "Balanced Decompositions" Akatov and Gottlob.
package main

import (
	"log"
	"reflect"
	"runtime"
)

type balKDecomp struct {
	graph Graph
}

func (d Decomp) checkWidth() int {
	var output = 0

	current := []Node{d.root}

	// iterate over decomp in BFS
	for len(current) > 0 {
		children := []Node{}
		for _, n := range current {
			if len(n.cover) > output {
				output = len(n.cover)
			}

			for _, c := range n.children {
				children = append(children, c) // build up the next level of the tree
			}
		}
		current = children
	}

	return output
}

// Takes the output of balKDecomp and ``blows it up'' to GHD
func (d *Decomp) blowup() Decomp {
	var output Decomp
	output.graph = d.graph
	output.root = d.root
	current := []Node{output.root}

	// iterate over decomp in BFS to add union
	for len(current) > 0 {
		children := []Node{}
		for _, n := range current {
			lambda := n.cover
			nchildren := n.children
			for _, c := range nchildren {
				// fmt.Println("Cover prior: ", c.cover)
				c.cover = removeDuplicateEdges(append(c.cover, lambda...)) // merge lambda with direct ancestor

				// fmt.Println("Cover after: ", c.cover)
				c.bag = Vertices(c.cover)      // fix the bag
				children = append(children, c) // build up the next level of the tree
			}
			n.children = nchildren
		}
		current = children
	}

	// fmt.Println("GHD WIDTH: ", output.checkWidth())
	return output
}

func (b balKDecomp) findDecomp(K int, H Graph) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	gen := getCombin(len(H.edges), K)

OUTER:
	for gen.hasNext() {
		balsep := getSubset(H.edges, gen.combination)

		log.Printf("Testing: %v\n", Graph{edges: balsep})
		gen.confirm()
		if !H.checkBalancedSep(balsep, []Special{}) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, _, _ := H.getComponents(balsep, []Special{})

		var subtrees []Node
		for _, c := range comps {
			decomp := b.findDecomp(K, c)
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{edges: balsep}, c)
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		node := Node{bag: Vertices(balsep), cover: balsep, children: subtrees}
		return Decomp{graph: H, root: node}
	}

	log.Printf("REJECT: Couldn't find balsep for H %v\n", H)
	return Decomp{} // using empty decomp as failure
}

func (b balKDecomp) findDecompParallelFull(K int, H Graph) Decomp {
	log.Printf("\n\nCurrent Subgraph: %v\n", H)

	var balsep []Edge
	var decomposed = false
	var subtrees []Node
	generators := splitCombin(len(H.edges), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		parallelSearch(H, []Special{}, H.edges, &found, generators)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v \n", H)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = getSubset(H.edges, found)

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, _, _ := H.getComponents(balsep, []Special{})

		ch := make(chan Decomp)
		for _, c := range comps {
			go func(K int, c Graph) {
				ch <- b.findDecompParallelFull(K, c)
			}(K, c)
		}

		for i := 0; i < len(comps); i++ {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v\n", Graph{edges: balsep})
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				subtrees = []Node{}
				continue OUTER
			}
			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}
		decomposed = true

	}

	node := Node{bag: Vertices(balsep), cover: balsep, children: subtrees}
	return Decomp{graph: H, root: node}
}

func (b balKDecomp) findBD(K int) Decomp {
	return b.findDecomp(K, b.graph)
}

func (b balKDecomp) findBDFullParallel(K int) Decomp {
	return b.findDecompParallelFull(K, b.graph)
}
