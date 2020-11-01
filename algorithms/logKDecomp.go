// Parallel Algorithm for computing HD with log-depth recursion depth
package algorithms

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type LogKDecomp struct {
	Graph     Graph
	K         int
	cache     map[uint32]*CompCache
	cacheMux  sync.RWMutex
	BalFactor int
}

func (l *LogKDecomp) addPositive(sep []int, comp Graph, Sp []Special) {
	l.cacheMux.Lock()
	l.cache[IntHash(sep)].Succ = append(l.cache[IntHash(sep)].Succ, comp.Edges.HashExtended(Sp))
	l.cacheMux.Unlock()
}

func (l *LogKDecomp) addNegative(sep []int, comp Graph, Sp []Special) {
	l.cacheMux.Lock()
	l.cache[IntHash(sep)].Fail = append(l.cache[IntHash(sep)].Fail, comp.Edges.HashExtended(Sp))
	l.cacheMux.Unlock()
}

func (l *LogKDecomp) checkNegative(sep []int, comp Graph, Sp []Special) bool {
	l.cacheMux.RLock()
	defer l.cacheMux.RUnlock()

	compCachePrev, _ := l.cache[IntHash(sep)]
	for i := range compCachePrev.Fail {
		if comp.Edges.HashExtended(Sp) == compCachePrev.Fail[i] {
			log.Println("Comp ", comp, "(hash ", comp.Edges.Hash(), ")  known as negative for sep ", sep)
			return true
		}

	}

	return false
}

func (l *LogKDecomp) checkPositive(sep []int, comp Graph, Sp []Special) bool {
	l.cacheMux.RLock()
	defer l.cacheMux.RUnlock()

	compCachePrev, _ := l.cache[IntHash(sep)]
	for i := range compCachePrev.Fail {
		if comp.Edges.HashExtended(Sp) == compCachePrev.Succ[i] {
			//  log.Println("Comp ", comp, " known as negative for sep ", sep)
			return true
		}

	}

	return false
}

func (l LogKDecomp) Name() string {
	return "LogKDecomp"
}

// func (l LogKDecomp) FindDecompSeq(K int) Decomp {
// 	l.cache = make(map[uint32]*CompCache)
// 	return l.findHD(l.Graph, []Special{}, []int{}, l.Graph.Edges)
// }

func (l LogKDecomp) FindDecomp(K int) Decomp {
	l.cache = make(map[uint32]*CompCache)
	return l.findHDParallel(l.Graph, []Special{}, []int{}, l.Graph.Edges)
}

func (l LogKDecomp) FindDecompGraph(Graph Graph, K int) Decomp {
	l.Graph = Graph
	return l.FindDecomp(K)
}

// determine whether we have reached a (positive or negative) base case
func (l LogKDecomp) baseCaseCheck(lenE int, lenSp int, lenAE int) bool {
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

func (l LogKDecomp) baseCase(H Graph, Sp []Special, lenAE int) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var output Decomp

	// cover faiure cases

	if H.Edges.Len() == 0 && len(Sp) > 1 {
		return Decomp{}
	}
	if lenAE == 0 && (H.Edges.Len()+len(Sp)) >= 0 {
		return Decomp{}
	}

	// construct a decomp in the remaining two

	if H.Edges.Len() <= l.K && len(Sp) == 0 {
		output = Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges}}
	}
	if H.Edges.Len() == 0 && len(Sp) == 1 {
		sp1 := Sp[0]
		output = Decomp{Graph: H,
			Root: Node{Bag: sp1.Vertices, Cover: sp1.Edges}}

	}

	return output
}

//attach the two subtrees to form one
func (l LogKDecomp) attachingSubtrees(subtreeAbove Node, subtreeBelow Node, connecting Special) Node {
	log.Println("Two Nodes enter: ", subtreeAbove, subtreeBelow)
	log.Println("Connecting: ", PrintVertices(connecting.Vertices))
	// up = Inter(up, parent.Vertices())
	//finding connecting leaf in parent
	leaf := subtreeAbove.CombineNodes(subtreeBelow, connecting)

	if leaf == nil {
		fmt.Println("\n \n Connection ", PrintVertices(connecting.Vertices))
		fmt.Println("subtreeAbove ", subtreeAbove)

		log.Panicln("subtreeAbove doesn't contain connecting node!")
	}

	// //attaching subtree to parent
	// leaf.Children = []Node{subtree}
	// log.Println("Leaf ", leaf)
	log.Println("One Node leaves: ", leaf)
	return *leaf
}

func (l LogKDecomp) findHD(H Graph, Sp []Special, Conn []int, allowedFull Edges) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Allowed Edges: %v\n", allowedFull)
	log.Printf("Current Special Edges: %v\n", Sp)
	log.Println("Conn: ", PrintVertices(Conn), "\n\n")

	// Base Case
	if l.baseCaseCheck(H.Edges.Len(), len(Sp), allowedFull.Len()) {
		return l.baseCase(H, Sp, allowedFull.Len())
	}
	//all vertices within (H ∪ Sp)
	Vertices_H := append(H.Vertices(), VerticesSpecial(Sp)...)

	allowed := FilterVertices(allowedFull, Vertices_H)

	// Set up iterator for child
	genChild := GetCombin(allowed.Len(), l.K)

	// checks all possibles nodes in H, together with PARENT loops, it covers all parent-child pairings
CHILD:
	for genChild.HasNext() {

		childλ := GetSubset(allowed, genChild.Combination)
		genChild.Confirm()

		comps_c, compsSp_c, _, _ := H.GetComponents(childλ, Sp)

		// Check if child is balanced
		for i := range comps_c {
			if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
				continue CHILD
				// lowFound = true
				// comp_low_index = i //keep track of the index for composing comp_up later
				// comp_low = comps_p[i]
				// compSp_low = compsSp_p[i]
			}

		}

		log.Println("Balanced Child found, ", childλ)

		// Check if child is possible root
		if Subset(Conn, childλ.Vertices()) {
			// log.Println("parent breaks connectivity: ")
			// log.Println("Conn: ", PrintVertices(Conn))
			// log.Println("V(parentλ) = ", PrintVertices(parentλ.Vertices()))
			// continue PARENT

			log.Printf("Child-Root cover chosen: %v\n", Graph{Edges: childλ})
			log.Printf("Comps of Child-Root: %v\n", comps_c)

			childχ := Inter(childλ.Vertices(), Vertices_H)

			// check chache for previous encounters
			l.cacheMux.RLock()
			_, ok := l.cache[IntHash(childχ)]
			l.cacheMux.RUnlock()
			if !ok {
				var newCache CompCache
				l.cacheMux.Lock()
				l.cache[IntHash(childχ)] = &newCache
				l.cacheMux.Unlock()

			} else {
				for j := range comps_c {
					if l.checkNegative(childχ, comps_c[j], Sp) { //TODO: Add positive check and cutNodes
						log.Println("Skipping a child sep", childχ)
						continue CHILD
					}

				}
			}

			var subtrees []Node
			for y := range comps_c {
				V_comp_c := append(comps_c[y].Vertices(), VerticesSpecial(compsSp_c[y])...)
				Conn_y := Inter(V_comp_c, childχ)

				decomp := l.findHD(comps_c[y], compsSp_c[y], Conn_y, allowedFull)
				if reflect.DeepEqual(decomp, Decomp{}) {
					log.Println("Rejecting child-root")
					log.Printf("\nCurrent SubGraph: %v\n", H)
					log.Printf("Current Allowed Edges: %v\n", allowed)
					log.Printf("Current Special Edges: %v\n", Sp)
					log.Println("Conn: ", PrintVertices(Conn), "\n\n")
					l.addNegative(childχ, comps_c[y], compsSp_c[y])
					continue CHILD
				}

				// log.Printf("Produced Decomp w Child-Root: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)
			}

			root := Node{Bag: childχ, Cover: childλ, Children: subtrees}
			return Decomp{Graph: H, Root: root}
		}

		// //Connectivity Check Parent
		// vertCompLow := append(comp_low.Vertices(), VerticesSpecial(compSp_low)...)
		// vertCompLowProper := Diff(vertCompLow, parentλ.Vertices())

		// if !Subset(Inter(vertCompLow, Conn), parentλ.Vertices()) {
		// 	log.Println("Parent breaks connectivity wrt. comp_low", len(Inter(parentλ.Vertices(), Conn)))
		// 	continue PARENT
		// }

		// log.Printf("Parent cover chosen: %v\n", Graph{Edges: parentλ})

		// log.Printf("Comps of Parent: %v\n", comps_p)

		// log.Println("Lower comp: \n", comp_low)

		// Conn_child := Inter(vertCompLow, parentλ.Vertices())
		// bound := FilterVertices(allowed, Conn_child)

		// genChildCover := NewCover(l.K, Inter(vertCompLow, parentλ.Vertices()), bound, vertCompLow)

		// allowedInsideComplow := FilterVertices(allowed, vertCompLowProper)

		//Parent loops checks all possible parents that fit the chosen child node

		// Set up iterator for parent

		// parentMap := CreateOrderingMap(allowed, append(childλ.Vertices(), Conn...))

		// allowedParent := FilterVertices(allowed, append(childλ.Vertices(), Conn...))

		genParent := GetCombin(allowed.Len(), l.K)

	PARENT:
		for genParent.HasNext() {

			// parentλ := GetSubsetMap(allowed, genParent.Combination, parentMap)
			parentλ := GetSubset(allowed, genParent.Combination)
			// parentλ := GetSubset(allowedParent, genParent.Combination)
			genParent.Confirm()

			// 	var tryPartAlone bool

			// 	out := genChildCover.NextSubset()

			// 	if out == -1 {
			// 		if genChildCover.HasNext {
			// 			log.Panicln(" -1 but hasNext not false!")
			// 		}
			// 		continue CHILD_OUTER
			// 	}

			// 	childPart := GetSubset(bound, genChildCover.Subset)
			// 	if len(Inter(childPart.Vertices(), vertCompLowProper)) > 0 {
			// 		tryPartAlone = true
			// 	}

			// 	remainingAllowed := allowedInsideComplow.Diff(childPart)

			// 	genChild := GetCombin(remainingAllowed.Len(), l.K-childPart.Len())
			// CHILD_INNER:
			// 	for genChild.HasNext() || tryPartAlone {
			// 		if !genChild.HasNext() { // run once on just childPart if it extends into comp
			// 			tryPartAlone = false
			// 		}

			// 		var childλ Edges
			// 		if !genChild.HasNext() {
			// 			childλ = childPart
			// 		} else {
			// 			childAdded := GetSubset(remainingAllowed, genChild.Combination)
			// 			childλ = NewEdges(append(childAdded.Slice(), childPart.Slice()...))
			// 		}
			// 		genChild.Confirm()

			// 		childχ := Inter(childλ.Vertices(), vertCompLow)

			// 		// Connectivity check
			// 		if !Subset(Inter(vertCompLow, parentλ.Vertices()), childχ) {
			// 			// log.Println("Child ", childλ, " breaks connectivity!")
			// 			log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
			// 			log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))

			// 			log.Println("Child_part ", childPart)

			// 			log.Println("Child Full ", childλ)

			// 			log.Panicln("new loop not working, connectivity broken")
			// 			continue CHILD_INNER
			// 		}

			// // Check if progress made (TODO: implement this properly with minimal covers + extensions !)
			// if Subset(childχ, parentλ.Vertices()) {
			//  // log.Println("No progress made!")

			//  log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
			//  log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))

			//  log.Println("Child_part ", childPart)

			//  log.Println("Child Full ", childλ)

			//  log.Panicln("New loop not working, no progress made")
			//  continue CHILD_INNER
			// }

			comps_p, compsSp_p, _, isolatedEdges := H.GetComponents(parentλ, Sp)

			foundLow := false
			var comp_low_index int
			var comp_low Graph
			var compSp_low []Special

			// Check if parent is un-balanced
			for i := range comps_p {
				if comps_p[i].Edges.Len()+len(compsSp_p[i]) > (H.Edges.Len()+len(Sp))/2 {
					foundLow = true
					comp_low_index = i //keep track of the index for composing comp_up later
					comp_low = comps_p[i]
					compSp_low = compsSp_p[i]
				}
			}
			if !foundLow {
				continue PARENT
			}

			vertCompLow := append(comp_low.Vertices(), VerticesSpecial(compSp_low)...)
			childχ := Inter(childλ.Vertices(), vertCompLow)

			// Connectivity checks

			if !Subset(Inter(vertCompLow, Conn), parentλ.Vertices()) {
				// log.Println("Conn not covered by parent")

				// log.Println("Conn: ", PrintVertices(Conn))
				// log.Println("V(parentλ) \\cap Conn", PrintVertices(Inter(parentλ.Vertices(), Conn)))
				// log.Println("V(Comp_low) \\cap Conn ", PrintVertices(Inter(vertCompLow, Conn)))
				continue PARENT
			}

			// Connectivity check
			if !Subset(Inter(vertCompLow, parentλ.Vertices()), childχ) {
				// log.Println("Child not connected to parent!")
				// log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
				// log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))

				// log.Println("Child", childλ)

				continue PARENT
			}

			// determine which componenents of child are inside comp_low

			//CHECK IF THIS ACTUALLY MAKES A DIFFERENCE
			comps_c, compsSp_c, _, _ := comp_low.GetComponents(childλ, compSp_low)

			//omitting the check for balancedness as it's guaranteed to still be conserved at this point

			// check chache for previous encounters
			l.cacheMux.RLock()
			_, ok := l.cache[IntHash(childχ)]
			l.cacheMux.RUnlock()
			if !ok {
				var newCache CompCache
				l.cacheMux.Lock()
				l.cache[IntHash(childχ)] = &newCache
				l.cacheMux.Unlock()

			} else {
				for j := range comps_c {
					if l.checkNegative(childχ, comps_c[j], Sp) { //TODO: Add positive check and cutNodes
						// log.Println("Skipping a child sep", childχ)
						continue PARENT
					}
				}
			}

			// log.Println("Size of H' : ", H.Edges.Len()+len(Sp))

			// // Check childχ is balanced separator
			// for i := range comps_c {
			// 	if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
			// 		log.Println("Child is not a bal. sep. !")
			// 		continue CHILD_INNER
			// 	}
			// }

			log.Printf("Parent Found: %v (%s) \n", Graph{Edges: parentλ}, PrintVertices(parentλ.Vertices()))

			log.Printf("Child chosen: %v (%s) \n", Graph{Edges: childλ}, PrintVertices(childχ))
			log.Printf("Comps of Child: %v\n", comps_c)

			//Computing subcomponents of Child

			var subtrees []Node
			for x := range comps_c {
				Conn_x := Inter(append(comps_c[x].Vertices(), VerticesSpecial(compsSp_c[x])...), childχ)

				decomp := l.findHD(comps_c[x], compsSp_c[x], Conn_x, allowedFull)
				if reflect.DeepEqual(decomp, Decomp{}) {
					l.addNegative(childχ, comps_c[x], compsSp_c[x])
					log.Println("Rejecting child")
					continue PARENT
				}

				log.Printf("Produced Decomp: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)

			}

			//Computing upper component

			var comp_up Graph
			var decompUp Decomp
			var specialChild Special
			var compSp_up []Special
			tempEdgeSlice := []Edge{}

			tempEdgeSlice = append(tempEdgeSlice, isolatedEdges...)
			for i := range comps_p {
				if i != comp_low_index {
					tempEdgeSlice = append(tempEdgeSlice, comps_p[i].Edges.Slice()...)
					compSp_up = append(compSp_up, compsSp_p[i]...)
				}
			}

			// if no comps_p, other than comp_low, just use parent as is
			if len(comps_p) == 1 {
				comp_up.Edges = parentλ

				// adding new Special Edge to connect Child to comp_up
				specialChild = Special{Vertices: childχ, Edges: childλ}

				decompUp = Decomp{Graph: comp_up, Root: Node{Bag: Inter(parentλ.Vertices(), Vertices_H),
					Cover: parentλ, Children: []Node{Node{Bag: childχ, Cover: childλ}}}}

				if !Subset(Conn, parentλ.Vertices()) {

					log.Println("Comps of p", comps_p)

					log.Panicln("Conn not covered in parent, Wait, what?")
				}

			} else if len(tempEdgeSlice) > 0 { // otherwise compute decomp for comp_up

				comp_up.Edges = NewEdges(tempEdgeSlice)

				log.Println("Upper component:", comp_up)

				//Reducing the allowed edges
				allowedReduced := allowedFull.Diff(comp_low.Edges)

				// adding new Special Edge to connect Child to comp_up
				specialChild = Special{Vertices: childχ, Edges: childλ}
				compSp_up = append(compSp_up, specialChild)

				decompUp = l.findHD(comp_up, compSp_up, Conn, allowedReduced)

				if reflect.DeepEqual(decompUp, Decomp{}) {

					// l.addNegative(childχ, comp_up, compSp_up) TODO think about how to cache this
					log.Println("Rejecting comp_up")

					continue PARENT
				}

			}

			// rearrange subtrees to form one that covers total of H
			rootChild := Node{Bag: childχ, Cover: childλ, Children: subtrees}

			var finalRoot Node
			if len(tempEdgeSlice) > 0 {
				finalRoot = l.attachingSubtrees(decompUp.Root, rootChild, specialChild)
			} else {
				finalRoot = rootChild
			}

			log.Printf("Produced Decomp: %v\n", finalRoot)
			return Decomp{Graph: H, Root: finalRoot}

		}
		log.Println("ran out of parents,\n\n")
	}

	// exhausted search space
	return Decomp{}
}

// a parallel version of logK, using goroutines and channels to run various parts concurrently
func (l LogKDecomp) findHDParallel(H Graph, Sp []Special, Conn []int, allowedFull Edges) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Allowed Edges: %v\n", allowedFull)
	log.Printf("Current Special Edges: %v\n", Sp)
	log.Println("Conn: ", PrintVertices(Conn), "\n\n")

	// Base Case
	if l.baseCaseCheck(H.Edges.Len(), len(Sp), allowedFull.Len()) {
		return l.baseCase(H, Sp, allowedFull.Len())
	}
	//all vertices within (H ∪ Sp)
	Vertices_H := append(H.Vertices(), VerticesSpecial(Sp)...)

	allowed := FilterVertices(allowedFull, Vertices_H)

	// Set up iterator for child

	genChild := SplitCombin(allowed.Len(), l.K, runtime.GOMAXPROCS(-1), false)

	parallelSearch := Search{H: &H, Sp: Sp, Edges: &allowed, BalFactor: l.BalFactor, Generators: genChild}

	pred := BalancedCheck{}

	parallelSearch.FindNext(pred) // initial Search

	// genChild := GetCombin(allowed.Len(), l.K)

	// checks all possibles nodes in H, together with PARENT loops, it covers all parent-child pairings
CHILD:
	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {

		// var found []int

		// parallelSearch(H, Sp, allowed, &found, genChild, l.BalFactor)

		// if len(found) == 0 { // meaning that the search above never found anything
		// 	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
		// 	genChild_hasNext = false
		// 	continue
		// }

		childλ := GetSubset(allowed, parallelSearch.Result)

		comps_c, compsSp_c, _, _ := H.GetComponents(childλ, Sp)

		// // Check if child is balanced
		// for i := range comps_c {
		// 	if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
		// 		continue CHILD
		// 		// lowFound = true
		// 		// comp_low_index = i //keep track of the index for composing comp_up later
		// 		// comp_low = comps_p[i]
		// 		// compSp_low = compsSp_p[i]
		// 	}

		// }

		log.Println("Balanced Child found, ", childλ)

		// Check if child is possible root
		if Subset(Conn, childλ.Vertices()) {

			log.Printf("Child-Root cover chosen: %v\n", Graph{Edges: childλ})
			log.Printf("Comps of Child-Root: %v\n", comps_c)

			childχ := Inter(childλ.Vertices(), Vertices_H)

			// check chache for previous encounters
			l.cacheMux.RLock()
			_, ok := l.cache[IntHash(childχ)]
			l.cacheMux.RUnlock()
			if !ok {
				var newCache CompCache
				l.cacheMux.Lock()
				l.cache[IntHash(childχ)] = &newCache
				l.cacheMux.Unlock()

			} else {
				for j := range comps_c {
					if l.checkNegative(childχ, comps_c[j], Sp) { //TODO: Add positive check and cutNodes
						log.Println("Skipping a child sep", childχ)
						continue CHILD
					}

				}
			}

			var subtrees []Node
			for y := range comps_c {
				V_comp_c := append(comps_c[y].Vertices(), VerticesSpecial(compsSp_c[y])...)
				Conn_y := Inter(V_comp_c, childχ)

				decomp := l.findHDParallel(comps_c[y], compsSp_c[y], Conn_y, allowedFull)
				if reflect.DeepEqual(decomp, Decomp{}) {
					log.Println("Rejecting child-root")
					log.Printf("\nCurrent SubGraph: %v\n", H)
					log.Printf("Current Allowed Edges: %v\n", allowed)
					log.Printf("Current Special Edges: %v\n", Sp)
					log.Println("Conn: ", PrintVertices(Conn), "\n\n")
					l.addNegative(childχ, comps_c[y], compsSp_c[y])
					continue CHILD
				}

				// log.Printf("Produced Decomp w Child-Root: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)
			}

			root := Node{Bag: childχ, Cover: childλ, Children: subtrees}
			return Decomp{Graph: H, Root: root}
		}

		// genParent := GetCombin(allowed.Len(), l.K)

		genParent := SplitCombin(allowed.Len(), l.K, runtime.GOMAXPROCS(-1), false)

		parentalSearch := Search{H: &H, Sp: Sp, Edges: &allowed, BalFactor: l.BalFactor, Generators: genParent}

		predPar := ParentCheck{Conn: Conn, Child: childλ.Vertices()}

		parentalSearch.FindNext(predPar)

	PARENT:
		for ; !parentalSearch.ExhaustedSearch; parentalSearch.FindNext(predPar) {

			// var foundParent []int

			// parallelSearch(H, Sp, allowed, &foundParent, genParent, l.BalFactor)

			// if len(foundParent) == 0 { // meaning that the search above never found anything
			// 	log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			// 	genChild_hasNext = false
			// 	continue
			// }

			// parentλ := GetSubsetMap(allowed, genParent.Combination, parentMap)
			parentλ := GetSubset(allowed, parentalSearch.Result)
			// parentλ := GetSubset(allowedParent, genParent.Combination)

			log.Println("Looking at parent ", parentλ)

			comps_p, compsSp_p, _, isolatedEdges := H.GetComponents(parentλ, Sp)

			foundLow := false
			var comp_low_index int
			var comp_low Graph
			var compSp_low []Special

			// Check if parent is un-balanced
			for i := range comps_p {
				if comps_p[i].Edges.Len()+len(compsSp_p[i]) > (H.Edges.Len()+len(Sp))/2 {
					foundLow = true
					comp_low_index = i //keep track of the index for composing comp_up later
					comp_low = comps_p[i]
					compSp_low = compsSp_p[i]
				}
			}
			if !foundLow {
				continue PARENT
			}

			vertCompLow := append(comp_low.Vertices(), VerticesSpecial(compSp_low)...)
			childχ := Inter(childλ.Vertices(), vertCompLow)

			// determine which componenents of child are inside comp_low

			//CHECK IF THIS ACTUALLY MAKES A DIFFERENCE
			comps_c, compsSp_c, _, _ := comp_low.GetComponents(childλ, compSp_low)

			//omitting the check for balancedness as it's guaranteed to still be conserved at this point

			// check chache for previous encounters
			l.cacheMux.RLock()
			_, ok := l.cache[IntHash(childχ)]
			l.cacheMux.RUnlock()
			if !ok {
				var newCache CompCache
				l.cacheMux.Lock()
				l.cache[IntHash(childχ)] = &newCache
				l.cacheMux.Unlock()

			} else {
				for j := range comps_c {
					if l.checkNegative(childχ, comps_c[j], Sp) { //TODO: Add positive check and cutNodes
						// log.Println("Skipping a child sep", childχ)
						continue PARENT
					}
				}
			}

			// log.Println("Size of H' : ", H.Edges.Len()+len(Sp))

			// // Check childχ is balanced separator
			// for i := range comps_c {
			// 	if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
			// 		log.Println("Child is not a bal. sep. !")
			// 		continue CHILD_INNER
			// 	}
			// }

			log.Printf("Parent Found: %v (%s) \n", Graph{Edges: parentλ}, PrintVertices(parentλ.Vertices()))

			log.Printf("Child chosen: %v (%s) \n", Graph{Edges: childλ}, PrintVertices(childχ))
			log.Printf("Comps of Child: %v\n", comps_c)

			//Computing subcomponents of Child

			var subtrees []Node
			for x := range comps_c {
				Conn_x := Inter(append(comps_c[x].Vertices(), VerticesSpecial(compsSp_c[x])...), childχ)

				decomp := l.findHDParallel(comps_c[x], compsSp_c[x], Conn_x, allowedFull)
				if reflect.DeepEqual(decomp, Decomp{}) {
					l.addNegative(childχ, comps_c[x], compsSp_c[x])
					log.Println("Rejecting child")
					continue PARENT
				}

				log.Printf("Produced Decomp: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)

			}

			//Computing upper component

			var comp_up Graph
			var decompUp Decomp
			var specialChild Special
			var compSp_up []Special
			tempEdgeSlice := []Edge{}

			tempEdgeSlice = append(tempEdgeSlice, isolatedEdges...)
			for i := range comps_p {
				if i != comp_low_index {
					tempEdgeSlice = append(tempEdgeSlice, comps_p[i].Edges.Slice()...)
					compSp_up = append(compSp_up, compsSp_p[i]...)
				}
			}

			// if no comps_p, other than comp_low, just use parent as is
			if len(comps_p) == 1 {
				comp_up.Edges = parentλ

				// adding new Special Edge to connect Child to comp_up
				specialChild = Special{Vertices: childχ, Edges: childλ}

				decompUp = Decomp{Graph: comp_up, Root: Node{Bag: Inter(parentλ.Vertices(), Vertices_H),
					Cover: parentλ, Children: []Node{Node{Bag: childχ, Cover: childλ}}}}

				if !Subset(Conn, parentλ.Vertices()) {

					log.Println("Comps of p", comps_p)

					log.Panicln("Conn not covered in parent, Wait, what?")
				}

			} else if len(tempEdgeSlice) > 0 { // otherwise compute decomp for comp_up

				comp_up.Edges = NewEdges(tempEdgeSlice)

				log.Println("Upper component:", comp_up)

				//Reducing the allowed edges
				allowedReduced := allowedFull.Diff(comp_low.Edges)

				// adding new Special Edge to connect Child to comp_up
				specialChild = Special{Vertices: childχ, Edges: childλ}

				decompUp = l.findHDParallel(comp_up, append(compSp_up, specialChild), Conn, allowedReduced)

				if reflect.DeepEqual(decompUp, Decomp{}) {

					// l.addNegative(childχ, comp_up, Sp)
					log.Println("Rejecting comp_up")

					continue PARENT
				}

			}

			// rearrange subtrees to form one that covers total of H
			rootChild := Node{Bag: childχ, Cover: childλ, Children: subtrees}

			var finalRoot Node
			if len(tempEdgeSlice) > 0 {
				finalRoot = l.attachingSubtrees(decompUp.Root, rootChild, specialChild)
			} else {
				finalRoot = rootChild
			}

			log.Printf("Produced Decomp: %v\n", finalRoot)
			return Decomp{Graph: H, Root: finalRoot}

		}
		log.Println("ran out of parents", parentalSearch.ExhaustedSearch, "\n\n")
	}

	// exhausted search space
	return Decomp{}

}
