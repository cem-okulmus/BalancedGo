// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to
// DetKDecomp

package algorithms

import (
	"reflect"
	"runtime"
	"strconv"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

// BalSepHybrid implements a hybridised algorithm, using BalSep Local and DetKDecomp in tandem
type BalSepHybrid struct {
	K         int
	Graph     Graph
	BalFactor int
	Depth     int // how many rounds of balSep are used
}

// SetWidth sets the current width parameter of the algorithm
func (b *BalSepHybrid) SetWidth(K int) {
	b.K = K
}

func (b BalSepHybrid) findGHD(currentGraph Graph) Decomp {
	return b.findDecomp(b.Depth, currentGraph)
}

// FindDecomp finds a decomp
func (b BalSepHybrid) FindDecomp() Decomp {
	return b.findGHD(b.Graph)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (b BalSepHybrid) FindDecompGraph(G Graph) Decomp {
	return b.findGHD(G)
}

// Name returns the name of the algorithm
func (b BalSepHybrid) Name() string {
	return "BalSep / DetK - Hybrid with Depth " + strconv.Itoa(b.Depth+1)
}

func decrease(count int) int {
	output := count - 1
	if output < 1 {
		return 0
	}
	return output
}

func (b BalSepHybrid) findDecomp(currentDepth int, H Graph) Decomp {
	// log.Println("Current Depth: ", (b.Depth - currentDepth))
	// log.Printf("Current SubGraph: %+v\n", H)
	// log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmart(b.Graph, H)
	}

	//Early termination
	if H.Edges.Len() <= b.K && len(H.Special) == 1 {
		return earlyTermination(H)
	}

	var balsep Edges

	//find a balanced separator
	edges := CutEdges(b.Graph.Edges, append(H.Vertices()))
	generators := SplitCombin(edges.Len(), b.K, runtime.GOMAXPROCS(-1), true)
	parallelSearch := Search{H: &H, Edges: &edges, BalFactor: b.BalFactor, Generators: generators}
	pred := BalancedCheck{}
	parallelSearch.FindNext(pred) // initial Search

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	// OUTER:
	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {
		balsep = GetSubset(edges, parallelSearch.Result)

		//  balsepOrig := balsep
		var sepSub *SepSub

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})
		exhaustedSubedges := false

	INNER:
		for !exhaustedSubedges {
			comps, _, _ := H.GetComponents(balsep)

			// log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := NewEdges(balsep.Slice())

			ch := make(chan Decomp)
			var subtrees []Decomp

			//var outDecomp []Decomp
			for i := range comps {

				if currentDepth > 0 {
					go func(i int, comps []Graph, SepSpecial Edges) {
						comps[i].Special = append(comps[i].Special, SepSpecial)
						ch <- b.findDecomp(decrease(currentDepth), comps[i])
					}(i, comps, SepSpecial)
				} else {
					go func(i int, comps []Graph, SepSpecial Edges) {

						// Base case handling
						//stop if there are at most two special edges left
						if comps[i].Len() <= 2 {
							ch <- baseCaseSmart(b.Graph, comps[i])
							//outDecomp = append(outDecomp, baseCaseSmart(b.Graph, comps[i], Sp))
							return
						}

						//Early termination
						if comps[i].Edges.Len() <= b.K && len(comps[i].Special) == 1 {
							ch <- earlyTermination(comps[i])
							//outDecomp = append(outDecomp, earlyTermination(comps[i], Sp[0]))
							return
						}

						det := DetKDecomp{K: b.K, Graph: b.Graph, BalFactor: b.BalFactor, SubEdge: true}

						// edgesFromSpecial := EdgesSpecial(Sp)
						// comps[i].Edges.Append(edgesFromSpecial...)

						// det.cache = make(map[uint64]*CompCache)
						det.cache.Init()
						result := det.findDecomp(comps[i], balsep.Vertices())
						if !reflect.DeepEqual(result, Decomp{}) && currentDepth == 0 {
							result.SkipRerooting = true
						} else {
							// comps[i].Special = append(comps[i].Special, SepSpecial)
							// res2 := b.findDecompBalSep(1000, comps[i])
							// if !reflect.DeepEqual(res2, Decomp{}) {
							// 	fmt.Println("Result, ", res2)
							// 	fmt.Println("H: ", comps[i], "balsep ", balsep)
							// 	log.Panicln("Something is rotten in the state of this program")

							// }
						}
						ch <- result
						// outDecomp = append(outDecomp, result)
					}(i, comps, SepSpecial)
				}

			}

			for i := 0; i < len(comps); i++ {
				//  decomp := outDecomp[i]
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					// log.Printf("balDet REJECTING %v: couldn't decompose a component of H %v \n",
					//        Graph{Edges: balsep}, H)
					// log.Println("\n\nCurrent Depth: ", (b.Depth - currentDepth))
					// log.Printf("Current SubGraph: %+v\n", H)
					// log.Printf("Current Special Edges: %+v\n\n", Sp)

					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(b.Graph.Edges, balsep, b.K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}

							if pred.Check(&H, &balsep, b.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							exhaustedSubedges = true
							continue INNER
						}
					}
					continue INNER
				}

				subtrees = append(subtrees, decomp)
			}

			output := Node{Bag: balsep.Vertices(), Cover: balsep}

			for _, s := range subtrees {
				//TODO: Reroot only after all subtrees received
				if currentDepth == 0 && s.SkipRerooting {
					// log.Println("\nFrom detK on", decomp.Graph, ":\n", decomp)
					// local := BalSepGlobal{Graph: b.Graph, BalFactor: b.BalFactor}
					// decomp_deux := local.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
					// fmt.Println("Output from Balsep: ", decomp_deux)
				} else {
					s.Root = s.Root.Reroot(Node{Bag: balsep.Vertices(), Cover: balsep})
					s.Root = s.Root.Children[0]
					// log.Printf("Produced Decomp (with balsep %v): %+v\n", balsep, decomp)
				}

				output.Children = append(output.Children, s.Root)
			}

			return Decomp{Graph: H, Root: output}
		}
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifying reject
}
