package algorithms

import (
	"log"
	"reflect"
	"runtime"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type BalSepGlobal struct {
	K         int
	Graph     Graph
	BalFactor int
}

func (b *BalSepGlobal) SetWidth(K int) {
	b.K = K
}

func (g BalSepGlobal) FindGHD() Decomp {
	return g.findDecompParallelFull(g.Graph)
}

func (g BalSepGlobal) FindGHDParallelFull() Decomp {
	return g.findDecompParallelFull(g.Graph)
}

// func (g BalSepGlobal) FindGHDParallelSearch(K int) Decomp {
// 	return g.findDecompParallelSearch(K, g.Graph, []Special{})
// }

// func (g BalSepGlobal) FindGHDParallelComp(K int) Decomp {
// 	return g.findDecompParallelComp(K, g.Graph, []Special{})
// }

func (g BalSepGlobal) FindDecomp() Decomp {
	return g.findDecompParallelFull(g.Graph)
}

func (g BalSepGlobal) FindDecompGraph(G Graph) Decomp {
	return g.findDecompParallelFull(G)
}

func (g BalSepGlobal) Name() string {
	return "BalSep Global"
}

func baseCaseSmart(g Graph, H Graph) Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", H.Special.Len() )
	var output Decomp

	if H.Edges.Len() <= 2 && len(H.Special) == 0 {
		output = Decomp{Graph: H,
			Root: Node{Bag: H.Vertices(), Cover: H.Edges}}
	} else if H.Edges.Len() == 1 && len(H.Special) == 1 {
		sp1 := H.Special[0]
		output = Decomp{Graph: H,
			Root: Node{Bag: H.Vertices(), Cover: H.Edges,
				Children: []Node{Node{Bag: sp1.Vertices(), Cover: sp1}}}}
	} else {
		return baseCase(g, H)
	}
	return output
}

func baseCase(g Graph, H Graph) Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", H.Special.Len())
	var output Decomp
	switch len(H.Special) {
	case 0:
		output = Decomp{Graph: g} // use g here to avoid reject
	case 1:
		sp1 := H.Special[0]
		output = Decomp{Graph: H,
			Root: Node{Bag: sp1.Vertices(), Cover: sp1}}
	case 2:
		sp1 := H.Special[0]
		sp2 := H.Special[1]
		output = Decomp{Graph: H,
			Root: Node{Bag: sp1.Vertices(), Cover: sp1,
				Children: []Node{Node{Bag: sp2.Vertices(), Cover: sp2}}}}

	}
	return output
}

func earlyTermination(H Graph) Decomp {
	//We assume that H as less than K edges, and only one special edge
	return Decomp{Graph: H,
		Root: Node{Bag: H.Vertices(), Cover: H.Edges,
			Children: []Node{Node{Bag: H.Special[0].Vertices(), Cover: H.Special[0]}}}}
}

func rerooting(H Graph, balsep Edges, subtrees []Decomp) Decomp {

	//Create a new GHD for H
	rerootNode := Node{Bag: balsep.Vertices(), Cover: balsep}
	output := Node{Bag: balsep.Vertices(), Cover: balsep}

	// log.Printf("Node to reRoot: %v\n", rerootNode)
	// log.Printf("My subtrees: \n")
	// for _, s := range subtrees {
	//  log.Printf("%v \n", s)
	// }
	for _, s := range subtrees {
		// fmt.Println("H ", H, "balsep ", balsep, "comp ", s.Graph)
		s.Root = s.Root.Reroot(rerootNode)
		log.Printf("Rerooted Decomp: %v\n", s)
		output.Children = append(output.Children, s.Root.Children...)
	}
	// log.Println("H: ", H, "output: ", output)
	return Decomp{Graph: H, Root: output}
}

func isHinge(sep Edges, comp Graph) bool {
	inter := Inter(sep.Vertices(), comp.Vertices())

	for _, e := range sep.Slice() {
		if Subset(inter, e.Vertices) {
			return true
		}
	}

	return false
}

func (g BalSepGlobal) findDecompParallelFull(H Graph) Decomp {
	// log.Printf("Current SubGraph: %+v\n", H)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmart(g.Graph, H)
	}

	//Early termination
	if H.Edges.Len() <= g.K && len(H.Special) == 1 {
		return earlyTermination(H)
	}

	var balsep Edges

	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices()))

	generators := SplitCombin(edges.Len(), g.K, runtime.GOMAXPROCS(-1), false)

	parallelSearch := Search{H: &H, Edges: &edges, BalFactor: g.BalFactor, Generators: generators}

	pred := BalancedCheck{}

	parallelSearch.FindNext(pred) // initial Search

OUTER:
	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {

		balsep = GetSubset(edges, parallelSearch.Result)

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

		comps, _, _ := H.GetComponents(balsep)

		// log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := NewEdges(balsep.Slice())

		var subtrees []Decomp
		ch := make(chan Decomp)

		for i := range comps {
			go func(i int, comps []Graph, SepSpecial Edges) {
				comps[i].Special = append(comps[i].Special, SepSpecial)
				ch <- g.findDecompParallelFull(comps[i])
			}(i, comps, SepSpecial)
		}

		for i := 0; i < len(comps); i++ {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
				//  append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//log.Printf("\n\nCurrent SubGraph: %v\n", H)
				//log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}
			// log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

// func (g BalSepGlobal) findDecomp(H Graph) Decomp {

// 	log.Printf("\n\nCurrent SubGraph: %v\n", H.Edges)
// 	log.Printf("Current Special Edges: %v\n\n", H.Special)

// 	//stop if there are at most two special edges left
// 	if H.Len() <= 2 {
// 		return baseCaseSmart(g.Graph, H, Sp)
// 	}

// 	// //Early termination
// 	// if len(H.Edges) <= K && len(Sp) == 1 {
// 	//  return earlyTermination(H, Sp[0])
// 	// }

// 	//find a balanced separator
// 	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices()))

// 	// log.Printf("Starting Search: Edges: %v K: %v\n", edges.Len(), K)

// 	gen := GetCombin(edges.Len(), g.K)

// 	pred := BalancedCheck{}

// OUTER:
// 	for gen.HasNext() {
// 		balsep := GetSubset(edges, gen.Combination)

// 		// log.Printf("Testing: %v\n", Graph{Edges: balsep})
// 		gen.Confirm()

// 		if !pred.Check(&H, &balsep, g.BalFactor) {
// 			continue
// 		}

// 		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

// 		comps, _, _ := H.GetComponents(balsep)

// 		// log.Printf("Comps of Sep: %v\n", comps)

// 		SepSpecial := NewEdges(balsep.Slice())

// 		var subtrees []Decomp
// 		for i := range comps {
// 			comps[i].Special = append(comps[i].Special, SepSpecial)
// 			decomp := g.findDecomp(K, comps[i])
// 			if reflect.DeepEqual(decomp, Decomp{}) {
// 				log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{Edges: balsep}, comps[i])
// 				log.Printf("\n\nCurrent SubGraph: %v\n", H)
// 				continue OUTER
// 			}

// 			// log.Printf("Produced Decomp: %v\n", decomp)

// 			subtrees = append(subtrees, decomp)
// 		}

// 		return rerooting(H, balsep, subtrees)
// 	}

// 	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
// 	return Decomp{} // empty Decomp signifiyng reject

// }

// func (g BalSepGlobal) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

// 	// log.Printf("Current SubGraph: %+v\n", H)
// 	// log.Printf("Current Special Edges: %+v\n\n", Sp)

// 	//stop if there are at most two special edges left
// 	if H.Edges.Len()+len(Sp) <= 2 {
// 		return baseCaseSmart(g.Graph, H, Sp)
// 	}

// 	//Early termination
// 	if H.Edges.Len() <= K && len(Sp) == 1 {
// 		return earlyTermination(H, Sp[0])
// 	}

// 	var balsep Edges

// 	var decomposed = false
// 	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

// 	// var numProc = runtime.GOMAXPROCS(-1)
// 	// var wg sync.WaitGroup
// 	// wg.Add(numProc)
// 	// result := make(chan []int)
// 	// input := make(chan []int, 100)
// 	// for i := 0; i < numProc; i++ {
// 	//  go g.workerSimple(H, Sp, result, input, &wg)
// 	// }
// 	// generator := GetCombin(len(g.Graph.Edges), K)

// 	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), false)

// 	var subtrees []Decomp
// 	// done := make(chan struct{})

// 	//find a balanced separator
// OUTER:
// 	for !decomposed {
// 		var found []int

// 		//g.startSearchSimple(&found, &generator, result, input, &wg)
// 		parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)

// 		if len(found) == 0 { // meaning that the search above never found anything
// 			// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
// 			return Decomp{}
// 		}

// 		//wait until first worker finds a balanced sep
// 		balsep = GetSubset(edges, found)
// 		// close(done) // signal to workers to stop

// 		// log.Printf("Balanced Sep chosen: %+v\n", balsep)

// 		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

// 		// log.Printf("Comps of Sep: %+v\n", comps)

// 		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

// 		for i := range comps {
// 			decomp := g.findDecompParallelSearch(K, comps[i], append(compsSp[i], SepSpecial))
// 			if reflect.DeepEqual(decomp, Decomp{}) {
// 				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
// 				//    append(compsSp[i], SepSpecial))
// 				subtrees = []Decomp{}
// 				// log.Printf("\n\nCurrent SubGraph: %v\n", H)
// 				// log.Printf("Current Special Edges: %v\n\n", Sp)
// 				continue OUTER
// 			}

// 			// log.Printf("Produced Decomp: %v\n", decomp)

// 			subtrees = append(subtrees, decomp)
// 		}

// 		decomposed = true
// 	}

// 	return rerooting(H, balsep, subtrees)
// }

// func (g BalSepGlobal) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

// 	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
// 	// log.Printf("Current Special Edges: %v\n\n", Sp)

// 	//stop if there are at most two special edges left
// 	if H.Edges.Len()+len(Sp) <= 2 {
// 		return baseCaseSmart(g.Graph, H, Sp)
// 	}

// 	//Early termination
// 	if H.Edges.Len() <= K && len(Sp) == 1 {
// 		return earlyTermination(H, Sp[0])
// 	}

// 	//find a balanced separator
// 	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

// 	gen := GetCombin(edges.Len(), K)
// OUTER:
// 	for gen.HasNext() {
// 		balsep := GetSubset(edges, gen.Combination)
// 		gen.Confirm()
// 		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
// 			continue
// 		}

// 		// log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

// 		comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

// 		// log.Printf("Comps of Sep: %v\n", comps)

// 		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

// 		var subtrees []Decomp

// 		ch := make(chan Decomp)
// 		for i := range comps {
// 			go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
// 				ch <- g.findDecompParallelComp(K, comps[i], append(compsSp[i], SepSpecial))
// 			}(K, i, comps, compsSp, SepSpecial)
// 		}

// 		for i := 0; i < len(comps); i++ {
// 			decomp := <-ch
// 			if reflect.DeepEqual(decomp, Decomp{}) {
// 				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
// 				//    append(compsSp[i], SepSpecial))
// 				subtrees = []Decomp{}
// 				//adapt search space for next iteration
// 				// log.Printf("\n\nCurrent SubGraph: %v\n", H)
// 				// log.Printf("Current Special Edges: %v\n\n", Sp)
// 				continue OUTER
// 			}

// 			// log.Printf("Produced Decomp: %+v\n", decomp)

// 			subtrees = append(subtrees, decomp)
// 		}

// 		return rerooting(H, balsep, subtrees)

// 	}

// 	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
// 	return Decomp{} // empty Decomp signifiyng reject
// }
