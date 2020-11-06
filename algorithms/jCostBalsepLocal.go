package algorithms

import (
	"container/heap"
	"log"
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type JCostBalSepLocal struct {
	Graph     Graph
	BalFactor int
	JCosts    EdgesCostMap
}

func jCostSearchSubEdge(g *JCostBalSepLocal, H *Graph, Sp []Special, balsepOrig Edges, sepSub *SepSub, K int) Edges {

	balsep := balsepOrig

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Special Edges: %v\n\n", Sp)
	if sepSub == nil {
		balsep = CutEdges(balsep, H.Vertices())
		sepSub = GetSepSub(g.Graph.Edges, balsep, K)
	}
	nextBalsepFound := false

	for !nextBalsepFound {
		if sepSub.HasNext() {
			balsep = sepSub.GetCurrent()
			// log.Printf("Testing SSSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
			//        Graph{Edges: balsepOrig}, Sp)
			if H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
				nextBalsepFound = true
			}
		} else {
			return Edges{}
		}
	}
	log.Println("Sub Sep chosen: ", Graph{Edges: balsep})
	return balsep
}

func (g JCostBalSepLocal) findDecomp(K int, H Graph, Sp []Special) Decomp {

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	//find a balanced separator
	edges := CutEdges(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := GetCombinUnextend(edges.Len(), K)
OUTER:
	for gen.HasNext() {
		balsep := GetSubset(edges, gen.Combination)
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
			continue
		}
		var sepSub *SepSub

		// log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})
		// balsepOrig := balsep
	INNER:
		for {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			// log.Printf("Comps of Sep: %v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			var subtrees []Decomp

			for i := range comps {
				decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
				if reflect.DeepEqual(decomp, Decomp{}) {

					if sepSub == nil {
						sepSub = GetSepSub(g.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false

					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
							//        Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							//  log.Println(s.Combination)
							// }
							if H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
								nextBalsepFound = true
							}
						} else {
							// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: balsep},
					//        Graph{Edges: balsepOrig}, Sp)
					continue INNER

				}

				// log.Printf("Produced Decomp: %v\n", decomp)

				subtrees = append(subtrees, decomp)
			}

			return rerooting(H, balsep, subtrees)

		}

	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject

}

func (g JCostBalSepLocal) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}
	var balsep Edges

	//find a balanced separator
	var decomposed = false
	edges := CutEdges(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), true)
	var subtrees []Decomp

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	// collect separators
	var seps [][]int
	var found []int
	parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)
	if len(found) == 0 { // meaning that the search above never found anything
		log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
		return Decomp{}
	}
	lenFound := len(found)
	for i := 0; lenFound != 0; i++ {
		seps = append(seps, make([]int, len(found)))
		copy(seps[i], found)
		var found []int
		parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)
		lenFound = len(found)
	}

	// populate heap
	jh := make(JoinHeap, len(seps))
	for i, fnd := range seps {
		s := make([]int, len(fnd))
		for i, f := range fnd {
			s[i] = edges.Slice()[f].Name
		}
		cost := g.JCosts.Cost(s)
		jh[i] = &Separator{
			Found:    fnd,
			EdgeComb: s,
			Cost:     cost,
		}
	}
	heap.Init(&jh)

OUTER:
	for !decomposed {
		if jh.Len() > 0 {
			found = heap.Pop(&jh).(*Separator).Found
		} else { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)
		var sepSub *SepSub
		// balsepOrig := balsep
		// log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

	INNER:
		for !decomposed {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			// log.Printf("Comps of Sep: %v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			ch := make(chan Decomp)
			for i := range comps {
				go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {

					ch <- g.findDecompParallelFull(K, comps[i], append(compsSp[i], SepSpecial))

				}(K, i, comps, compsSp, SepSpecial)
			}

			for i := 0; i < len(comps); i++ {
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(g.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
							//        Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							//  log.Println(s.Combination)
							// }
							if len(balsep.Vertices()) == 0 {
								continue thisLoop
							}
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					// log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()), " of ",
					//        balsepOrig, " , ", Sp)
					continue INNER
				}

				// log.Printf("Produced Decomp: %+v\n", decomp)

				subtrees = append(subtrees, decomp)
			}

			decomposed = true
		}
	}

	return rerooting(H, balsep, subtrees)
}

func (g JCostBalSepLocal) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

	// log.Printf("Current SubGraph: %+v\n", H)
	// log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}
	var balsep Edges

	var decomposed = false
	edges := CutEdges(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	//  go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := GetCombin(len(g.Graph.Edges), K)

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), true)

	var subtrees []Decomp

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)

		if len(found) == 0 { // meaning that the search above never found anything
			// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)
		var sepSub *SepSub
		balsepOrig := balsep
		// log.Printf("Balanced Sep chosen: %+v\n", balsep)
	INNER:
		for !decomposed {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			for i := range comps {

				decomp := g.findDecompParallelSearch(K, comps[i], append(compsSp[i], SepSpecial))
				if reflect.DeepEqual(decomp, Decomp{}) {
					// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep},
					//        comps[i], append(compsSp[i], SepSpecial))
					subtrees = []Decomp{}
					// log.Printf("\n\nCurrent SubGraph: %v\n", H)
					// log.Printf("Current Special Edges: %v\n\n", Sp)
					if sepSub == nil {
						sepSub = GetSepSub(g.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false

					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
								Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							//  log.Println(s.Combination)
							// }
							if H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
								nextBalsepFound = true
							}
						} else {
							// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: balsep},
					//        Graph{Edges: balsepOrig}, Sp)
					continue INNER
				}

				// log.Printf("Produced Decomp: %v\n", decomp)

				subtrees = append(subtrees, decomp)
			}
			decomposed = true
		}

	}

	return rerooting(H, balsep, subtrees)
}

func (g JCostBalSepLocal) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	//find a balanced separator
	edges := CutEdges(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := GetCombin(edges.Len(), K)
OUTER:
	for gen.HasNext() {
		balsep := GetSubset(edges, gen.Combination)
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
			continue
		}
		var sepSub *SepSub

		// log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})
		//balsepOrig := balsep
	INNER:
		for {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			// log.Printf("Comps of Sep: %v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			var subtrees []Decomp

			ch := make(chan Decomp)
			quit := make(chan struct{})
			for i := range comps {
				go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
					select {
					case <-quit:
						return
					case ch <- g.findDecompParallelComp(K, comps[i], append(compsSp[i], SepSpecial)):
					}

				}(K, i, comps, compsSp, SepSpecial)
			}

			for i := 0; i < len(comps); i++ {
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {

					close(quit)
					// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep},
					//        comps[i], append(compsSp[i], SepSpecial))

					subBalSep := jCostSearchSubEdge(&g, &H, Sp, balsep, sepSub, K)
					if subBalSep.Len() == 0 {
						continue OUTER
					}
					balsep = subBalSep
					continue INNER
				}

				// log.Printf("Produced Decomp: %+v\n", decomp)

				subtrees = append(subtrees, decomp)
			}

			return rerooting(H, balsep, subtrees)
		}
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

func (g JCostBalSepLocal) FindGHD(K int) Decomp {
	return g.findDecomp(K, g.Graph, []Special{})
}

func (g JCostBalSepLocal) FindGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g.Graph, []Special{})
}

func (g JCostBalSepLocal) FindGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g.Graph, []Special{})
}

func (g JCostBalSepLocal) FindGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g.Graph, []Special{})
}

func (g JCostBalSepLocal) FindDecomp(K int) Decomp {
	return g.findDecompParallelFull(K, g.Graph, []Special{})
}

func (g JCostBalSepLocal) FindDecompGraph(G Graph, K int) Decomp {
	return g.findDecompParallelFull(K, G, []Special{})
}

func (g JCostBalSepLocal) Name() string {
	return "BalSep Local + Join Optimization"
}
