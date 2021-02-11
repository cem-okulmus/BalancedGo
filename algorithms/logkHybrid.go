package algorithms

// Hybrid algorithm of log-k-decomp and det-k-decomp.

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// HybridPredicate is used to determine when to switch from LogKDecomp to using DetKDecomp
type HybridPredicate = func(H lib.Graph, K int) bool

type recursiveCall = func(H lib.Graph, Conn []int, allwowed lib.Edges) lib.Decomp

// LogKHybrid implements a hybridised algorithm, using LogKDecomp and DetKDecomp in tandem
type LogKHybrid struct {
	Graph     lib.Graph
	K         int
	cache     lib.Cache
	BalFactor int
	Predicate HybridPredicate // used to determine when to switch to DetK
	Size      int
}

// OneRoundPred will match the behaviour of BalDetK, with Depth 1
func (l *LogKHybrid) OneRoundPred(H lib.Graph, K int) bool {

	// log.Println("One Round Predicate")

	return true
}

// NumberEdgesPred checks the number of edges of the subgraph
func (l *LogKHybrid) NumberEdgesPred(H lib.Graph, K int) bool {

	output := H.Edges.Len() < l.Size

	if output {
		// log.Println("Predicate NumberEdgesPred")
		// log.Println("Current Graph: ", H.Edges.Len(), " Edges / ", l.Size)
	}

	return output
}

// SumEdgesPred checks the sum over all edges of the subgraph
func (l *LogKHybrid) SumEdgesPred(H lib.Graph, K int) bool {
	count := 0

	for i := range H.Edges.Slice() {
		count = count + len(H.Edges.Slice()[i].Vertices)
	}

	output := count < l.Size

	if output {
		// log.Println("Predicate SumEdgesPred")
		// log.Println("Current Graph: ", count, " Sum of Edges / ", l.Size)
	}

	return output

}

// ETimesKDivAvgEdgePred checks a complex formula over the subgraph and used K
func (l *LogKHybrid) ETimesKDivAvgEdgePred(H lib.Graph, K int) bool {

	count := 0

	for i := range H.Edges.Slice() {
		count = count + len(H.Edges.Slice()[i].Vertices)
	}

	avgEdgeSize := count / H.Edges.Len()

	output := ((H.Edges.Len() * l.K) / avgEdgeSize) < l.Size

	if output {
		// log.Println("Predicate ETimesKDivAvgEdgePred")
		// log.Println("Current Graph: ", ((H.Edges.Len() * l.K) / avgEdgeSize), "  / ", l.Size)
	}

	return output

}

// SetWidth sets the current width parameter of the algorithm
func (l *LogKHybrid) SetWidth(K int) {
	l.K = K
}

// Name returns the name of the algorithm
func (l *LogKHybrid) Name() string {
	return "LogKHybrid"
}

// FindDecomp finds a decomp
func (l *LogKHybrid) FindDecomp() lib.Decomp {
	// l.cache = make(map[uint32]*CompCache)
	l.cache.Init()
	return l.findDecomp(l.Graph, []int{}, l.Graph.Edges)
}

// FindDecompGraph finds a decomp, for an explicit graph
func (l *LogKHybrid) FindDecompGraph(Graph lib.Graph) lib.Decomp {
	l.Graph = Graph
	return l.FindDecomp()
}

func (l *LogKHybrid) detKWrapper(H lib.Graph, Conn []int, allwowed lib.Edges) lib.Decomp {

	det := DetKDecomp{K: l.K, Graph: lib.Graph{Edges: allwowed}, BalFactor: l.BalFactor, SubEdge: false}

	// TODO: reuse the same cache as for Logk?
	// det.Cache.Init()
	l.cache.CopyRef(&det.cache)

	return det.findDecomp(H, Conn)
}

// determine whether we have reached a (positive or negative) base case
func (l *LogKHybrid) baseCaseCheck(lenE int, lenSp int, lenAE int) bool {
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

func (l *LogKHybrid) baseCase(H lib.Graph, lenAE int) lib.Decomp {
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

func (l *LogKHybrid) findDecomp(H lib.Graph, Conn []int, allowedFull lib.Edges) lib.Decomp {

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

	// Deterime the function to use for the recursive calls
	var recCall recursiveCall

	if l.Predicate(H, l.K) {
		recCall = l.detKWrapper
	} else {
		recCall = l.findDecomp
	}

	//all vertices within (H ∪ Sp)
	verticesH := append(H.Vertices())

	allowed := lib.FilterVertices(allowedFull, verticesH)

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

		// log.Println("Balanced Child found, ", childλ)

		// Check if child is possible root
		if lib.Subset(Conn, childλ.Vertices()) {

			// log.Printf("Child-Root cover chosen: %v\n", Graph{Edges: childλ})
			// log.Printf("Comps of Child-Root: %v\n", comps_c)

			childχ := lib.Inter(childλ.Vertices(), verticesH)

			// check cache for previous encounters
			if l.cache.CheckNegative(childλ, comps_c) {
				// log.Println("Skipping a child sep", childχ)
				continue CHILD
			}

			var subtrees []lib.Node
			for y := range comps_c {
				V_comp_c := comps_c[y].Vertices()
				Conn_y := lib.Inter(V_comp_c, childχ)

				decomp := recCall(comps_c[y], Conn_y, allowedFull)
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

			// Check if parent is un-balanced
			for i := range comps_p {
				if comps_p[i].Len() > H.Len()/2 {
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
				fmt.Println("Child ", childλ, "  ", lib.PrintVertices(childλ.Vertices()))
				fmt.Println("Comps of child ", comps_c)
				fmt.Println("parent ", parentλ, "( ", parentalSearch.Result, " ) from the set: ", allowedParent)
				fmt.Println("Comps of p: ")
				for i := range comps_p {
					fmt.Println("Component: ", comps_p[i], " Len: ", comps_p[i].Len())

				}

				log.Panicln("the parallel search didn't actually find a valid parent")
			}

			vertCompLow := compLow.Vertices()
			childχ := lib.Inter(childλ.Vertices(), vertCompLow)

			// determine which components of child are inside comp_low

			//CHECK IF THIS ACTUALLY MAKES A DIFFERENCE
			comps_c, _, _ := compLow.GetComponents(childλ)

			//omitting the check for balancedness as it's guaranteed to still be conserved at this point

			// check cache for previous encounters
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

			chanUp := make(chan lib.Decomp)

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
			specialChild = lib.NewEdges([]lib.Edge{lib.Edge{Vertices: childχ}})

			// if no comps_p, other than comp_low, just use parent as is
			if len(comps_p) == 1 {
				compUp.Edges = parentλ

				// adding new Special Edge to connect Child to comp_up
				compUp.Special = append(compUp.Special, specialChild)

				decompTemp := lib.Decomp{Graph: compUp, Root: lib.Node{Bag: lib.Inter(parentλ.Vertices(), verticesH),
					Cover: parentλ, Children: []lib.Node{lib.Node{Bag: childχ, Cover: childλ}}}}

				go func(decomp lib.Decomp) {
					chanUp <- decomp
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
					chanUp <- recCall(comp_up, Conn, allowedReduced)
				}(compUp, Conn, allowedReduced)

			}

			// Parallel Recursive Calls:

			ch := make(chan decompInt)
			var subtrees []lib.Node

			for x := range comps_c {
				Conn_x := lib.Inter(comps_c[x].Vertices(), childχ)

				go func(x int, comps_c []lib.Graph, Conn_x []int, allowedFull lib.Edges) {
					var out decompInt
					out.Decomp = recCall(comps_c[x], Conn_x, allowedFull)
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

						// l.cache.AddNegative(childλ, comps_c[x])
						// log.Println("Rejecting child")
						continue PARENT
					}

					// log.Printf("Produced Decomp: %+v\n", decomp)
					subtrees = append(subtrees, decompInt.Decomp.Root)

				case decompUpChan := <-chanUp:

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
						fmt.Println("Child ", childλ, "  ", lib.PrintVertices(childλ.Vertices()))
						fmt.Println("Comps of child ", comps_c)
						fmt.Println("parent ", parentλ, "( ", parentalSearch.Result, " ) from the set: ", allowedParent)
						fmt.Println("comp_up ", compUp, " V(comp_up) ", lib.PrintVertices(compUp.Vertices()))
						fmt.Println("Decomp up:  ", decompUpChan)
						fmt.Println("Comps of p", comps_p)
						fmt.Println("Compare against PredSearch: ", predPar.Check(&H, &parentλ, l.BalFactor))

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
		// log.Println("Rejecting child ", childλ, " for H ", H)
		// log.Printf("\nCurrent SubGraph: %v\n", H)
		// log.Printf("Current Allowed Edges: %v\n", allowed)
		// log.Println("Conn: ", PrintVertices(Conn), "\n\n")
		// }
	}

	// exhausted search space
	return lib.Decomp{}
}
