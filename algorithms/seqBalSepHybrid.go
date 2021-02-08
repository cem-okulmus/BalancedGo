package algorithms

import (
	"reflect"
	"strconv"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

// SeqBalSepHybrid is a purely sequential version of BalSepHybrid
type SeqBalSepHybrid struct {
	K         int
	Graph     Graph
	BalFactor int
	Depth     int // how many rounds of balSep are used
}

// SetWidth sets the current width parameter of the algorithm
func (s *SeqBalSepHybrid) SetWidth(K int) {
	s.K = K
}

func (s SeqBalSepHybrid) findGHD(currentGraph Graph) Decomp {
	return s.findDecomp(s.Depth, currentGraph)
}

// FindDecomp finds a decomp
func (s SeqBalSepHybrid) FindDecomp() Decomp {
	return s.findGHD(s.Graph)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (s SeqBalSepHybrid) FindDecompGraph(G Graph) Decomp {
	return s.findGHD(G)
}

// Name returns the name of the algorithm
func (s SeqBalSepHybrid) Name() string {
	return "BalSep / DetK - Hybrid with Depth " + strconv.Itoa(s.Depth+1)
}

func (s SeqBalSepHybrid) findDecomp(currentDepth int, H Graph) Decomp {
	// log.Println("Current Depth: ", (b.Depth - currentDepth))
	// log.Printf("Current SubGraph: %+v\n", H)
	// log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmart(s.Graph, H)
	}

	//Early termination
	if H.Edges.Len() <= s.K && len(H.Special) == 1 {
		return earlyTermination(H)
	}

	var balsep Edges

	//find a balanced separator
	edges := CutEdges(s.Graph.Edges, append(H.Vertices()))
	generators := SplitCombin(edges.Len(), s.K, 1, true) // create just one goroutine, making this sequential
	parallelSearch := Search{H: &H, Edges: &edges, BalFactor: s.BalFactor, Generators: generators}
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

			var subtrees []Decomp
			var outDecomps []Decomp

			//var outDecomp []Decomp
			for i := range comps {
				var out Decomp

				if currentDepth > 0 {
					out = func(i int, comps []Graph, SepSpecial Edges) Decomp {
						comps[i].Special = append(comps[i].Special, SepSpecial)
						return s.findDecomp(decrease(currentDepth), comps[i])
					}(i, comps, SepSpecial)
				} else {
					out = func(i int, comps []Graph, SepSpecial Edges) Decomp {

						// Base case handling
						comps[i].Special = append(comps[i].Special, SepSpecial)

						//stop if there are at most two special edges left
						if comps[i].Len() <= 2 {
							return baseCaseSmart(s.Graph, comps[i])
							//outDecomp = append(outDecomp, baseCaseSmart(b.Graph, comps[i], Sp))

						}

						//Early termination
						if comps[i].Edges.Len() <= s.K && len(comps[i].Special) == 1 {
							return earlyTermination(comps[i])
							//outDecomp = append(outDecomp, earlyTermination(comps[i], Sp[0]))

						}

						det := DetKDecomp{K: s.K, Graph: s.Graph, BalFactor: s.BalFactor, SubEdge: true}

						// edgesFromSpecial := EdgesSpecial(Sp)
						// comps[i].Edges.Append(edgesFromSpecial...)

						// det.cache = make(map[uint64]*CompCache)
						det.cache.Init()
						result := det.findDecomp(comps[i], balsep.Vertices())
						if !reflect.DeepEqual(result, Decomp{}) && currentDepth == 0 {
							result.SkipRerooting = true
						}
						return result
					}(i, comps, SepSpecial)
				}

				outDecomps = append(outDecomps, out)

			}

			for i := range outDecomps {
				decomp := outDecomps[i]
				if reflect.DeepEqual(decomp, Decomp{}) {
					// log.Printf("balDet REJECTING %v: couldn't decompose a component of H %v \n",
					//        Graph{Edges: balsep}, H)
					// log.Println("\n\nCurrent Depth: ", (b.Depth - currentDepth))
					// log.Printf("Current SubGraph: %+v\n", H)
					// log.Printf("Current Special Edges: %+v\n\n", Sp)

					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(s.Graph.Edges, balsep, s.K)
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

							if pred.Check(&H, &balsep, s.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							exhaustedSubedges = true
							continue INNER
						}
					}
					//      log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()),
					//         " of ", balsepOrig, " , ", Sp)
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
