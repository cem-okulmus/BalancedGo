// Parallel Algorithm for computing HD with log-depth recursion depth
package algorithms

import (
	"fmt"
	"log"
	"reflect"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type LogKDecomp struct {
	Graph Graph
	K     int
}

func (l LogKDecomp) Name() string {
	return "LogKDecomp"
}

func (l LogKDecomp) FindDecomp(K int) Decomp {
	return l.findHD(l.Graph, []Special{}, []int{}, l.Graph.Edges)
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

func (l LogKDecomp) findHD(H Graph, Sp []Special, Conn []int, allowed Edges) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n", Sp)
	log.Println("Conn: ", PrintVertices(Conn), "\n\n")

	// Base Case
	if l.baseCaseCheck(H.Edges.Len(), len(Sp), allowed.Len()) {
		return l.baseCase(H, Sp, allowed.Len())
	}

	// Set up iterator for parent
	genParent := GetCombin(allowed.Len(), l.K)

	// checks all possibles nodes in H, together with Child loops, it covers all parent-child pairings
PARENT:
	for genParent.HasNext() {

		parentλ := GetSubset(allowed, genParent.Combination)
		genParent.Confirm()

		comps_p, compsSp_p, _, isolatedEdges := H.GetComponents(parentλ, Sp)

		var comp_low_index int
		var comp_low Graph
		var compSp_low []Special

		//all vertices within (H ∪ Sp)
		Vertices_H := append(H.Vertices(), VerticesSpecial(Sp)...)

		lowFound := false

		// Check if lower component present
		for i := range comps_p {
			if comps_p[i].Edges.Len()+len(compsSp_p[i]) > (H.Edges.Len()+len(Sp))/2 {
				lowFound = true
				comp_low_index = i //keep track of the index for composing comp_up later
				comp_low = comps_p[i]
				compSp_low = compsSp_p[i]
			}

		}

		// Check if Parent is possible root
		if !lowFound {

			// Connectivity Check
			if !Subset(Conn, parentλ.Vertices()) {
				// log.Println("parent breaks connectivity: ")
				// log.Println("Conn: ", PrintVertices(Conn))
				// log.Println("V(parentλ) = ", PrintVertices(parentλ.Vertices()))
				continue PARENT
			}

			log.Printf("Parent cover chosen: %v\n", Graph{Edges: parentλ})

			log.Printf("Comps of Parent: %v\n", comps_p)
			log.Println("parentλ is bal. sep. !")

			var subtrees []Node
			for y := range comps_p {
				Conn_y := Inter(append(comps_p[y].Vertices(), VerticesSpecial(compsSp_p[y])...), parentλ.Vertices())

				decomp := l.findHD(comps_p[y], compsSp_p[y], Conn_y, allowed)
				if reflect.DeepEqual(decomp, Decomp{}) {
					log.Println("Rejecting parent as root")
					continue PARENT
				}

				log.Printf("Produced Decomp w Parent as Root: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)

			}

			root := Node{Bag: Inter(parentλ.Vertices(), Vertices_H), Cover: parentλ, Children: subtrees}
			return Decomp{Graph: H, Root: root}
		}

		//Connectivity Check Parent
		vertCompLow := append(comp_low.Vertices(), VerticesSpecial(compSp_low)...)
		vertCompLowProper := Diff(vertCompLow, parentλ.Vertices())

		if !Subset(Inter(vertCompLow, Conn), parentλ.Vertices()) {
			// log.Println("Parent breaks connectivity wrt. comp_low")
			continue PARENT
		}

		log.Printf("Parent cover chosen: %v\n", Graph{Edges: parentλ})

		log.Printf("Comps of Parent: %v\n", comps_p)

		log.Println("Lower comp: \n", comp_low)

		Conn_child := Inter(vertCompLow, parentλ.Vertices())
		bound := FilterVertices(allowed, Conn_child)

		genChildCover := NewCover(l.K, Inter(vertCompLow, parentλ.Vertices()), bound, vertCompLow)

		allowedInsideComplow := FilterVertices(allowed, vertCompLowProper)

		//Child loops checks all possible child nodes for parent
	CHILD_OUTER:
		for genChildCover.HasNext {
			var tryPartAlone bool

			out := genChildCover.NextSubset()

			if out == -1 {
				if genChildCover.HasNext {
					log.Panicln(" -1 but hasNext not false!")
				}
				continue CHILD_OUTER
			}

			childPart := GetSubset(bound, genChildCover.Subset)
			if len(Inter(childPart.Vertices(), vertCompLowProper)) > 0 {
				tryPartAlone = true
			}

			remainingAllowed := allowedInsideComplow.Diff(childPart)

			genChild := GetCombin(remainingAllowed.Len(), l.K-childPart.Len())
		CHILD_INNER:
			for genChild.HasNext() || tryPartAlone {
				if !genChild.HasNext() { // run once on just childPart if it extends into comp
					tryPartAlone = false
				}

				var childλ Edges
				if !genChild.HasNext() {
					childλ = childPart
				} else {
					childAdded := GetSubset(remainingAllowed, genChild.Combination)
					childλ = NewEdges(append(childAdded.Slice(), childPart.Slice()...))
				}
				genChild.Confirm()

				childχ := Inter(childλ.Vertices(), vertCompLow)

				// Connectivity check
				if !Subset(Inter(vertCompLow, parentλ.Vertices()), childχ) {
					// log.Println("Child ", childλ, " breaks connectivity!")
					log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
					log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))

					log.Println("Child_part ", childPart)

					log.Println("Child Full ", childλ)

					log.Panicln("new loop not working, connectivity broken")
					continue CHILD_INNER
				}

				// // Check if progress made (TODO: implement this properly with minimal covers + extensions !)
				// if Subset(childχ, parentλ.Vertices()) {
				// 	// log.Println("No progress made!")

				// 	log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
				// 	log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))

				// 	log.Println("Child_part ", childPart)

				// 	log.Println("Child Full ", childλ)

				// 	log.Panicln("New loop not working, no progress made")
				// 	continue CHILD_INNER
				// }

				comps_c, compsSp_c, _, _ := comp_low.GetComponents(childλ, compSp_low)

				// log.Println("Size of H' : ", H.Edges.Len()+len(Sp))

				// Check childχ is balanced separator
				for i := range comps_c {
					if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
						log.Println("Child is not a bal. sep. !")
						continue CHILD_INNER
					}
				}

				log.Printf("Child chosen: %v\n", Graph{Edges: childλ})
				log.Printf("Comps of Child: %v\n", comps_c)

				//Computing subcomponents of Child

				var subtrees []Node
				for x := range comps_c {
					Conn_x := Inter(append(comps_c[x].Vertices(), VerticesSpecial(compsSp_c[x])...), childχ)

					decomp := l.findHD(comps_c[x], compsSp_c[x], Conn_x, allowed)
					if reflect.DeepEqual(decomp, Decomp{}) {
						log.Println("Rejecting parent as root")
						continue CHILD_INNER
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

				//TODO: if no comps_p, other than comp_low, just use parent as is

				if len(tempEdgeSlice) > 0 {

					comp_up.Edges = NewEdges(tempEdgeSlice)

					log.Println("Upper component:", comp_up)

					//Reducing the allowed edges
					allowedReduced := allowed.Diff(comp_low.Edges)

					// adding new Special Edge to connect Child to comp_up
					specialChild = Special{Vertices: childχ, Edges: childλ}

					decompUp = l.findHD(comp_up, append(compSp_up, specialChild), Conn, allowedReduced)

					if reflect.DeepEqual(decompUp, Decomp{}) {
						log.Println("Rejecting parent as root")
						continue CHILD_INNER
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

		}

	}

	// exhausted search space
	return Decomp{}
}
