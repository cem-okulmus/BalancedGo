package main

import (
	"log"
	"reflect"
	"runtime"
	"sync"
)

type detKDecomp struct {
	graph Graph
}

//Note: as implemented this breaks Special Condition (bag must be limited by oldSep)
func baseCaseDetK(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children Node

	switch len(Sp) {
	case 0:
		return Decomp{graph: H, root: Node{bag: H.Vertices(), cover: H.edges}}
	case 1:
		children = Node{bag: Sp[0].vertices, cover: Sp[0].edges}
	case 2:
		children = Node{bag: Sp[0].vertices, cover: Sp[0].edges,
			children: []Node{Node{bag: Sp[1].vertices, cover: Sp[1].edges}}}

	}

	if len(H.edges) == 0 {
		return Decomp{graph: H, root: children}
	}
	return Decomp{graph: H, root: Node{bag: H.Vertices(), cover: H.edges, children: []Node{children}}}
}

// TODO add caching to this
func (d detKDecomp) findDecomp(K int, H Graph, oldSep []Edge, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if len(H.edges) <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.graph, H, Sp)
	}

	//TODO: think about whether filtering here is allowed, and if it should be strict or not
	edges := filterVerticesStrict(d.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))
	gen := getCombin(len(edges), K)

OUTER:
	for gen.hasNext() {
		gen.confirm()
		sep := getSubset(edges, gen.combination)

		verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
		// check if sep covers the intersection of oldsep and H
		if !subset(inter(Vertices(oldSep), verticesCurrent), Vertices(sep)) {
			continue
		}
		//check if sep "makes some progress" into separating H
		if len(inter(Vertices(sep), diff(verticesCurrent, Vertices(oldSep)))) == 0 {
			continue
		}

		comps, compsSp, _ := H.getComponents(sep, Sp)

		var subtrees []Node
		for i := range comps {
			decomp := d.findDecomp(K, comps[i], sep, compsSp[i])
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		bag := inter(Vertices(sep), append(Vertices(oldSep), H.Vertices()...))
		return Decomp{graph: H, root: Node{bag: bag, cover: sep, children: subtrees}}
	}

	return Decomp{} // Reject if no separator could be found
}

func parallelSearchDetK(H Graph, Sp []Special, edges []Edge, result *[]int, generators []*CombinationIterator, oldSep []Edge) {
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
		go workerDetK(i, H, Sp, edges, found, generators[i], &wg, &finished, oldSep)
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

func workerDetK(workernum int, H Graph, Sp []Special, edges []Edge, found chan []int, gen *CombinationIterator, wg *sync.WaitGroup, finished *bool, oldSep []Edge) {
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

		if H.checkNextSep(getSubset(edges, j), oldSep, Sp) {
			found <- j
			log.Printf("Worker %d \" won \"", workernum)
			gen.confirm()
			*finished = true
			return
		}
		gen.confirm()
	}
}

func (d detKDecomp) findDecompParallelFull(K int, H Graph, oldSep []Edge, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if len(H.edges) <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.graph, H, Sp)
	}

	var sep []Edge
	var subtrees []Node
	var decomposed = false

	//TODO: think about whether filtering here is allowed, and if it should be strict or not
	edges := filterVerticesStrict(d.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearchDetK(H, Sp, edges, &found, generators, oldSep)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find sep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		sep = getSubset(edges, found)

		log.Printf("Sep chosen: %+v\n", Graph{edges: sep})

		comps, compsSp, _ := H.getComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, sep []Edge) {
				ch <- d.findDecompParallelFull(K, comps[i], sep, compsSp[i])
			}(K, i, comps, compsSp, sep)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				subtrees = []Node{}
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		decomposed = true
	}

	bag := inter(Vertices(sep), append(Vertices(oldSep), H.Vertices()...))
	return Decomp{graph: H, root: Node{bag: bag, cover: sep, children: subtrees}}

}

func (d detKDecomp) findDecompParallelSearch(K int, H Graph, oldSep []Edge, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if len(H.edges) <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.graph, H, Sp)
	}

	var sep []Edge
	var subtrees []Node
	var decomposed = false

	//TODO: think about whether filtering here is allowed, and if it should be strict or not
	edges := filterVerticesStrict(d.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearchDetK(H, Sp, edges, &found, generators, oldSep)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find sep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		sep = getSubset(edges, found)

		log.Printf("Sep chosen: %+v\n", Graph{edges: sep})

		comps, compsSp, _ := H.getComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		for i := range comps {
			decomp := d.findDecompParallelSearch(K, comps[i], sep, compsSp[i])
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				subtrees = []Node{}
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		decomposed = true
	}

	bag := inter(Vertices(sep), append(Vertices(oldSep), H.Vertices()...))
	return Decomp{graph: H, root: Node{bag: bag, cover: sep, children: subtrees}}

}

func (d detKDecomp) findDecompParallelDecomp(K int, H Graph, oldSep []Edge, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if len(H.edges) <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.graph, H, Sp)
	}

	//TODO: think about whether filtering here is allowed, and if it should be strict or not
	edges := filterVerticesStrict(d.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))
	gen := getCombin(len(edges), K)

OUTER:
	for gen.hasNext() {
		gen.confirm()
		sep := getSubset(edges, gen.combination)

		verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
		// check if sep covers the intersection of oldsep and H
		if !subset(inter(Vertices(oldSep), verticesCurrent), Vertices(sep)) {
			continue
		}
		//check if sep "makes some progress" into separating H
		if len(inter(Vertices(sep), diff(verticesCurrent, Vertices(oldSep)))) == 0 {
			continue
		}

		log.Printf("Sep chosen: %+v\n", Graph{edges: sep})

		comps, compsSp, _ := H.getComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)
		var subtrees []Node
		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, sep []Edge) {
				ch <- d.findDecompParallelDecomp(K, comps[i], sep, compsSp[i])
			}(K, i, comps, compsSp, sep)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		bag := inter(Vertices(sep), append(Vertices(oldSep), H.Vertices()...))
		return Decomp{graph: H, root: Node{bag: bag, cover: sep, children: subtrees}}
	}

	return Decomp{}
}

func (d detKDecomp) findHD(K int, Sp []Special) Decomp {
	return d.findDecomp(K, d.graph, []Edge{}, Sp)
}

func (d detKDecomp) findHDParallelFull(K int, Sp []Special) Decomp {
	return d.findDecompParallelFull(K, d.graph, []Edge{}, Sp)
}

func (d detKDecomp) findHDParallelSearch(K int, Sp []Special) Decomp {
	return d.findDecompParallelSearch(K, d.graph, []Edge{}, Sp)
}

func (d detKDecomp) findHDParallelDecomp(K int, Sp []Special) Decomp {
	return d.findDecompParallelDecomp(K, d.graph, []Edge{}, Sp)
}
