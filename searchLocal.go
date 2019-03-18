package main

import (
	"log"
	"reflect"
	"runtime"
)

type LocalSearch struct {
	graph Graph
}

func (g LocalSearch) findDecomp(K int, H Graph, Sp []Special) Decomp {

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
	edges := cutEdges(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := getCombinUnextend(len(edges), K)
OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}
		var sepSub *SepSub

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})
		//balsep_orig := balsep
	INNER:
		for {
			comps, compsSp, _ := H.getComponents(balsep, Sp)

			log.Printf("Comps of Sep: %v\n", comps)

			SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

			var subtrees []Decomp

			for i := range comps {
				decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
				if reflect.DeepEqual(decomp, Decomp{}) {

					log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
					// log.Printf("\n\nCurrent Subgraph: %v\n", H)
					// log.Printf("Current Special Edges: %v\n\n", Sp)
					if sepSub == nil {
						sepSub = getSepSub(g.graph.edges, balsep, K)
					}
					next_balsep_found := false

					for !next_balsep_found {
						if sepSub.hasNext() {
							balsep = sepSub.getCurrent()
							//log.Printf("Testing Sep: %v of %v , Special Edges %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
							if H.checkBalancedSep(balsep, Sp) {
								next_balsep_found = true
							}
						} else {
							continue OUTER
						}
					}
					//log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
					continue INNER
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

	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject

}

func (g LocalSearch) findGHD(K int) Decomp {
	return g.findDecomp(K, g.graph, []Special{})
}

func (g LocalSearch) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {

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
	edges := cutEdges(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := getCombin(len(g.graph.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), true)

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
		var sepSub *SepSub

		log.Printf("Balanced Sep chosen: %+v\n", balsep)
		balsep_orig := balsep
	INNER:
		for !decomposed {
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
					// log.Printf("\n\nCurrent Subgraph: %v\n", H)
					// log.Printf("Current Special Edges: %v\n\n", Sp)

					if sepSub == nil {
						sepSub = getSepSub(g.graph.edges, balsep, K)
					}
					next_balsep_found := false

					for !next_balsep_found {
						if sepSub.hasNext() {
							balsep = sepSub.getCurrent()
							log.Printf("Testing Sep: %v of %v , Special Edges %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
							if H.checkBalancedSep(balsep, Sp) {
								next_balsep_found = true
							}
						} else {
							continue OUTER
						}
					}
					//log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
					continue INNER
				}

				log.Printf("Produced Decomp: %+v\n", decomp)

				subtrees = append(subtrees, decomp)

			}
			decomposed = true
		}

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

func (g LocalSearch) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

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
	edges := cutEdges(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := getCombin(len(g.graph.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1), true)

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
		var sepSub *SepSub

		log.Printf("Balanced Sep chosen: %+v\n", balsep)
	INNER:
		for !decomposed {
			comps, compsSp, _ := H.getComponents(balsep, Sp)

			log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := Special{edges: balsep, vertices: Vertices(balsep)}

			for i := range comps {
				decomp := g.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
				if reflect.DeepEqual(decomp, Decomp{}) {
					log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], append(compsSp[i], SepSpecial))
					subtrees = []Decomp{}
					// log.Printf("\n\nCurrent Subgraph: %v\n", H)
					// log.Printf("Current Special Edges: %v\n\n", Sp)
					if sepSub == nil {
						sepSub = getSepSub(g.graph.edges, balsep, K)
					}
					next_balsep_found := false

					for !next_balsep_found {
						if sepSub.hasNext() {
							balsep = sepSub.getCurrent()
							//log.Printf("Testing Sep: %v of %v , Special Edges %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
							if H.checkBalancedSep(balsep, Sp) {
								next_balsep_found = true
							}
						} else {
							continue OUTER
						}
					}
					//log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
					continue INNER
				}

				log.Printf("Produced Decomp: %v\n", decomp)

				subtrees = append(subtrees, decomp)
			}
			decomposed = true
		}

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

func (g LocalSearch) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

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
	edges := cutEdges(g.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := getCombin(len(edges), K)
OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}
		var sepSub *SepSub

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})
		//balsep_orig := balsep
	INNER:
		for {
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
					// log.Printf("\n\nCurrent Subgraph: %v\n", H)
					// log.Printf("Current Special Edges: %v\n\n", Sp)
					if sepSub == nil {
						sepSub = getSepSub(g.graph.edges, balsep, K)
					}
					next_balsep_found := false

					for !next_balsep_found {
						if sepSub.hasNext() {
							balsep = sepSub.getCurrent()
							//log.Printf("Testing Sep: %v of %v , Special Edges %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
							if H.checkBalancedSep(balsep, Sp) {
								next_balsep_found = true
							}
						} else {
							continue OUTER
						}
					}
					//log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{edges: balsep}, Graph{edges: balsep_orig}, Sp)
					continue INNER
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
	}

	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return Decomp{} // empty Decomp signifiyng reject
}

func (g LocalSearch) findGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g.graph, []Special{})
}

func (g LocalSearch) findGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g.graph, []Special{})
}

func (g LocalSearch) findGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g.graph, []Special{})
}
