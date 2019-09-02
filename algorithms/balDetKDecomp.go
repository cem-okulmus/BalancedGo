// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to DetKDecomp
package algorithms

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type BalDetKDecomp struct {
	Graph     Graph
	BalFactor int
	Depth     int // how many rounds of balSep are used
}

func decrease(count int) int {
	output := count - 1
	if output < 1 {
		return 0
	} else {
		return output
	}
}

func (b BalDetKDecomp) findDecompBalSep(K int, currentDepth int, H Graph, Sp []Special) Decomp {
	log.Printf("Current SubGraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(b.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	var balsep Edges

	var decomposed = false
	edges := FilterVerticesStrict(b.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators, b.BalFactor)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)

		log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

		comps, compsSp, _ := H.GetComponents(balsep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		ch := make(chan Decomp)
		for i := range comps {

			if currentDepth > 0 {
				go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
					ch <- b.findDecompBalSep(K, decrease(currentDepth), comps[i], append(compsSp[i], SepSpecial))
				}(K, i, comps, compsSp, SepSpecial)
			} else {
				go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
					det := DetKDecomp{Graph: b.Graph, BalFactor: b.BalFactor, SubEdge: true}
					ch <- det.findDecomp(K, comps[i], []int{}, append(compsSp[i], SepSpecial))
				}(K, i, comps, compsSp, SepSpecial)
			}

		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {

				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %+v\n", decomp)
			if currentDepth == 1 {
				fmt.Println("From detK with Special Edges ", append(compsSp[i], SepSpecial), ":\n", decomp)
			}

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	return rerooting(H, balsep, subtrees)
}

func (b BalDetKDecomp) FindGHD(K int) Decomp {
	return b.findDecompBalSep(K, b.Depth, b.Graph, []Special{})
}
