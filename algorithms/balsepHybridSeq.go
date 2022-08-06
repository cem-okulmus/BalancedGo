package algorithms

import (
	"reflect"
	"strconv"

	"github.com/cem-okulmus/BalancedGo/lib"
	"github.com/cem-okulmus/disjoint"
)

// BalSepHybridSeq is a purely sequential version of BalSepHybrid
type BalSepHybridSeq struct {
	K         int
	Graph     lib.Graph
	BalFactor int
	Depth     int // how many rounds of balSep are used
	Generator lib.SearchGenerator
}

// SetGenerator defines the type of Search to use
func (b *BalSepHybridSeq) SetGenerator(Gen lib.SearchGenerator) {
	b.Generator = Gen
}

// SetWidth sets the current width parameter of the algorithm
func (s *BalSepHybridSeq) SetWidth(K int) {
	s.K = K
}

func (s BalSepHybridSeq) findGHD(currentGraph lib.Graph) lib.Decomp {
	return s.findDecomp(s.Depth, currentGraph)
}

// FindDecomp finds a decomp
func (s BalSepHybridSeq) FindDecomp() lib.Decomp {
	return s.findGHD(s.Graph)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (s BalSepHybridSeq) FindDecompGraph(G lib.Graph) lib.Decomp {
	return s.findGHD(G)
}

// Name returns the name of the algorithm
func (s BalSepHybridSeq) Name() string {
	return "BalSep / DetK - Hybrid with Depth " + strconv.Itoa(s.Depth+1)
}

func (s BalSepHybridSeq) findDecomp(currentDepth int, H lib.Graph) lib.Decomp {
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

	var balsep lib.Edges

	//find a balanced separator
	edges := lib.CutEdges(s.Graph.Edges, append(H.Vertices()))
	generators := lib.SplitCombin(edges.Len(), s.K, 1, true) // create just one goroutine, making this sequential
	parallelSearch := s.Generator.GetSearch(&H, &edges, s.BalFactor, generators)
	pred := lib.BalancedCheck{}
	var Vertices = make(map[int]*disjoint.Element)
	parallelSearch.FindNext(pred) // initial Search

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	// OUTER:
	for ; !parallelSearch.SearchEnded(); parallelSearch.FindNext(pred) {

		balsep = lib.GetSubset(edges, parallelSearch.GetResult())

		//  balsepOrig := balsep
		var sepSub *lib.SepSub

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})
		exhaustedSubedges := false

	INNER:
		for !exhaustedSubedges {
			comps, _, _ := H.GetComponents(balsep, Vertices)

			// log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := lib.NewEdges(balsep.Slice())

			var subtrees []lib.Decomp
			var outDecomps []lib.Decomp

			//var outDecomp []Decomp
			for i := range comps {
				var out lib.Decomp

				if currentDepth > 0 {
					out = func(i int, comps []lib.Graph, SepSpecial lib.Edges) lib.Decomp {
						comps[i].Special = append(comps[i].Special, SepSpecial)
						return s.findDecomp(decrease(currentDepth), comps[i])
					}(i, comps, SepSpecial)
				} else {
					out = func(i int, comps []lib.Graph, SepSpecial lib.Edges) lib.Decomp {

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
						result := det.findDecomp(comps[i], balsep.Vertices(), 0)
						if !reflect.DeepEqual(result, lib.Decomp{}) && currentDepth == 0 {
							result.SkipRerooting = true
						}
						return result
					}(i, comps, SepSpecial)
				}

				outDecomps = append(outDecomps, out)

			}

			for i := range outDecomps {
				decomp := outDecomps[i]
				if reflect.DeepEqual(decomp, lib.Decomp{}) {
					// log.Printf("balDet REJECTING %v: couldn't decompose a component of H %v \n",
					//        Graph{Edges: balsep}, H)
					// log.Println("\n\nCurrent Depth: ", (b.Depth - currentDepth))
					// log.Printf("Current SubGraph: %+v\n", H)
					// log.Printf("Current Special Edges: %+v\n\n", Sp)

					subtrees = []lib.Decomp{}
					if sepSub == nil {
						sepSub = lib.GetSepSub(s.Graph.Edges, balsep, s.K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							_, ok := cache[lib.IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}

							if pred.Check(&H, &balsep, s.BalFactor, Vertices) {
								cache[lib.IntHash(balsep.Vertices())] = lib.Empty
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

			output := lib.Node{Bag: balsep.Vertices(), Cover: balsep}

			for _, s := range subtrees {
				if currentDepth == 0 && s.SkipRerooting {
					// log.Println("\nFrom detK on", decomp.Graph, ":\n", decomp)
					// local := BalSepGlobal{Graph: b.Graph, BalFactor: b.BalFactor}
					// decomp_deux := local.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
					// fmt.Println("Output from Balsep: ", decomp_deux)
				} else {
					s.Root = s.Root.Reroot(lib.Node{Bag: balsep.Vertices(), Cover: balsep})
					s.Root = s.Root.Children[0]
					// log.Printf("Produced Decomp (with balsep %v): %+v\n", balsep, decomp)
				}

				output.Children = append(output.Children, s.Root)
			}

			return lib.Decomp{Graph: H, Root: output}
		}
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return lib.Decomp{} // empty Decomp signifying reject
}
