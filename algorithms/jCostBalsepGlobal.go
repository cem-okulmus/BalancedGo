package algorithms

import (
	"container/heap"
	"fmt"
	"log"
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type JCostBalSepGlobal struct {
	Graph     Graph
	BalFactor int
	JCosts    EdgesCostMap
}

func (g JCostBalSepGlobal) findDecomp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	// //Early termination
	// if len(H.Edges) <= K && len(Sp) == 1 {
	//  return earlyTermination(H, Sp[0])
	// }

	//find a balanced separator
	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	// log.Printf("Starting Search: Edges: %v K: %v\n", edges.Len(), K)

	gen := GetCombin(edges.Len(), K)

OUTER:
	for gen.HasNext() {
		balsep := GetSubset(edges, gen.Combination)

		// log.Printf("Testing: %v\n", Graph{Edges: balsep})
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

		// log.Printf("Comps of Sep: %v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		var subtrees []Decomp
		for i := range comps {
			decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
					append(compsSp[i], SepSpecial))
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			// log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject

}

func (g JCostBalSepGlobal) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {
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
	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	//  go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := GetCombin(len(g.Graph.Edges), K)

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp

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
		//for _, g := range generators {
		//	fmt.Print(*g, " ")
		//}
		//fmt.Println()
		//fmt.Println("edges=", edges)
		//fmt.Println("found=", found)
		seps = append(seps, make([]int, len(found)))
		copy(seps[i], found)
		var found []int
		parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)
		lenFound = len(found)
	}
	fmt.Println()

	// populate heap
	fmt.Println(edges)
	jh := make(JoinHeap, len(seps))
	for i, fnd := range seps {
		fmt.Println("f=", fnd)
		s := make([]int, len(fnd))
		for i, f := range fnd {
			s[i] = edges.Slice()[f].Name
		}
		fmt.Println("s=", s)
		cost := g.JCosts.Cost(s)
		fmt.Println("cost=", cost)
		jh[i] = &Separator{
			Found:    fnd,
			EdgeComb: s,
			Cost:     cost,
		}
	}
	fmt.Println()
	heap.Init(&jh)
	/*for jh.Len() > 0 {
		s := heap.Pop(&jh).(*Separator)
		fmt.Println(s.EdgeComb, s.Cost)
		balsep = GetSubset(edges, s.Found)
		comps, _, _, _ := H.GetComponents(balsep, Sp)
		fmt.Println("comps=", comps)
		fmt.Println()
	}
	return Decomp{}
	*/
	//find a balanced separator
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

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

		// log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
				ch <- g.findDecompParallelFull(K, comps[i], append(compsSp[i], SepSpecial))
			}(K, i, comps, compsSp, SepSpecial)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				// if hinge {
				//  if isHinge(balsep, decomp.Graph) {
				//      return Decomp{}
				//  }
				// }
				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
				//  append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//log.Printf("\n\nCurrent SubGraph: %v\n", H)
				//log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}
			i = i
			// log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	return rerooting(H, balsep, subtrees)
}

func (g JCostBalSepGlobal) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

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
	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	// var numProc = runtime.GOMAXPROCS(-1)
	// var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int, 100)
	// for i := 0; i < numProc; i++ {
	//  go g.workerSimple(H, Sp, result, input, &wg)
	// }
	// generator := GetCombin(len(g.Graph.Edges), K)

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp
	// done := make(chan struct{})

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
		// close(done) // signal to workers to stop

		// log.Printf("Balanced Sep chosen: %+v\n", balsep)

		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

		// log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		for i := range comps {
			decomp := g.findDecompParallelSearch(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
				//    append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				// log.Printf("\n\nCurrent SubGraph: %v\n", H)
				// log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			// log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	return rerooting(H, balsep, subtrees)
}

func (g JCostBalSepGlobal) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

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
	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := GetCombin(edges.Len(), K)
OUTER:
	for gen.HasNext() {
		balsep := GetSubset(edges, gen.Combination)
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
			continue
		}

		// log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

		// log.Printf("Comps of Sep: %v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		var subtrees []Decomp

		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
				ch <- g.findDecompParallelComp(K, comps[i], append(compsSp[i], SepSpecial))
			}(K, i, comps, compsSp, SepSpecial)
		}

		for i := 0; i < len(comps); i++ {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
				//    append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//adapt search space for next iteration
				// log.Printf("\n\nCurrent SubGraph: %v\n", H)
				// log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			// log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)

	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

func (g JCostBalSepGlobal) FindGHD(K int) Decomp {
	return g.findDecomp(K, g.Graph, []Special{})
}

func (g JCostBalSepGlobal) FindGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g.Graph, []Special{})
}

func (g JCostBalSepGlobal) FindGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g.Graph, []Special{})
}

func (g JCostBalSepGlobal) FindGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g.Graph, []Special{})
}

func (g JCostBalSepGlobal) FindDecomp(K int) Decomp {
	return g.findDecompParallelFull(K, g.Graph, []Special{})
}

func (b JCostBalSepGlobal) FindDecompGraph(G Graph, K int) Decomp {
	return b.findDecompParallelFull(K, G, []Special{})
}

func (g JCostBalSepGlobal) Name() string {
	return "BalSep Global + Join Optimization"
}
