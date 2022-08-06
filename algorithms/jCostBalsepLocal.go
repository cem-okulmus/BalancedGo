package algorithms

import (
	"container/heap"
	"reflect"
	"runtime"

	"github.com/cem-okulmus/BalancedGo/lib"
	"github.com/cem-okulmus/disjoint"
)

// BalSepLocal implements the local Balanced Separator algorithm for computing GHDs.
// This will look for subedges locally, i.e. create them for each subgraph as needed.
type JCostBalSepLocal struct {
	K         int
	Graph     lib.Graph
	BalFactor int
	Generator lib.SearchGenerator
	JCosts    lib.EdgesCostMap
}

// SetGenerator defines the type of Search to use
func (b *JCostBalSepLocal) SetGenerator(Gen lib.SearchGenerator) {
	b.Generator = Gen
}

// SetWidth sets the current width parameter of the algorithm
func (b *JCostBalSepLocal) SetWidth(K int) {
	b.K = K
}

func (b JCostBalSepLocal) findGHD(K int) lib.Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecomp finds a decomp
func (b JCostBalSepLocal) FindDecomp() lib.Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (b JCostBalSepLocal) FindDecompGraph(G lib.Graph) lib.Decomp {
	return b.findDecomp(G)
}

// Name returns the name of the algorithm
func (b JCostBalSepLocal) Name() string {
	return "BalSep Local + Join Optimization"
}

func baseCaseSmartCosts(g lib.Graph, H lib.Graph, jc lib.EdgesCostMap) lib.Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output lib.Decomp

	var cost float64 = 0
	if H.Edges.Len() > 0 {
		s := make([]int, H.Edges.Len())
		for i, e := range H.Edges.Slice() {
			s[i] = e.Name
		}
		cost = jc.Cost(s)
	}

	if H.Edges.Len() <= 2 && len(H.Special) == 0 {
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: H.Vertices(), Cover: H.Edges, Cost: cost}}
	} else if H.Edges.Len() == 1 && len(H.Special) == 1 {
		sp1 := H.Special[0]
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: H.Edges.Vertices(), Cover: H.Edges, Cost: cost,
				Children: []lib.Node{lib.Node{Bag: sp1.Vertices(), Cover: sp1}}}}
	} else {
		return baseCase(g, H)
	}
	return output
}

func earlyTerminationCosts(H lib.Graph, jc lib.EdgesCostMap) lib.Decomp {
	//We assume that H as less than K edges, and only one special edge
	var cost float64 = 0
	if H.Edges.Len() > 0 {
		s := make([]int, H.Edges.Len())
		for i, e := range H.Edges.Slice() {
			s[i] = e.Name
		}
		cost = jc.Cost(s)
	}

	return lib.Decomp{Graph: H,
		Root: lib.Node{Bag: H.Edges.Vertices(), Cover: H.Edges, Cost: cost,
			Children: []lib.Node{lib.Node{Bag: H.Special[0].Vertices(), Cover: H.Special[0]}}}}
}

func rerootingCosts(H lib.Graph, balsep lib.Edges, subtrees []lib.Decomp, cost float64) lib.Decomp {
	//Create a new GHD for H
	rerootNode := lib.Node{Bag: balsep.Vertices(), Cover: balsep}
	output := lib.Node{Bag: balsep.Vertices(), Cover: balsep, Cost: cost}

	for _, s := range subtrees {
		// fmt.Println("H ", H, "balsep ", balsep, "comp ", s.Graph)
		s.Root = s.Root.Reroot(rerootNode)
		// log.Printf("Rerooted Decomp: %v\n", s)
		output.Children = append(output.Children, s.Root.Children...)
	}
	// log.Println("H: ", H, "output: ", output)
	return lib.Decomp{Graph: H, Root: output}
}

func orderSeparators(b JCostBalSepLocal, edges lib.Edges, ps lib.Search, pred lib.Predicate) []*lib.Separator {
	var seps [][]int
	var found []int
	ps.FindNext(pred) // initial search
	for ; !ps.SearchEnded(); ps.FindNext(pred) {
		found = ps.GetResult()

		newSep := make([]int, len(found))
		copy(newSep, found)
		seps = append(seps, newSep)
	}

	//fmt.Println("Checking seps:")
	//for i, v := range seps {
	//	fmt.Println(i, v)
	//}
	//fmt.Println()

	// populate heap
	jh := make(lib.JoinHeap, len(seps))
	for i, fnd := range seps {
		s := make([]int, len(fnd))
		for i, f := range fnd {
			s[i] = edges.Slice()[f].Name
		}
		cost := b.JCosts.Cost(s)
		jh[i] = &lib.Separator{
			Found:    fnd,
			EdgeComb: s,
			Cost:     cost,
		}
	}
	heap.Init(&jh)

	//fmt.Println("Checking heap:")
	//for i, sf := range jh {
	//	fmt.Println(i, sf.Found)
	//}
	//fmt.Println()

	var res []*lib.Separator
	for jh.Len() > 0 {
		sep := heap.Pop(&jh).(*lib.Separator)
		res = append(res, sep)
		//fmt.Println("currSep=", sep.Found)
		//fmt.Println("currEdgeComb=", sep.EdgeComb)
		//fmt.Println("currCost=", sep.Cost)
		//fmt.Println()
	}

	return res
}

func (b JCostBalSepLocal) findDecomp(H lib.Graph) lib.Decomp {
	// log.Printf("\n\nCurrent SubGraph: %v\n", H)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmartCosts(b.Graph, H, b.JCosts)
	}

	//Early termination
	if H.Edges.Len() <= b.K && len(H.Special) == 1 {
		return earlyTerminationCosts(H, b.JCosts)
	}
	var balsep lib.Edges

	edges := lib.CutEdges(b.Graph.Edges, append(H.Vertices()))
	generators := lib.SplitCombin(edges.Len(), b.K, runtime.GOMAXPROCS(-1), false)
	parallelSearch := b.Generator.GetSearch(&H, &edges, b.BalFactor, generators)
	pred := lib.BalancedCheck{}
	var Vertices = make(map[int]*disjoint.Element)
	// parallelSearch.FindNext(pred) // initial Search

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	separators := orderSeparators(b, edges, parallelSearch, pred)

	for _, sep := range separators {
		//for ; !parallelSearch.SearchEnded(); parallelSearch.FindNext(pred) {

		balsep = lib.GetSubset(edges, sep.Found)

		var sepSub *lib.SepSub
		// balsepOrig := balsep
		// log.Printf("Balanced Sep chosen: %v for H %v \n", balsep, H)

		exhaustedSubedges := false

	INNER:
		for !exhaustedSubedges {
			comps, _, _ := H.GetComponents(balsep, Vertices)

			// log.Printf("Comps of Sep: %v for H %v \n", comps, H)

			SepSpecial := lib.NewEdges(balsep.Slice())

			ch := make(chan lib.Decomp)
			var subtrees []lib.Decomp

			for i := range comps {
				go func(i int, comps []lib.Graph, SepSpecial lib.Edges) {
					comps[i].Special = append(comps[i].Special, SepSpecial)
					ch <- b.findDecomp(comps[i])
				}(i, comps, SepSpecial)
			}

			for i := 0; i < len(comps); i++ {
				decomp := <-ch
				if reflect.DeepEqual(decomp, lib.Decomp{}) {
					subtrees = []lib.Decomp{}
					if sepSub == nil {
						sepSub = lib.GetSepSub(b.Graph.Edges, balsep, b.K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							if len(balsep.Vertices()) == 0 {
								continue thisLoop
							}
							_, ok := cache[lib.IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if pred.Check(&H, &balsep, b.BalFactor, Vertices) {
								cache[lib.IntHash(balsep.Vertices())] = lib.Empty
								nextBalsepFound = true
							}
						} else {
							// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							exhaustedSubedges = true
							continue INNER
						}
					}
					// log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()), " of ",
					// 	balsepOrig, " , ", Sp)
					continue INNER
				}

				// log.Printf("Produced Decomp: %+v\n", decomp)
				subtrees = append(subtrees, decomp)
			}

			return rerootingCosts(H, balsep, subtrees, sep.Cost)
		}
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return lib.Decomp{} // empty Decomp signifying reject
}
