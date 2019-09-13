// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to DetKDecomp
package algorithms

import (
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

	//find a balanced separator
	var decomposed = false
	edges := CutEdges(b.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), true)
	var subtrees []Decomp

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators, b.BalFactor)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("balDet REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)
		balsepOrig := balsep
		var sepSub *SepSub

		log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

	INNER:
		for !decomposed {
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
						Sp := append(compsSp[i], SepSpecial)
						if comps[i].Edges.Len() == 0 {
							edgesFromSpecial := EdgesSpecial(Sp)
							comps[i].Edges.Append(edgesFromSpecial...)
						}
						det.cache = make(map[uint32]*CompCache)
						ch <- det.findDecomp(K, comps[i], balsep.Vertices(), Sp)
					}(K, i, comps, compsSp, SepSpecial)
				}

			}

			for i := 0; i < len(comps); i++ {
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(b.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep}, Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							// 	log.Println(s.Combination)
							// }
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if H.CheckBalancedSep(balsep, Sp, b.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()), " of ", balsepOrig, " , ", Sp)
					continue INNER
				}

				log.Printf("Produced Decomp: %+v\n", decomp)
				if currentDepth == 0 {
					log.Println("\nFrom detK with Special Edges ", append(compsSp[i], SepSpecial), ":\n", decomp)
					// local := BalSepGlobal{Graph: b.Graph, BalFactor: b.BalFactor}
					// decomp_deux := local.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
					// fmt.Println("Output from Balsep: ", decomp_deux)

				}

				subtrees = append(subtrees, decomp)
			}

			decomposed = true
		}
	}

	if currentDepth > 0 {
		return rerooting(H, balsep, subtrees)
	} else {
		output := Node{Bag: balsep.Vertices(), Cover: balsep}

		for _, s := range subtrees {
			output.Children = append(output.Children, s.Root)
		}

		return Decomp{Graph: H, Root: output}
	}
}

func (b BalDetKDecomp) FindGHD(K int) Decomp {
	return b.findDecompBalSep(K, b.Depth, b.Graph, []Special{})
}
