// Decomposition method based on "Balanced Decompositions" Akatov and Gottlob.
package algorithms

import (
	"log"
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type BalKDecomp struct {
	Graph     Graph
	BalFactor int
}

func (b BalKDecomp) findDecomp(K int, H Graph) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	gen := GetCombin(H.Edges.Len(), K)

OUTER:
	for gen.HasNext() {
		balsep := GetSubset(H.Edges, gen.Combination)

		log.Printf("Testing: %v\n", Graph{Edges: balsep})
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, []Special{}, b.BalFactor) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, _, _ := H.GetComponents(balsep, []Special{})

		var subtrees []Node
		for _, c := range comps {
			decomp := b.findDecomp(K, c)
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{Edges: balsep}, c)
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.Root)
		}

		node := Node{Bag: balsep.Vertices(), Cover: balsep, Children: subtrees}
		return Decomp{Graph: H, Root: node}
	}

	log.Printf("REJECT: Couldn't find balsep for H %v\n", H)
	return Decomp{} // using empty decomp as failure
}

func (b BalKDecomp) findDecompParallelFull(K int, H Graph) Decomp {
	log.Printf("\n\nCurrent SubGraph: %v\n", H)

	var balsep Edges
	var decomposed = false
	var subtrees []Node
	generators := SplitCombin(H.Edges.Len(), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		parallelSearch(H, []Special{}, H.Edges, &found, generators, b.BalFactor)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v \n", H)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(H.Edges, found)

		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, _, _ := H.GetComponents(balsep, []Special{})

		ch := make(chan Decomp)
		for _, c := range comps {
			go func(K int, c Graph) {
				ch <- b.findDecompParallelFull(K, c)
			}(K, c)
		}

		for i := 0; i < len(comps); i++ {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v\n", Graph{Edges: balsep})
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				subtrees = []Node{}
				continue OUTER
			}
			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.Root)
		}
		decomposed = true

	}

	node := Node{Bag: balsep.Vertices(), Cover: balsep, Children: subtrees}
	return Decomp{Graph: H, Root: node}
}

func (b BalKDecomp) FindBD(K int) Decomp {
	return b.findDecomp(K, b.Graph)
}

func (b BalKDecomp) FindBDFullParallel(K int) Decomp {
	return b.findDecompParallelFull(K, b.Graph)
}

func (b BalKDecomp) FindDecomp(K int) Decomp {
	return b.findDecompParallelFull(K, b.Graph)
}

func (b BalKDecomp) Name() string {
	return "Akatov"
}
