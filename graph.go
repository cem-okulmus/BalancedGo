package main

import (
	"bytes"
	"github.com/spakin/disjoint"
	"log"
	"reflect"
	"runtime"
	"sync"
)

type Graph struct {
	edges []Edge
	m     map[int]string
}

func (g Graph) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, e := range g.edges {
		buffer.WriteString(e.String())
		if i != len(g.edges)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (g Graph) Vertices() []int {
	var output []int
	for _, otherE := range g.edges {
		output = append(output, otherE.nodes...)
	}
	return removeDuplicates(output)
}

func getSubset(edges []Edge, s []int) []Edge {
	var output []Edge
	for _, i := range s {
		output = append(output, edges[i])
	}
	return output
}

func (g Graph) getSubset(s []int) []Edge {
	return getSubset(g.edges, s)
}

// Uses Disjoint Set data structure to compute connected components
func (g Graph) getComponents(sep []Edge, Sp []Special) ([]Graph, [][]Special) {
	var outputG []Graph
	var outputS [][]Special

	var nodes = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Special)

	balsepVert := Vertices(sep)

	//  Set up the disjoint sets for each node
	for _, i := range append(Vertices(g.edges), VerticesSpecial(Sp)...) {
		nodes[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for _, e := range g.edges {
		actualNodes := diff(e.nodes, balsepVert)
		for i := 0; i < len(actualNodes)-1; i++ {
			disjoint.Union(nodes[actualNodes[i]], nodes[actualNodes[i+1]])
		}
	}

	for _, s := range Sp {
		actualNodes := diff(s.vertices, balsepVert)
		for i := 0; i < len(actualNodes)-1; i++ {
			disjoint.Union(nodes[actualNodes[i]], nodes[actualNodes[i+1]])
		}
	}

	//sort each edge and special edge to a corresponding component
	for _, e := range g.edges {
		actualNodes := diff(e.nodes, balsepVert)
		if len(actualNodes) > 0 {
			comps[nodes[actualNodes[0]].Find()] = append(comps[nodes[actualNodes[0]].Find()], e)
		}
	}
	var isolatedSp []Special
	for _, s := range Sp {
		actualNodes := diff(s.vertices, balsepVert)
		if len(actualNodes) > 0 {
			compsSp[nodes[actualNodes[0]].Find()] = append(compsSp[nodes[actualNodes[0]].Find()], s)
		} else {
			isolatedSp = append(isolatedSp, s)
		}
	}

	// Store the components as graphs
	for k, _ := range comps {
		g := Graph{edges: comps[k]}
		outputG = append(outputG, g)
		outputS = append(outputS, compsSp[k])

	}

	for k, _ := range compsSp {
		_, ok := comps[k]
		if ok {
			continue
		}
		g := Graph{}
		outputG = append(outputG, g)
		outputS = append(outputS, compsSp[k])
	}

	for _, s := range isolatedSp {
		g := Graph{}
		outputG = append(outputG, g)
		outputS = append(outputS, []Special{s})
	}

	return outputG, outputS
}

func filterVertices(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if subset(e.nodes, vertices) {
			output = append(output, e)
		}
	}

	return output

}

func (g Graph) checkBalancedSep(sep []Edge, sp []Special) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, compSps := g.getComponents(sep, sp)
	// log.Printf("Components of sep %+v\n", comps)
	for i := range comps {
		if len(comps[i].edges)+len(compSps[i]) > ((len(g.edges) + len(sp)) / 2) {
			// log.Printf("Component %+v has weight%d instead of %d\n", comps[i], len(comps[i].edges)+len(compSps[i]), ((len(g.edges) + len(sp)) / 2))
			return false
		}
	}

	// // Check if subset of V(H) + Vertices of Sp
	// var allowedVertices = append(g.Vertices(), VerticesSpecial(sp)...)
	// if !subset(Vertices(sep), allowedVertices) {
	// 	// log.Println("Subset condition violated")
	// 	return false
	// }

	// Make sure that "special seps can never be used as separators"
	for _, s := range sp {
		if reflect.DeepEqual(s.vertices, Vertices(sep)) {
			// log.Println("Special edge %+v\n used again", s)
			return false
		}
	}

	return true
}

func (g Graph) computeSubEdges(K int) Graph {
	var output = g

	for _, e := range g.edges {
		graph_wihout_e := Graph{edges: diffEdges(g.edges, []Edge{e})}
		for _, l := range Combinations(len(graph_wihout_e.edges), K) {
			var tuple = Vertices(graph_wihout_e.getSubset(l))
			output.edges = append(output.edges, Edge{nodes: inter(e.nodes, tuple)}.subedges()...)
		}
	}

	output.edges = removeDuplicatesEdges(output.edges)
	return output
}

func (g Graph) findDecomp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {

		log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
		switch len(Sp) {
		case 0:
			return Decomp{graph: g} // use g here to avoid reject
		case 1:
			sp1 := Sp[0]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges}}
		case 2:
			sp1 := Sp[0]
			sp2 := Sp[1]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges,
					children: []Node{Node{lambda: sp2.edges}}}}

		}
	}

	//find a balanced separator
	edges := filterVertices(g.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := getCombin(len(edges), K)
OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, compsSp := H.getComponents(balsep, Sp)

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

func (g Graph) findGHD(K int) Decomp {
	return g.findDecomp(K, g, []Special{})
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

func (g Graph) findDecompParallelFull(K int, H Graph, Sp []Special) Decomp {

	log.Printf("Current Subgraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {

		log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
		switch len(Sp) {
		case 2:
			sp1 := Sp[0]
			sp2 := Sp[1]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges,
					children: []Node{Node{lambda: sp2.edges}}}}
		default:
			log.Panicf("Wrong base case: ", Sp)

		}
	}
	var balsep []Edge

	var decomposed = false
	edges := filterVertices(g.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	//var numProc = runtime.GOMAXPROCS(-1)
	//var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	//generator := getCombin(len(g.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1))

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

		comps, compsSp := H.getComponents(balsep, Sp)

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

func (g Graph) findDecompParallelSearch(K int, H Graph, Sp []Special) Decomp {

	log.Printf("Current Subgraph: %+v\n", H)
	log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {

		log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
		switch len(Sp) {
		case 2:
			sp1 := Sp[0]
			sp2 := Sp[1]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges,
					children: []Node{Node{lambda: sp2.edges}}}}
		default:
			log.Panicf("Wrong base case: ", Sp)

		}
	}
	var balsep []Edge

	var decomposed = false
	edges := filterVertices(g.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	// var numProc = runtime.GOMAXPROCS(-1)
	// var wg sync.WaitGroup
	// wg.Add(numProc)
	// result := make(chan []int)
	// input := make(chan []int, 100)
	// for i := 0; i < numProc; i++ {
	// 	go g.workerSimple(H, Sp, result, input, &wg)
	// }
	// generator := getCombin(len(g.edges), K)

	generators := splitCombin(len(edges), K, runtime.GOMAXPROCS(-1))

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

		comps, compsSp := H.getComponents(balsep, Sp)

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

func (g Graph) findDecompParallelComp(K int, H Graph, Sp []Special) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)

	//stop if there are at most two special edges left
	if len(H.edges) == 0 && len(Sp) <= 2 {

		log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
		switch len(Sp) {
		case 0:
			return Decomp{graph: g} // use g here to avoid reject
		case 1:
			sp1 := Sp[0]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges}}
		case 2:
			sp1 := Sp[0]
			sp2 := Sp[1]
			return Decomp{graph: H,
				root: Node{lambda: sp1.edges,
					children: []Node{Node{lambda: sp2.edges}}}}

		}
	}

	//find a balanced separator
	edges := filterVertices(g.edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	gen := getCombin(len(edges), K)
OUTER:
	for gen.hasNext() {
		balsep := getSubset(edges, gen.combination)
		gen.confirm()
		if !H.checkBalancedSep(balsep, Sp) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, compsSp := H.getComponents(balsep, Sp)

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

func (g Graph) findGHDParallelFull(K int) Decomp {
	return g.findDecompParallelFull(K, g, []Special{})
}

func (g Graph) findGHDParallelSearch(K int) Decomp {
	return g.findDecompParallelSearch(K, g, []Special{})
}

func (g Graph) findGHDParallelComp(K int) Decomp {
	return g.findDecompParallelComp(K, g, []Special{})
}
