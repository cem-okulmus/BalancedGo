package algorithms

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type DetKDecomp struct {
	Graph     Graph
	BalFactor int
}

type CompCache struct {
	succ []uint32
	fail []uint32
}

var cache map[uint32]CompCache

//Note: as implemented this breaks Special Condition (bag must be limited by oldSep)
func baseCaseDetK(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children Node

	switch len(Sp) {
	case 0:
		return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges}}
	case 1:
		children = Node{Bag: Sp[0].Vertices, Cover: Sp[0].Edges}
	case 2:
		children = Node{Bag: Sp[0].Vertices, Cover: Sp[0].Edges,
			Children: []Node{Node{Bag: Sp[1].Vertices, Cover: Sp[1].Edges}}}

	}

	if H.Edges.Len() == 0 {
		return Decomp{Graph: H, Root: children}
	}
	return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges, Children: []Node{children}}}
}

// TODO add caching to this
func (d DetKDecomp) findDecomp(K int, H Graph, oldSep Edges, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if H.Edges.Len() <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.Graph, H, Sp)
	}
	verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
	conn := Inter(oldSep.Vertices(), verticesCurrent)
	compVertices := Diff(verticesCurrent, oldSep.Vertices())
	bound := FilterVertices(d.Graph.Edges, conn)

	gen := NewCover(K, conn, bound, H.Edges)

OUTER:
	for gen.HasNext {
		gen.NextSubset()
		var sep Edges
		sep = GetSubset(bound, gen.Subset)

		addEdges := false

		// // check if sep covers the intersection of oldsep and H
		// if !Subset(conn, sep.Vertices()) {
		// 	continue
		// }

		//check if sep "makes some progress" into separating H
		if len(Inter(sep.Vertices(), compVertices)) == 0 {
			addEdges = true
		}

		if !addEdges || K-sep.Len() > 0 {
			i_add := 0

		addingEdges:
			for !addEdges || i_add < H.Edges.Len() {
				var sepActual Edges

				if addEdges {
					sepActual = NewEdges(append(sep.Slice(), H.Edges.Slice()[i_add]))
				} else {
					sepActual = sep
				}

				comps, compsSp, _ := H.GetComponents(sepActual, Sp)

				_, ok := cache[sep.Hash()]
				if !ok {
					cache[sep.Hash()] = CompCache{}
				}
				compCache, _ := cache[sepActual.Hash()]
				for comp := range compCache.fail {
					if reflect.DeepEqual(comp, H) {
						fmt.Println("Seen before, skipping")
						continue OUTER
					}
				}

				var subtrees []Node
				for i := range comps {
					decomp := d.findDecomp(K, comps[i], sepActual, compsSp[i])
					if reflect.DeepEqual(decomp, Decomp{}) {
						log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sep}, comps[i], compsSp[i])
						log.Printf("\n\nCurrent SubGraph: %v\n", H)
						log.Printf("Current Special Edges: %v\n\n", Sp)

						compCache.fail = append(compCache.fail, H.Edges.Hash())
						if addEdges {
							i_add++
							continue addingEdges
						} else {
							continue OUTER
						}
					}

					log.Printf("Produced Decomp: %v\n", decomp)
					subtrees = append(subtrees, decomp.Root)
				}
				compCache.succ = append(compCache.succ, H.Edges.Hash())
				bag := Inter(sepActual.Vertices(), append(oldSep.Vertices(), H.Vertices()...))
				return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sepActual, Children: subtrees}}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}

func parallelSearchDetK(H Graph, Sp []Special, edges Edges, result *[]int, generators []*CombinationIterator, oldSep Edges) {
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

func workerDetK(workernum int, H Graph, Sp []Special, edges Edges, found chan []int, gen *CombinationIterator, wg *sync.WaitGroup, finished *bool, oldSep Edges) {
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

		if H.CheckNextSep(GetSubset(edges, j), oldSep, Sp) {
			found <- j
			log.Printf("Worker %d \" won \"", workernum)
			gen.Confirm()
			*finished = true
			return
		}
		gen.Confirm()
	}
}

func (d DetKDecomp) findDecompParallelFull(K int, H Graph, oldSep Edges, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if H.Edges.Len() <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.Graph, H, Sp)
	}

	var sep Edges
	var subtrees []Node
	var decomposed = false

	generators := SplitCombin(len(d.Graph.Edges.Slice()), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearchDetK(H, Sp, d.Graph.Edges, &found, generators, oldSep)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find sep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		sep = GetSubset(d.Graph.Edges, found)

		log.Printf("Sep chosen: %+v\n", Graph{Edges: sep})

		comps, compsSp, _ := H.GetComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, sep Edges) {
				ch <- d.findDecompParallelFull(K, comps[i], sep, compsSp[i])
			}(K, i, comps, compsSp, sep)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				subtrees = []Node{}
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.Root)
		}

		decomposed = true
	}

	bag := Inter(sep.Vertices(), append(oldSep.Vertices(), H.Vertices()...))
	return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sep, Children: subtrees}}

}

func (d DetKDecomp) findDecompParallelSearch(K int, H Graph, oldSep Edges, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if H.Edges.Len() <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.Graph, H, Sp)
	}

	var sep Edges
	var subtrees []Node
	var decomposed = false

	generators := SplitCombin(len(d.Graph.Edges.Slice()), K, runtime.GOMAXPROCS(-1), false)

OUTER:
	for !decomposed {
		var found []int

		//g.startSearchSimple(&found, &generator, result, input, &wg)
		parallelSearchDetK(H, Sp, d.Graph.Edges, &found, generators, oldSep)

		if len(found) == 0 { // meaning that the search above never found anything
			log.Printf("REJECT: Couldn't find sep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		sep = GetSubset(d.Graph.Edges, found)

		log.Printf("Sep chosen: %+v\n", Graph{Edges: sep})

		comps, compsSp, _ := H.GetComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)

		for i := range comps {
			decomp := d.findDecompParallelSearch(K, comps[i], sep, compsSp[i])
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				subtrees = []Node{}
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.Root)
		}

		decomposed = true
	}

	bag := Inter(sep.Vertices(), append(oldSep.Vertices(), H.Vertices()...))
	return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sep, Children: subtrees}}

}

func (d DetKDecomp) findDecompParallelDecomp(K int, H Graph, oldSep Edges, Sp []Special) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if H.Edges.Len() <= K && len(Sp) <= 2 {
		return baseCaseDetK(d.Graph, H, Sp)
	}

	gen := GetCombin(len(d.Graph.Edges.Slice()), K)

OUTER:
	for gen.HasNext() {
		gen.Confirm()
		sep := GetSubset(d.Graph.Edges, gen.Combination)

		verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
		// check if sep covers the intersection of oldsep and H
		if !Subset(Inter(oldSep.Vertices(), verticesCurrent), sep.Vertices()) {
			continue
		}
		//check if sep "makes some progress" into separating H
		if len(Inter(sep.Vertices(), Diff(verticesCurrent, oldSep.Vertices()))) == 0 {
			continue
		}

		log.Printf("Sep chosen: %+v\n", Graph{Edges: sep})

		comps, compsSp, _ := H.GetComponents(sep, Sp)

		log.Printf("Comps of Sep: %+v\n", comps)
		var subtrees []Node
		ch := make(chan Decomp)
		for i := range comps {
			go func(K int, i int, comps []Graph, compsSp [][]Special, sep Edges) {
				ch <- d.findDecompParallelDecomp(K, comps[i], sep, compsSp[i])
			}(K, i, comps, compsSp, sep)
		}

		for i := range comps {
			decomp := <-ch
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent SubGraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.Root)
		}

		bag := Inter(sep.Vertices(), append(oldSep.Vertices(), H.Vertices()...))
		return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sep, Children: subtrees}}
	}

	return Decomp{}
}

func (d DetKDecomp) FindHD(K int, Sp []Special) Decomp {
	cache = make(map[uint32]CompCache)
	return d.findDecomp(K, d.Graph, Edges{}, Sp)
}

func (d DetKDecomp) FindHDParallelFull(K int, Sp []Special) Decomp {
	return d.findDecompParallelFull(K, d.Graph, Edges{}, Sp)
}

func (d DetKDecomp) FindHDParallelSearch(K int, Sp []Special) Decomp {
	return d.findDecompParallelSearch(K, d.Graph, Edges{}, Sp)
}

func (d DetKDecomp) FindHDParallelDecomp(K int, Sp []Special) Decomp {
	return d.findDecompParallelDecomp(K, d.Graph, Edges{}, Sp)
}
