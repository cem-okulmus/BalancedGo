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
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
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
	// log.Println("Two Nodes enter: ", subtreeAbove, subtreeBelow)
	// up = Inter(up, parent.Vertices())
	//finding connecting leaf in parent
	leaf := subtreeAbove.CombineNodes(subtreeBelow, connecting)
	if reflect.DeepEqual(*leaf, Node{}) {
		fmt.Println("\n \n Connection ", PrintVertices(connecting.Vertices))
		fmt.Println("subtreeAbove ", subtreeAbove)

		log.Panicln("subtreeAbove doesn't contain connecting node!")
	}

	// //attaching subtree to parent
	// leaf.Children = []Node{subtree}
	// log.Println("Leaf ", leaf)
	// log.Println("One Node leaves: ", parent)
	return *leaf
}

func (l LogKDecomp) findHD(H Graph, Sp []Special, Conn []int, allowed Edges) Decomp {

	log.Printf("\n\nCurrent SubGraph: %v\n", H)
	log.Printf("Current Special Edges: %v\n\n", Sp)
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

		lowFound := false

		log.Printf("Parent cover chosen: %v\n", Graph{Edges: parentλ})

		log.Printf("Comps of Parent: %v\n", comps_p)

		// Check if lower component present
		for i := range comps_p {
			if comps_p[i].Edges.Len()+len(compsSp_p[i]) >= (H.Edges.Len() + len(Sp)/2) {
				lowFound = true
				comp_low_index = i //keep track of the index for composing comp_up later
				comp_low = comps_p[i]
				compSp_low = compsSp_p[i]
			}
		}

		// Check if Parent is possible root
		if !lowFound {
			log.Println("parentλ is bal. sep. !")
			// Connectivity Check
			if !Subset(Conn, parentλ.Vertices()) {
				log.Println("parent breaks connectivity: ")
				log.Println("Conn: ", PrintVertices(Conn))
				log.Println("V(parentλ) = ", PrintVertices(parentλ.Vertices()))
				continue PARENT
			}

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

			root := Node{Bag: Inter(parentλ.Vertices(), H.Vertices()), Cover: parentλ, Children: subtrees}
			return Decomp{Graph: H, Root: root}
		}

		log.Println("Lower comp: \n", comp_low)

		genChild := GetCombin(allowed.Len(), l.K)

		//Child loop checks all possible child nodes for parent
	CHILD:
		for genChild.HasNext() {

			childλ := GetSubset(allowed, genChild.Combination)
			genChild.Confirm()

			log.Printf("Child chosen: %v\n", Graph{Edges: childλ})

			vertCompLow := append(comp_low.Vertices(), VerticesSpecial(compSp_low)...)

			childχ := Inter(childλ.Vertices(), vertCompLow)

			// Connectivity check
			if !Subset(Inter(vertCompLow, parentλ.Vertices()), childχ) {
				log.Println("Child ", childλ, " breaks connectivity!")
				continue CHILD
			}

			// Check if progress made (TODO: implement this properly with minimal covers + extensions !)
			if Subset(childχ, parentλ.Vertices()) {
				log.Println("No progress made!")
				continue CHILD
			}

			comps_c, compsSp_c, _, _ := comp_low.GetComponents(childλ, Sp)

			log.Printf("Comps of Child: %v\n", comps_c)
			log.Println("Size of H' : ", H.Edges.Len()+len(Sp))

			// Check childχ is balanced separator
			for i := range comps_c {
				if comps_c[i].Edges.Len()+len(compsSp_c[i]) > (H.Edges.Len()+len(Sp))/2 {
					log.Println("Child is not a bal. sep. !")
					continue CHILD
				}
			}

			//Computing subcomponents of Child

			var subtrees []Node
			for x := range comps_c {
				Conn_x := Inter(append(comps_c[x].Vertices(), VerticesSpecial(compsSp_c[x])...), parentλ.Vertices())

				decomp := l.findHD(comps_c[x], compsSp_c[x], Conn_x, allowed)
				if reflect.DeepEqual(decomp, Decomp{}) {
					log.Println("Rejecting parent as root")
					continue PARENT
				}

				log.Printf("Produced Decomp: %+v\n", decomp)
				subtrees = append(subtrees, decomp.Root)

			}

			//Computing upper component

			var comp_up Graph
			var compSp_up []Special
			tempEdgeSlice := []Edge{}

			tempEdgeSlice = append(tempEdgeSlice, isolatedEdges...)
			for i := range comps_p {
				if i != comp_low_index {
					tempEdgeSlice = append(tempEdgeSlice, comps_p[i].Edges.Slice()...)
					compSp_up = append(compSp_up, compsSp_p[i]...)
				}
			}

			comp_up.Edges = NewEdges(tempEdgeSlice)

			//Reducing the allowed edges
			allowedReduced := allowed.Diff(comp_low.Edges)

			// adding new Special Edge to connect Child to comp_up
			specialChild := Special{Vertices: childχ, Edges: childλ}

			decompUp := l.findHD(comp_up, append(compSp_up, specialChild), Conn, allowedReduced)

			if reflect.DeepEqual(decompUp, Decomp{}) {
				log.Println("Rejecting parent as root")
				continue PARENT
			}
			subtrees = append(subtrees, decompUp.Root)

			// rearrange subtrees to form one that covers total of H
			rootChild := Node{Bag: childχ, Cover: childλ, Children: subtrees}
			finalRoot := l.attachingSubtrees(decompUp.Root, rootChild, specialChild)

			log.Printf("Produced Decomp: %v\n", finalRoot)
			return Decomp{Graph: H, Root: finalRoot}

		}

	}

	// exhausted search space
	return Decomp{}
}
