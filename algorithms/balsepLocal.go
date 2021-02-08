package algorithms

import (
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

// BalSepLocal implements the local Balanced Separator algorithm for computing GHDs.
// This will look for subedges locally, i.e. create them for each subgraph as needed.
type BalSepLocal struct {
	K         int
	Graph     Graph
	BalFactor int
}

// SetWidth sets the current width parameter of the algorithm
func (b *BalSepLocal) SetWidth(K int) {
	b.K = K
}

func (b BalSepLocal) findGHD(K int) Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecomp finds a decomp
func (b BalSepLocal) FindDecomp() Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (b BalSepLocal) FindDecompGraph(G Graph) Decomp {
	return b.findDecomp(G)
}

// Name returns the name of the algorithm
func (b BalSepLocal) Name() string {
	return "BalSep Local"
}

func searchSubEdge(g *BalSepLocal, H *Graph, balsepOrig Edges, sepSub *SepSub) Edges {
	balsep := balsepOrig

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Special Edges: %v\n\n", Sp)
	if sepSub == nil {
		balsep = CutEdges(balsep, H.Vertices())
		sepSub = GetSepSub(g.Graph.Edges, balsep, g.K)
	}
	nextBalsepFound := false
	pred := BalancedCheck{}

	for !nextBalsepFound {
		if sepSub.HasNext() {
			balsep = sepSub.GetCurrent()
			// log.Printf("Testing SSSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
			//        Graph{Edges: balsepOrig}, Sp)
			if pred.Check(H, &balsep, g.BalFactor) {
				nextBalsepFound = true
			}
		} else {
			return NewEdges([]Edge{})
		}
	}
	// log.Println("Sub Sep chosen: ", balsep)
	return balsep
}

func (b BalSepLocal) findDecomp(H Graph) Decomp {
	// log.Printf("\n\nCurrent SubGraph: %v\n", H)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmart(b.Graph, H)
	}

	//Early termination
	if H.Edges.Len() <= b.K && len(H.Special) == 1 {
		return earlyTermination(H)
	}
	var balsep Edges

	edges := CutEdges(b.Graph.Edges, append(H.Vertices()))
	generators := SplitCombin(edges.Len(), b.K, runtime.GOMAXPROCS(-1), true)
	parallelSearch := Search{H: &H, Edges: &edges, BalFactor: b.BalFactor, Generators: generators}
	pred := BalancedCheck{}
	parallelSearch.FindNext(pred) // initial Search

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {

		balsep = GetSubset(edges, parallelSearch.Result)

		var sepSub *SepSub
		// balsepOrig := balsep
		// log.Printf("Balanced Sep chosen: %v for H %v \n", balsep, H)

		exhaustedSubedges := false

	INNER:
		for !exhaustedSubedges {
			comps, _, _ := H.GetComponents(balsep)

			// log.Printf("Comps of Sep: %v for H %v \n", comps, H)

			SepSpecial := NewEdges(balsep.Slice())

			ch := make(chan Decomp)
			var subtrees []Decomp

			for i := range comps {
				go func(i int, comps []Graph, SepSpecial Edges) {
					comps[i].Special = append(comps[i].Special, SepSpecial)
					ch <- b.findDecomp(comps[i])
				}(i, comps, SepSpecial)
			}

			for i := 0; i < len(comps); i++ {
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(b.Graph.Edges, balsep, b.K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							if len(balsep.Vertices()) == 0 {
								continue thisLoop
							}
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if pred.Check(&H, &balsep, b.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
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

			return rerooting(H, balsep, subtrees)
		}
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifying reject
}
