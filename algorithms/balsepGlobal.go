package algorithms

import (
	"log"
	"reflect"
	"runtime"
	"sync"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type BalSepGlobal struct {
	Graph     Graph
	BalFactor int
}

func baseCaseSmart(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output Decomp

	if H.Edges.Len() <= 2 && len(Sp) == 0 {
		output = Decomp{Graph: H,
			Root: Node{Bag: H.Vertices(), Cover: H.Edges}}
	} else if H.Edges.Len() == 1 && len(Sp) == 1 {
		sp1 := Sp[0]
		output = Decomp{Graph: H,
			Root: Node{Bag: H.Vertices(), Cover: H.Edges,
				Children: []Node{Node{Bag: sp1.Vertices, Cover: sp1.Edges}}}}
	} else {
		return baseCase(g, H, Sp)
	}
	return output
}

func baseCase(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output Decomp
	switch len(Sp) {
	case 0:
		output = Decomp{Graph: g} // use g here to avoid reject
	case 1:
		sp1 := Sp[0]
		output = Decomp{Graph: H,
			Root: Node{Bag: sp1.Vertices, Cover: sp1.Edges}}
	case 2:
		sp1 := Sp[0]
		sp2 := Sp[1]
		output = Decomp{Graph: H,
			Root: Node{Bag: sp1.Vertices, Cover: sp1.Edges,
				Children: []Node{Node{Bag: sp2.Vertices, Cover: sp2.Edges}}}}

	}
	return output
}

func earlyTermination(H Graph, sp Special) Decomp {
	//We assume that H as less than K edges, and only one special edge
	return Decomp{Graph: H,
		Root: Node{Bag: H.Vertices(), Cover: H.Edges,
			Children: []Node{Node{Bag: sp.Vertices, Cover: sp.Edges}}}}
}

func rerooting(H Graph, balsep Edges, subtrees []Decomp) Decomp {

	//Create a new GHD for H
	rerootNode := Node{Bag: balsep.Vertices(), Cover: balsep}
	output := Node{Bag: balsep.Vertices(), Cover: balsep}

	// log.Printf("Node to reRoot: %v\n", rerootNode)
	// log.Printf("My subtrees: \n")
	// for _, s := range subtrees {
	// 	log.Printf("%v \n", s)
	// }
	for _, s := range subtrees {
		// fmt.Println("H ", H, "balsep ", balsep, "comp ", s.Graph)
		s.Root = s.Root.Reroot(rerootNode)
		log.Printf("Rerooted Decomp: %v\n", s)
		output.Children = append(output.Children, s.Root.Children...)
	}
	log.Println("H: ", H, "output: ", output)
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

func (g BalSepGlobal) findDecomp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(g.Graph, H, Sp)
	}

	// //Early termination
	// if len(H.Edges) <= K && len(Sp) == 1 {
	// 	return earlyTermination(H, Sp[0])
	// }

	//find a balanced separator
	edges := FilterVerticesStrict(g.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	log.Printf("Starting Search: Edges: %v K: %v\n", edges.Len(), K)

	gen := GetCombin(edges.Len(), K)

OUTER:
	for gen.HasNext() {
		balsep := GetSubset(edges, gen.Combination)

		log.Printf("Testing: %v\n", Graph{Edges: balsep})
		gen.Confirm()
		if !H.CheckBalancedSep(balsep, Sp, g.BalFactor) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, compsSp, _ := H.GetComponents(balsep, Sp)

		log.Printf("Comps of Sep: %v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		var subtrees []Decomp
		for i := range comps {
			decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)
	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject

}

func parallelSearch(H Graph, Sp []Special, edges Edges, result *[]int, generators []*CombinationIterator, BalFactor int) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	var numProc = runtime.GOMAXPROCS(-1)

	var wg sync.WaitGroup
	wg.Add(numProc)
	finished := false
	// SEARCH:
	found := make(chan []int)
	wait := make(chan bool)
	//start workers
	for i := 0; i < numProc; i++ {
		go worker(i, H, Sp, edges, found, generators[i], &wg, &finished, BalFactor)
	}

	go func() {
		wg.Wait()
		wait <- true
	}()

	select {
	case *result = <-found:
		close(found) //to terminate other workers waiting on found
	case <-wait:
	}

}

func worker(workernum int, H Graph, Sp []Special, edges Edges, found chan []int, gen *CombinationIterator, wg *sync.WaitGroup, finished *bool, BalFactor int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d 'forced' to quit, reason: %v", workernum, r)
			return
		}
	}()
	defer wg.Done()

	for gen.HasNext() {
		if *finished {
			log.Printf("Worker %d told to quit", workernum)
			return
		}
		j := gen.Combination

		if H.CheckBalancedSep(GetSubset(edges, j), Sp, BalFactor) {
			found <- j
			log.Printf("Worker %d \" won \"", workernum)
			gen.Confirm()
			*finished = true
			return
		}
		gen.Confirm()
	}
}

func (g BalSepGlobal) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {
	log.Printf("Current SubGraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

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
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := GetCombin(len(g.Graph.Edges), K)

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators, g.BalFactor)

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
			go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
				ch <- g.findDecompParallelFull(K, comps[i], append(compsSp[i], SepSpecial))
			}(K, i, comps, compsSp, SepSpecial)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				// if hinge {
				// 	if isHinge(balsep, decomp.Graph) {
				// 		return Decomp{}
				// 	}
				// }
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//log.Printf("\n\nCurrent SubGraph: %v\n", H)
				//log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	return rerooting(H, balsep, subtrees)
}

func (g BalSepGlobal) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

	log.Printf("Current SubGraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

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
	// 	go g.workerSimple(H, Sp, result, input, &wg)
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
			log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)
		// close(done) // signal to workers to stop

		log.Printf("Balanced Sep chosen: %+v\n", balsep)

		comps, compsSp, _ := H.GetComponents(balsep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

		for i := range comps {
			decomp := g.findDecompParallelSearch(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	return rerooting(H, balsep, subtrees)
}

func (g BalSepGlobal) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

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

		log.Printf("Balanced Sep chosen: %v\n", Graph{Edges: balsep})

		comps, compsSp, _ := H.GetComponents(balsep, Sp)

		log.Printf("Comps of Sep: %v\n", comps)

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
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//adapt search space for next iteration
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)

	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

func (g BalSepGlobal) FindGHD(K int) Decomp {
	return g.findDecomp(K, g.Graph, []Special{})
}

func (g BalSepGlobal) FindGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g.Graph, []Special{})
}

func (g BalSepGlobal) FindGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g.Graph, []Special{})
}

func (g BalSepGlobal) FindGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g.Graph, []Special{})
}
