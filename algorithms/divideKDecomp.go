// Follows the algorithm k divide decomp from Samer "Exploiting Parallelism in Decomposition Methods for Constraint Satisfaction"
package algorithms

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type DivideKDecomp struct {
	Graph     Graph
	K         int
	BalFactor int
}

func (d DivideKDecomp) parallelSearch(H DivideComp, edges Edges, result *[]int, gens []*CombinationIterator, covSep Edges) {
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
		go d.worker(i, H, edges, found, gens[i], &wg, &finished, covSep)
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

func (d DivideKDecomp) worker(workernum int, H DivideComp, edges Edges, found chan []int, gen *CombinationIterator, wg *sync.WaitGroup, finished *bool, covSep Edges) {
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

		balsep := GetSubset(edges, j)
		//extend balsep with sep above
		balsep = NewEdges(append(balsep.Slice(), covSep.Slice()...))

		comps, valid := H.GetComponents(balsep)
		if d.CheckBalancedSep(H, comps, valid) {
			found <- j
			log.Printf("Worker %d \" won \"", workernum)
			gen.Confirm()
			*finished = true
			return
		}
		gen.Confirm()
	}
}

func postProcess(tree Node, parentBag []int) Node {
	var output Node

	output = tree

	treeComp := append(output.Bag, parentBag...)

	output.Bag = Inter(output.Cover.Vertices(), treeComp)

	for i := range output.Children {
		output.Children[i] = postProcess(output.Children[i], output.Bag)
	}

	return output
}

func (d DivideKDecomp) CheckBalancedSep(comp DivideComp, comps []DivideComp, valid bool) bool {
	//check if up and low separated
	// constant check enough as all vertices in up (resp. low) part of the same comp
	if !valid {
		log.Println("Up and low not separated")
		log.Println("Current: ", comp, "\n\n")

		for i := range comps {
			log.Println(comps[i], "\n")
		}

		return false
	}

	// TODO, make this work only a single loop
	if len(comp.Low) == 0 {
		// log.Printf("Components of sep %+v\n", comps)
		for i := range comps {

			if comps[i].Edges.Len() == comp.Edges.Len() { // not made any progres
				log.Println("No progress made")
				return false
			}

			if comps[i].Length > (((comp.Edges.Len()) * (d.BalFactor - 1)) / d.BalFactor) {
				// log.Printf("Using component %+v has weight %d instead of %d\n", comps[i], comps[i].Edges.Len(), (((comp.Edges.Len()) * (d.BalFactor - 1)) / d.BalFactor))
				return false
			}
		}
	} else {
		for i := range comps {
			if comps[i].Edges.Len() == comp.Edges.Len() { // not made any progres
				// log.Println("No progress made")
				return false
			}
			if len(comps[i].Low) == 0 {
				continue
			}
			if comps[i].Length > (((comp.Edges.Len()) * (d.BalFactor - 1)) / d.BalFactor) {
				// log.Printf("Using component %+v has weight %d instead of %d\n", comps[i], comps[i].Edges.Len(), (((comp.Edges.Len()) * (d.BalFactor - 1)) / d.BalFactor))
				return false
			}
		}
		// if len(Inter(sep.Vertices(), comp.Edges.Vertices())) == 0 { //make some progress
		// 	return false
		// }
	}

	return true
}

func reorderComps(parent Node, subtree Node) Node {
	log.Println("Two Nodes enter: ", parent, subtree)
	// up = Inter(up, parent.Vertices())
	//finding connecting leaf in parent
	leaf := parent.CombineNodes(subtree)
	if reflect.DeepEqual(*leaf, Node{}) {
		fmt.Println("\n \n comp ", PrintVertices(parent.Low))
		fmt.Println("parent ", parent)

		log.Panicln("parent tree doesn't contain connecting node!")
	}

	// //attaching subtree to parent
	// leaf.Children = []Node{subtree}
	log.Println("Leaf ", leaf)
	log.Println("One Node leaves: ", parent)
	return *leaf
}

func (d DivideKDecomp) baseCase(comp DivideComp) Decomp {
	log.Println("Using base case")
	det := DetKDecomp{Graph: d.Graph, BalFactor: d.BalFactor, SubEdge: false, Divide: true}
	det.cache = make(map[uint32]*CompCache)
	var H Graph

	H.Edges = NewEdges(comp.Edges.Slice())

	log.Println("up", PrintVertices(comp.Up))
	log.Println("H", H.Edges)
	var output Decomp
	if len(comp.Low) == 0 {
		output = det.findDecomp(d.K, H, comp.Up, []Special{})
	} else {
		output = det.findDecomp(d.K, H, comp.Up, []Special{Special{Vertices: comp.Low}})

	}

	if !reflect.DeepEqual(output, Decomp{}) {
		output.UpConnecting = comp.UpConnecting
	}

	return output
}

func (d DivideKDecomp) decomp(comp DivideComp) Decomp {

	// if !Subset(comp.Low, comp.Edges.Vertices()) {
	// 	log.Println("comp ", comp)
	// 	log.Panicln("low set not inside edges")
	// }
	if !Subset(comp.Up, comp.Edges.Vertices()) {
		log.Println("comp ", comp)
		log.Panicln("up set not inside edges")
	}

	log.Printf("\n\nCurrent SubGraph: %v\n", comp)

	//base case: size of comp <= K

	if comp.Edges.Len() <= 2*d.K {
		output := d.baseCase(comp)
		log.Println("output ", output)
		if reflect.DeepEqual(output, Decomp{}) {
			log.Printf("REJECTING base: couldn't decompose %v \n", comp)
			return Decomp{}
		}
		return output
	}

	var conn []int
	var genCov Cover

	conn = Inter(comp.Low, comp.Up) // What happens if this is empty?
	edges := FilterVertices(d.Graph.Edges, comp.Edges.Vertices())
	genCov = NewCover(d.K, conn, edges, comp.Edges)

COVER:
	for genCov.HasNext {
		out := genCov.NextSubset()

		if out == -1 {
			if genCov.HasNext {
				log.Panicln(" -1 but hasNext not false!")
			}
			continue COVER
		}

		var sep Edges
		sep = GetSubset(edges, genCov.Subset)

		log.Println("Cover found", sep)

		var firstPass bool // used to skip extension for first run of OUTER if sep nonempty
		if sep.Len() > 0 {
			firstPass = true
		}

		gen := GetCombin(edges.Len(), d.K-sep.Len())

	OUTER:
		for gen.HasNext() {

			var balsep Edges

			if !firstPass {
				gen.Confirm()
				balsep = GetSubset(edges, gen.Combination)
				//extend balsep with sep above
				balsep = NewEdges(append(balsep.Slice(), sep.Slice()...))
			} else {
				balsep = sep

			}
			firstPass = false

			comps, valid := comp.GetComponents(balsep)

			if !d.CheckBalancedSep(comp, comps, valid) {
				continue
			}
			log.Println("Chosen Sep ", balsep)

			log.Printf("Comps of Sep: %v\n", comps)

			var parent Node
			var subtrees []Node
			var upconnecting bool
			// bag := Inter(balsep.Vertices(), verticesExtended)
			// log.Println("The bag is", PrintVertices(bag))
			for i, _ := range comps {
				child := d.decomp(comps[i])
				if reflect.DeepEqual(child, Decomp{}) {
					log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{Edges: balsep}, comps[i])
					log.Printf("\n\nCurrent SubGraph: %v\n", comp)
					continue OUTER
				}

				log.Printf("Produced Decomp: %v\n", child)

				if comps[i].UpConnecting {
					upconnecting = true
					parent = child.Root
					parent.Bag = comps[i].Edges.Vertices()
					parent.Up = comps[i].Up
					parent.Low = comps[i].Low

				} else {
					subtrees = append(subtrees, child.Root)
				}
			}

			var output Node
			upFlag := false
			if len(comp.Low) > 0 && Subset(comp.Low, balsep.Vertices()) {
				upFlag = true
			}

			var bag []int

			if upconnecting {

				for i := range subtrees {
					childComp := append(subtrees[i].Bag, subtrees[i].Up...)
					bag = append(bag, childComp...)
				}

			} else {
				bag = Diff(comp.Edges.Vertices(), comp.Up)
			}

			SubtreeRootedAtS := Node{LowConnecting: upFlag, Up: comp.Up, Low: comp.Low, Cover: balsep, Bag: bag, Children: subtrees}

			if reflect.DeepEqual(parent, Node{}) && (!Subset(comp.Up, balsep.Vertices())) {

				fmt.Println("Subtrees: ")
				for _, s := range subtrees {
					fmt.Println("\n\n", s)
				}

				log.Panicln("Parent missing")
			}

			if upconnecting {
				output = reorderComps(parent, SubtreeRootedAtS)

				log.Printf("Reordered Decomp: %v\n", output)
			} else {
				output = SubtreeRootedAtS
			}

			output.Up = make([]int, len(comp.Up))
			copy(output.Up, output.Up)
			output.Low = make([]int, len(comp.Low))
			copy(output.Low, comp.Low)

			return Decomp{Graph: d.Graph, Root: output}
		}
	}

	log.Println("REJECT: Couldn't find a sep for ", comp)

	return Decomp{} // using empty decomp as reject case

}

func (d DivideKDecomp) decompParallel(comp DivideComp) Decomp {

	if !Subset(comp.Up, comp.Edges.Vertices()) {
		log.Println("comp ", comp)
		log.Panicln("up set not inside edges")
	}

	log.Printf("\n\nCurrent SubGraph: %v\n", comp)

	//base case: size of comp <= K

	if comp.Edges.Len() <= 2*d.K {
		output := d.baseCase(comp)
		log.Println("output ", output)
		if reflect.DeepEqual(output, Decomp{}) {
			log.Printf("REJECTING base: couldn't decompose %v \n", comp)
			return Decomp{}
		}
		return output
	}

	var conn []int
	var genCov Cover

	conn = Inter(comp.Low, comp.Up) // What happens if this is empty?
	edges := FilterVertices(d.Graph.Edges, comp.Edges.Vertices())
	genCov = NewCover(d.K, conn, edges, comp.Edges)

COVER:
	for genCov.HasNext {
		out := genCov.NextSubset()

		if out == -1 {
			if genCov.HasNext {
				log.Panicln(" -1 but hasNext not false!")
			}
			continue COVER
		}

		var decomposed = false
		var sep Edges
		sep = GetSubset(edges, genCov.Subset)

		log.Println("Cover found", sep)

		var firstPass bool // used to skip extension for first run of OUTER if sep nonempty
		if sep.Len() > 0 {
			firstPass = true
		}

		// gen := GetCombin(edges.Len(), d.K-sep.Len())
		gens := SplitCombin(edges.Len(), d.K-sep.Len(), runtime.GOMAXPROCS(-1), false)

	OUTER:
		for !decomposed {
			var balsep Edges
			var found []int

			if !firstPass {
				// gen.Confirm()

				d.parallelSearch(comp, edges, &found, gens, sep)

				if len(found) == 0 { // meaning that the search above never found anything
					log.Printf("REJECT: Couldn't find balsep for Comp %v \n", comp)
					return Decomp{}
				}

				// balsep = GetSubset(edges, gen.Combination)

				//wait until first worker finds a balanced sep
				balsep = GetSubset(edges, found)
				//extend balsep with sep above
				balsep = NewEdges(append(balsep.Slice(), sep.Slice()...))
			} else {
				balsep = sep
			}

			comps, valid := comp.GetComponents(balsep)

			if firstPass {
				firstPass = false
				if !d.CheckBalancedSep(comp, comps, valid) {
					continue
				}
			}

			log.Println("Chosen Sep ", balsep)

			log.Printf("Comps of Sep: %v\n", comps)

			var parent Node
			var subtrees []Node
			var upconnecting bool

			// ch := make(chan Decomp)
			// for i := range comps {
			// 	go func(i int, comps []DivideComp) {

			// 		ch <- d.decompParallel(comps[i])

			// 	}(i, comps)
			// }

			for i := range comps {
				child := d.decompParallel(comps[i])
				// child := <-ch
				if reflect.DeepEqual(child, Decomp{}) {
					log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{Edges: balsep}, comps[i])
					log.Printf("\n\nCurrent SubGraph: %v\n", comp)
					continue OUTER
				}

				log.Printf("Produced Decomp: %v\n", child)

				if child.UpConnecting {
					upconnecting = true
					parent = child.Root
					parent.Bag = comps[i].Edges.Vertices()
					parent.Up = comps[i].Up
					parent.Low = comps[i].Low

				} else {
					subtrees = append(subtrees, child.Root)
				}
			}

			var output Node
			upFlag := false
			if len(comp.Low) > 0 && Subset(comp.Low, balsep.Vertices()) {
				upFlag = true
			}

			var bag []int

			if upconnecting {

				for i := range subtrees {
					childComp := append(subtrees[i].Bag, subtrees[i].Up...)
					bag = append(bag, childComp...)
				}

			} else {
				bag = Diff(comp.Edges.Vertices(), comp.Up)
			}

			SubtreeRootedAtS := Node{LowConnecting: upFlag, Up: comp.Up, Low: comp.Low, Cover: balsep, Bag: bag, Children: subtrees}

			if reflect.DeepEqual(parent, Node{}) && (!Subset(comp.Up, balsep.Vertices())) {

				fmt.Println("Subtrees: ")
				for _, s := range subtrees {
					fmt.Println("\n\n", s)
				}

				log.Panicln("Parent missing")
			}

			if upconnecting {
				output = reorderComps(parent, SubtreeRootedAtS)

				log.Printf("Reordered Decomp: %v\n", output)
			} else {
				output = SubtreeRootedAtS
			}

			output.Up = make([]int, len(comp.Up))
			copy(output.Up, output.Up)
			output.Low = make([]int, len(comp.Low))
			copy(output.Low, comp.Low)

			return Decomp{Graph: d.Graph, Root: output, UpConnecting: comp.UpConnecting}

			decomposed = true
		}
	}

	log.Println("REJECT: Couldn't find a sep for ", comp)

	return Decomp{} // using empty decomp as reject case

}

func (d DivideKDecomp) FindDecomp(K int) Decomp {
	output := d.decomp(DivideComp{Edges: d.Graph.Edges})
	output.Root = postProcess(output.Root, []int{})

	return output
}

func (d DivideKDecomp) Name() string {
	return "DivideK"
}

type DivideKDecompPar struct {
	Graph     Graph
	K         int
	BalFactor int
}

func (d DivideKDecompPar) FindDecomp(K int) Decomp {

	div := DivideKDecomp{Graph: d.Graph, K: d.K, BalFactor: d.BalFactor}

	output := div.decompParallel(DivideComp{Edges: d.Graph.Edges})
	output.Root = postProcess(output.Root, []int{})

	return output
}

func (d DivideKDecompPar) Name() string {
	return "DivideK-Par"
}

func test3() {
	_, parseGraph := GetGraph("hypergraphs/grid2d_15.hg")

	e1 := parseGraph.GetEdge("e1(A,B)")
	e2 := parseGraph.GetEdge("e2(C,B)")
	e3 := parseGraph.GetEdge("e3(C,E)")
	e4 := parseGraph.GetEdge("e4(F,E)")
	edges := NewEdges([]Edge{e1, e2, e3, e4})

	spE1 := parseGraph.GetEdge("e5(A,C,D)")
	spE2 := parseGraph.GetEdge("e6(D,C,F)")
	spEdges := NewEdges([]Edge{spE1, spE2})

	sp := Special{Edges: spEdges, Vertices: spEdges.Vertices()}
	Sp := []Special{sp}

	component := Graph{Edges: edges}

	sep := NewEdges([]Edge{e3, e4})

	comp, compSp, _ := component.GetComponents(sep, Sp)

	for i := range comp {
		fmt.Println("Compnent: ", comp[i])
		fmt.Println("Special: ", compSp[i])
	}

	return
}
