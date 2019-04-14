package main

import (
	"log"
	"reflect"
	"runtime"
	"sync"
)

type GlobalSearch struct {
	graph Graph
}

func baseCaseSmart(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output Decomp

	if len(H.edges) == 1 {
		sp1 := Sp[0]
		output = Decomp{graph: H,
			root: Node{lambda: H.edges,
				children: []Node{Node{lambda: sp1.edges}}}}
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
		output = Decomp{graph: g} // use g here to avoid reject
	case 1:
		sp1 := Sp[0]
		output = Decomp{graph: H,
			root: Node{lambda: sp1.edges}}
	case 2:
		sp1 := Sp[0]
		sp2 := Sp[1]
		output = Decomp{graph: H,
			root: Node{lambda: sp1.edges,
				children: []Node{Node{lambda: sp2.edges}}}}

	}
	return output
}

func earlyTermination(H Graph, Sp Special) Decomp {
	//We assume that H as less than K edges, and only one special edge
	return Decomp{graph: H, root: Node{lambda: H.edges, children: []Node{Node{lambda: Sp.edges}}}}
}

func (g GlobalSearch) findDecomp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {
		return baseCase(g.graph, H, Sp)
	}

	//Early termination
	if len(H.edges) <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	//find a balanced separator
	edges := filterVertices(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	log.Printf("Starting Search: Edges: %v K: %v\n", len(edges), K)

	gen := getCombin(len(edges), K)

	log.Printf("Remaming combin: %v \n", gen.left)

	log.Printf("Remaining combinLib: %v \n", gen.current.remaining)

OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)

		log.Printf("Testing: %v\n", Graph{edges: balsep})
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, compsSp, _ := H.getComponents(balsep, Sp)

		log.Printf("Comps of Sep: %v\n", comps)

		SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

		var subtrees []Decomp
		for i := range comps {
			decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		//Create a new GHD for H
		reroot_node := Node{lambda: balsep}
		for _, s := range subtrees {
			s.root = s.root.reroot(Node{lambda: balsep})
			log.Printf("Rerooted Decomp: %v\n", s)
			reroot_node.children = append(reroot_node.children, s.root.children...)
		}
		return Decomp{graph: H, root: reroot_node}
	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject

}

func (g GlobalSearch) findGHD(K int) Decomp {
	return g.findDecomp(K, g.graph, []Special{})
}

func parallelSearch(H Graph, Sp []Special, edges []Edge, result *[]int, generators []*Combin) {
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
		go worker(i, H, Sp, edges, found, generators[i], &wg, &finished)
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

func worker(workernum int, H Graph, Sp []Special, edges []Edge, found chan []int, gen *Combin, wg *sync.WaitGroup, finished *bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d 'forced' to quit, reason: %v", workernum, r)
			return
		}
	}()
	defer wg.Done()

	for gen.hasNext() {
		if *finished {
			log.Printf("Worker %d told to quit", workernum)
			return
		}
		j := gen.combination
		if H.checkBalancedSep(getSubset(edges, j), Sp) {
			log.Printf("Worker %d found a bal sep", workernum)
			found <- j
			log.Printf("Worker %d \" won \"", workernum)
			gen.confirm()
			*finished = true
			return
		}
		gen.confirm()
	}
}

func (g GlobalSearch) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {

	log.Printf("Current Subgraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {
		return baseCase(g.graph, H, Sp)
	}

	//Early termination
	if len(H.edges) <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	var balsep []Edge

	var decomposed = false
	edges := filterVertices(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := getCombin(len(g.graph.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = getSubset(edges, found)

		log.Printf("Balanced Sep chosen: %+v\n", balsep)

		comps, compsSp, _ := H.getComponents(balsep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
				ch <- g.findDecompParallelFull(K, comps[i], append(compsSp[i], SepSpecial))
			}(K, i, comps, compsSp, SepSpecial)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	//Create a new GHD for H
	reroot_node := Node{lambda: balsep}
	for _, s := range subtrees {
		s.root = s.root.reroot(Node{lambda: balsep})
		log.Printf("Rerooted Decomp: %+v\n", s)
		reroot_node.children = append(reroot_node.children, s.root.children...)
	}

	return Decomp{graph: H, root: reroot_node}
}

func (g GlobalSearch) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

	log.Printf("Current Subgraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {
		return baseCase(g.graph, H, Sp)
	}

	//Early termination
	if len(H.edges) <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	var balsep []Edge

	var decomposed = false
	edges := filterVertices(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	// var numProc = runtime.GOMAXPROCS(-1)
	// var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int, 100)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	// generator := getCombin(len(g.graph.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), false)

	var subtrees []Decomp
	// done := make(chan struct{})

	//find a balanced separator
OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearch(H, Sp, edges, &found, generators)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = getSubset(edges, found)
		// close(done) // signal to workers to stop

		log.Printf("Balanced Sep chosen: %+v\n", balsep)

		comps, compsSp, _ := H.getComponents(balsep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

		for i := range comps {
			decomp := g.findDecompParallelSearch(K, comps[i], append(compsSp[i], SepSpecial))
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		decomposed = true
	}

	//Create a new GHD for H
	reroot_node := Node{lambda: balsep}
	for _, s := range subtrees {
		s.root = s.root.reroot(Node{lambda: balsep})
		log.Printf("Rerooted Decomp: %+v\n", s)
		reroot_node.children = append(reroot_node.children, s.root.children...)
	}

	return Decomp{graph: H, root: reroot_node}
}

func (g GlobalSearch) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {
		return baseCase(g.graph, H, Sp)
	}

	//Early termination
	if len(H.edges) <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	//find a balanced separator
	edges := filterVertices(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := getCombin(len(edges), K)
OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, compsSp, _ := H.getComponents(balsep, Sp)

		log.Printf("Comps of Sep: %v\n", comps)

		SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

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
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
				subtrees = []Decomp{}
				//adapt search space for next iteration
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		//Create a new GHD for H
		reroot_node := Node{lambda: balsep}
		for _, s := range subtrees {
			s.root = s.root.reroot(Node{lambda: balsep})
			log.Printf("Rerooted Decomp: %v\n", s)
			reroot_node.children = append(reroot_node.children, s.root.children...)
		}

		return Decomp{graph: H, root: reroot_node}

	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

func (g GlobalSearch) findGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g.graph, []Special{})
}

func (g GlobalSearch) findGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g.graph, []Special{})
}

func (g GlobalSearch) findGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g.graph, []Special{})
}
