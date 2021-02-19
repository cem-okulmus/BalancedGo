package algorithms

// Parallel Algorithm for computing HD with log-depth recursion depth

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// LogKDecomp implements a parallel log-depth HD algorithm
type LogKDecomp struct {
	Graph     lib.Graph
	K         int
	cache     lib.Cache
	BalFactor int
}

// decompInt is used to keep track of returned decompositions during concurrent search
type decompInt struct {
	Decomp lib.Decomp
	Int    int
}

// SetWidth sets the current width parameter of the algorithm
func (l *LogKDecomp) SetWidth(K int) {
	l.cache.Reset() // reset the cache as the new width might invalidate any old results

	l.K = K
}

// Name returns the name of the algorithm
func (l *LogKDecomp) Name() string {
	return "LogKDecomp"
}

// FindDecomp finds a decomp
func (l *LogKDecomp) FindDecomp() lib.Decomp {
	l.cache.Init()
	return l.findDecomp(l.Graph, []int{}, l.Graph.Edges)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (l *LogKDecomp) FindDecompGraph(Graph lib.Graph) lib.Decomp {
	l.Graph = Graph
	return l.FindDecomp()
}

// determine whether we have reached a (positive or negative) base case
func (l *LogKDecomp) baseCaseCheck(lenE int, lenSp int, lenAE int) bool {
	if lenE <= l.K && lenSp == 0 {
		return true
	}
	if lenE == 0 && lenSp == 1 {
		return true
	}
	if lenE == 0 && lenSp > 1 {
		return true
	}
	if lenAE == 0 && (lenE+lenSp) >= 0 {
		return true
	}

	return false
}

func (l *LogKDecomp) baseCase(H lib.Graph, lenAE int) lib.Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output lib.Decomp

	// cover faiure cases
	if H.Edges.Len() == 0 && len(H.Special) > 1 {
		return lib.Decomp{}
	}
	if lenAE == 0 && (H.Len()) >= 0 {
		return lib.Decomp{}
	}

	// construct a decomp in the remaining two
	if H.Edges.Len() <= l.K && len(H.Special) == 0 {
		output = lib.Decomp{Graph: H, Root: lib.Node{Bag: H.Vertices(), Cover: H.Edges}}
	}
	if H.Edges.Len() == 0 && len(H.Special) == 1 {
		sp1 := H.Special[0]
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: sp1.Vertices(), Cover: sp1}}

	}

	return output
}

//attach the two subtrees to form one
func attachingSubtrees(subtreeAbove lib.Node, subtreeBelow lib.Node, connecting lib.Edges) lib.Node {
	// log.Println("Two Nodes enter: ", subtreeAbove, subtreeBelow)
	// log.Println("Connecting: ", PrintVertices(connecting.Vertices))

	//finding connecting leaf in parent
	leaf := subtreeAbove.CombineNodes(subtreeBelow, connecting)

	if leaf == nil {
		fmt.Println("\n \n Connection ", lib.PrintVertices(connecting.Vertices()))
		fmt.Println("subtreeAbove ", subtreeAbove)

		log.Panicln("subtreeAbove doesn't contain connecting node!")
	}

	return *leaf
}

func (l *LogKDecomp) findDecomp(H lib.Graph, Conn []int, allowedFull lib.Edges) lib.Decomp {

	// log.Printf("\n\nCurrent SubGraph: %v\n", H)
	// log.Printf("Current Allowed Edges: %v\n", allowedFull)
	// log.Println("Conn: ", PrintVertices(Conn), "\n\n")

	if !lib.Subset(Conn, H.Vertices()) {
		log.Panicln("You done fucked up. ")
	}

	// Base Case
	if l.baseCaseCheck(H.Edges.Len(), len(H.Special), allowedFull.Len()) {
		return l.baseCase(H, allowedFull.Len())
	}
	//all vertices within (H ∪ Sp)
	VerticesH := append(H.Vertices())

	allowed := lib.FilterVertices(allowedFull, VerticesH)

	// Set up iterator for child

	genChild := lib.SplitCombin(allowed.Len(), l.K, runtime.GOMAXPROCS(-1), false)
	parallelSearch := lib.Search{H: &H, Edges: &allowed, BalFactor: l.BalFactor, Generators: genChild}
	pred := lib.BalancedCheck{}
	parallelSearch.FindNext(pred) // initial Search

	// checks all possibles nodes in H, together with PARENT loops, it covers all parent-child pairings
CHILD:
	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {

		childλ := lib.GetSubset(allowed, parallelSearch.Result)
		comps_c, _, _ := H.GetComponents(childλ)

		// log.Println("Balanced Child found, ", childλ, "of H ", H)

		// Check if child is possible root
		if lib.Subset(Conn, childλ.Vertices()) {
			// log.Printf("Child-Root cover chosen: %v of %v \n", childλ, H)
			// log.Printf("Comps of Child-Root: %v\n", comps_c)

			childχ := lib.Inter(childλ.Vertices(), VerticesH)

			// check cache for previous encounters
			if l.cache.CheckNegative(childλ, comps_c) {
				// log.Println("Skipping a child sep", childχ)
				continue CHILD
			}

			var subtrees []lib.Node
			for y := range comps_c {
				V_comp_c := comps_c[y].Vertices()
				Conn_y := lib.Inter(V_comp_c, childχ)

				decomp := l.findDecomp(comps_c[y], Conn_y, allowedFull)
				if reflect.DeepEqual(decomp, lib.Decomp{}) {
					// log.Println("Rejecting child-root")
					// log.Printf("\nCurrent SubGraph: %v\n", H)
					// log.Printf("Current Allowed Edges: %v\n", allowed)
					// log.Println("Conn: ", PrintVertices(Conn), "\n\n")
					l.cache.AddNegative(childλ, comps_c[y])
					continue CHILD
				}

				// log.Printf("Produced Decomp w Child-Root: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)
			}

			root := lib.Node{Bag: childχ, Cover: childλ, Children: subtrees}
			return lib.Decomp{Graph: H, Root: root}
		}

		// Set up iterator for parent
		allowedParent := lib.FilterVertices(allowed, append(Conn, childλ.Vertices()...))
		genParent := lib.SplitCombin(allowedParent.Len(), l.K, runtime.GOMAXPROCS(-1), false)
		parentalSearch := lib.Search{H: &H, Edges: &allowedParent, BalFactor: l.BalFactor, Generators: genParent}
		predPar := lib.ParentCheck{Conn: Conn, Child: childλ.Vertices()}
		parentalSearch.FindNext(predPar)
		// parentFound := false
	PARENT:
		for ; !parentalSearch.ExhaustedSearch; parentalSearch.FindNext(predPar) {

			parentλ := lib.GetSubset(allowedParent, parentalSearch.Result)

			// log.Println("Looking at parent ", parentλ)

			comps_p, _, isolatedEdges := H.GetComponents(parentλ)

			// log.Println("Parent components ", comps_p)

			foundLow := false
			var compLowIndex int
			var compLow lib.Graph

			balancednessLimit := (((H.Len()) * (l.BalFactor - 1)) / l.BalFactor)

			// Check if parent is un-balanced
			for i := range comps_p {
				if comps_p[i].Len() > balancednessLimit {
					foundLow = true
					compLowIndex = i //keep track of the index for composing comp_up later
					compLow = comps_p[i]
				}
			}
			if !foundLow {
				fmt.Println("Current SubGraph, ", H)
				fmt.Println("Conn ", lib.PrintVertices(Conn))

				fmt.Printf("Current Allowed Edges: %v\n", allowed)
				fmt.Printf("Current Allowed Edges in Parent Search: %v\n", parentalSearch.Edges)

				fmt.Println("Child ", childλ)
				fmt.Println("Comps of child ", comps_c)
				fmt.Println("parent ", parentλ, " ( ", parentalSearch.Result, ")")

				fmt.Println("Comps of p: ")
				for i := range comps_p {
					fmt.Println("Component: ", comps_p[i], " Len: ", comps_p[i].Len())

				}

				log.Panicln("the parallel search didn't actually find a valid parent")
			}

			vertCompLow := compLow.Vertices()
			childχ := lib.Inter(childλ.Vertices(), vertCompLow)

			// determine which componenents of child are inside comp_low

			//CHECK IF THIS ACTUALLY MAKES A DIFFERENCE
			comps_c, _, _ := compLow.GetComponents(childλ)

			//omitting the check for balancedness as it's guaranteed to still be conserved at this point

			// check chache for previous encounters
			if l.cache.CheckNegative(childλ, comps_c) {
				// log.Println("Skipping a child sep", childχ)
				continue PARENT
			}

			// log.Printf("Parent Found: %v (%s) \n", parentλ, PrintVertices(parentλ.Vertices()))
			// parentFound = true
			// log.Println("Comp low: ", comp_low, "Vertices of comp_low", PrintVertices(vertCompLow))
			// log.Printf("Child chosen: %v (%s) for H %v \n", childλ, PrintVertices(childχ), H)
			// log.Printf("Comps of Child: %v\n", comps_c)

			//Computing subcomponents of Child

			// 1. CREATE GOROUTINES
			// ---------------------

			//Computing upper component in parallel

			chUp := make(chan lib.Decomp)

			var compUp lib.Graph
			var decompUp lib.Decomp
			var specialChild lib.Edges
			tempEdgeSlice := []lib.Edge{}
			tempSpecialSlice := []lib.Edges{}

			tempEdgeSlice = append(tempEdgeSlice, isolatedEdges...)
			for i := range comps_p {
				if i != compLowIndex {
					tempEdgeSlice = append(tempEdgeSlice, comps_p[i].Edges.Slice()...)
					tempSpecialSlice = append(tempSpecialSlice, comps_p[i].Special...)
				}
			}

			// specialChild = NewEdges([]Edge{Edge{Vertices: Inter(childχ, comp_up.Vertices())}})
			specialChild = lib.NewEdges([]lib.Edge{{Vertices: childχ}})

			// if no comps_p, other than comp_low, just use parent as is
			if len(comps_p) == 1 {
				compUp.Edges = parentλ

				// adding new Special Edge to connect Child to comp_up
				compUp.Special = append(compUp.Special, specialChild)

				decompTemp := lib.Decomp{Graph: compUp, Root: lib.Node{Bag: lib.Inter(parentλ.Vertices(), VerticesH),
					Cover: parentλ, Children: []lib.Node{{Bag: specialChild.Vertices(), Cover: childλ}}}}

				go func(decomp lib.Decomp) {
					chUp <- decomp
				}(decompTemp)

			} else if len(tempEdgeSlice) > 0 { // otherwise compute decomp for comp_up

				compUp.Edges = lib.NewEdges(tempEdgeSlice)
				compUp.Special = tempSpecialSlice

				// adding new Special Edge to connect Child to comp_up
				compUp.Special = append(compUp.Special, specialChild)

				// log.Println("Upper component:", comp_up)

				//Reducing the allowed edges
				allowedReduced := allowedFull.Diff(compLow.Edges)

				go func(comp_up lib.Graph, Conn []int, allowedReduced lib.Edges) {
					chUp <- l.findDecomp(comp_up, Conn, allowedReduced)
				}(compUp, Conn, allowedReduced)

			}

			// Parallel Recursive Calls:

			ch := make(chan decompInt)
			var subtrees []lib.Node

			for x := range comps_c {
				Conn_x := lib.Inter(comps_c[x].Vertices(), childχ)

				go func(x int, comps_c []lib.Graph, Conn_x []int, allowedFull lib.Edges) {
					var out decompInt
					out.Decomp = l.findDecomp(comps_c[x], Conn_x, allowedFull)
					out.Int = x
					ch <- out
				}(x, comps_c, Conn_x, allowedFull)

			}

			// 2. WAIT ON GOROUTINES TO FINISH
			// ---------------------

			for i := 0; i < len(comps_c)+1; i++ {
				select {
				case decompInt := <-ch:

					if reflect.DeepEqual(decompInt.Decomp, lib.Decomp{}) {

						l.cache.AddNegative(childλ, comps_c[decompInt.Int])
						// log.Println("Rejecting child")
						continue PARENT
					}

					// log.Printf("Produced Decomp: %+v\n", decomp)
					subtrees = append(subtrees, decompInt.Decomp.Root)

				case decompUpChan := <-chUp:

					if reflect.DeepEqual(decompUpChan, lib.Decomp{}) {

						// l.addNegative(childχ, comp_up, Sp)
						// log.Println("Rejecting comp_up ", comp_up, " of H ", H)

						continue PARENT
					}

					if !lib.Subset(Conn, decompUpChan.Root.Bag) {
						fmt.Println("Current SubGraph, ", H)
						fmt.Println("Conn ", lib.PrintVertices(Conn))

						fmt.Printf("Current Allowed Edges: %v\n", allowed)
						fmt.Printf("Current Allowed Edges in Parent Search: %v\n", parentalSearch.Edges)

						fmt.Println("Child ", childλ, "  ", lib.PrintVertices(childχ))
						fmt.Println("Comps of child ", comps_c)
						fmt.Println("parent ", parentλ, " ( ", parentalSearch.Result, ") Vertices(parent) ", lib.PrintVertices(parentλ.Vertices()))

						fmt.Println("comp_up ", compUp, " V(comp_up) ", lib.PrintVertices(compUp.Vertices()))

						fmt.Println("Decomp up:  ", decompUpChan)

						fmt.Println("Comps of p", comps_p)

						fmt.Println("Compare against PredSearch: ")

						log.Panicln("Conn not covered in parent, Wait, what?")
					}

					decompUp = decompUpChan

				}

			}

			// 3. POST-PROCESSING (sequentially)
			// ---------------------

			// rearrange subtrees to form one that covers total of H
			rootChild := lib.Node{Bag: childχ, Cover: childλ, Children: subtrees}

			var finalRoot lib.Node
			if len(tempEdgeSlice) > 0 {
				finalRoot = attachingSubtrees(decompUp.Root, rootChild, specialChild)
			} else {
				finalRoot = rootChild
			}

			// log.Printf("Produced Decomp: %v\n", finalRoot)
			return lib.Decomp{Graph: H, Root: finalRoot}

		}
		// if parentFound {
		// 	log.Println("Rejecting child ", childλ, " for H ", H)
		// 	log.Printf("\nCurrent SubGraph: %v\n", H)
		// 	log.Printf("Current Allowed Edges: %v\n", allowed)
		// 	log.Println("Conn: ", PrintVertices(Conn), "\n\n")
		// }

	}

	// exhausted search space
	return lib.Decomp{}
}
