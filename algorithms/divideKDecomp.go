// Follows the algorithm k divide decomp from Samer "Exploiting Parallelism in Decomposition Methods for Constraint Satisfaction"
package algorithms

import (
	"bytes"
	"fmt"
	"log"
	"reflect"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type divideComp struct {
	up           []int
	length       int
	edges        Edges
	low          []int
	upConnecting bool
}

func (c divideComp) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Up: ")
	buffer.WriteString(Edge{Vertices: c.up}.String())

	buffer.WriteString(" Edges: ")
	buffer.WriteString(Graph{Edges: c.edges}.String())

	buffer.WriteString(" Low: ")
	buffer.WriteString(Edge{Vertices: c.low}.String())

	buffer.WriteString(" upConnecting: ")
	buffer.WriteString(fmt.Sprintln(c.upConnecting))

	return buffer.String()
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

func (comp divideComp) getComponents(sep Edges) ([]divideComp, bool) {
	var output []divideComp

	// log.Println("Edges coming in, ", comp.edges)

	//special case
	if Subset(comp.edges.Vertices(), sep.Vertices()) {
		return []divideComp{}, true
	}

	var mustCover = Inter(comp.up, comp.low)
	if !Subset(mustCover, sep.Vertices()) {
		return output, false
	}

	// log.Println("Testing ", sep)

	var up []int
	var low []int
	if !Subset(comp.up, sep.Vertices()) {
		up = comp.up // ignore up connection if sep already fully covers it
	}
	if !Subset(comp.low, sep.Vertices()) {
		low = comp.low // ignore low connection if sep already fully covers it
	}

	sepBoth := NewEdges(sep.Both(comp.edges))

	comps, upIndex, lowIndex, isolatedComp := Graph{Edges: comp.edges}.GetComponentsIsolated(sep, sepBoth, up, low)

	// log.Println("Edges coming out, ", len(comps), " many comps:")

	// for i := range comps {
	// 	log.Println("comp ", comps[i].Edges)
	// }

	// log.Println("UpIndex ", upIndex, " LowIndex", lowIndex)

	if upIndex != -1 && upIndex == lowIndex { // reject case, up and low not seperated
		return output, false
	}

	unionOfUps := []int{}

	// take care of component, calculating connection sets
	for i := range comps {
		c := divideComp{}
		c.length = comps[i].Edges.Len() - len(sep.Both(comps[i].Edges))

		if i == upIndex { // Upper component
			c.upConnecting = true
			compEdges := comps[i].Edges.Slice()
			c.low = Inter(sep.Vertices(), comps[i].Edges.Vertices())

			c.edges = NewEdges(append(compEdges, sepBoth.Slice()...))
			c.edges.RemoveDuplicates()

			c.up = comp.up
		} else if i == lowIndex { // lower component
			c.edges = comps[i].Edges
			c.low = Inter(comp.low, c.edges.Vertices())
			c.up = Inter(sep.Vertices(), c.edges.Vertices())
			unionOfUps = append(unionOfUps, c.up...)
		} else {
			c.edges = comps[i].Edges
			c.up = Inter(sep.Vertices(), c.edges.Vertices())
			unionOfUps = append(unionOfUps, c.up...)

		}

		output = append(output, c)
	}
	if upIndex != -1 && len(unionOfUps) > 0 {
		output[upIndex].low = append(output[upIndex].low, unionOfUps...)
		output[upIndex].low = Inter(output[upIndex].edges.Vertices(), output[upIndex].low)
		output[upIndex].low = RemoveDuplicates(output[upIndex].low)

	}
	// //Add isolated comp to low set
	if upIndex != -1 && len(isolatedComp) > 0 {
		coveredEdges := NewEdges(isolatedComp)
		output[upIndex].low = append(output[upIndex].low, coveredEdges.Vertices()...)
		output[upIndex].low = Inter(output[upIndex].edges.Vertices(), output[upIndex].low)
		output[upIndex].low = RemoveDuplicates(output[upIndex].low)

	}

	return output, true
}

type DivideKDecomp struct {
	Graph     Graph
	K         int
	BalFactor int
}

func (d DivideKDecomp) CheckBalancedSep(comp divideComp, comps []divideComp, valid bool) bool {
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
	if len(comp.low) == 0 {
		// log.Printf("Components of sep %+v\n", comps)
		for i := range comps {

			if comps[i].edges.Len() == comp.edges.Len() { // not made any progres
				log.Println("No progress made")
				return false
			}

			if comps[i].length > (((comp.edges.Len()) * (d.BalFactor - 1)) / d.BalFactor) {
				// log.Printf("Using component %+v has weight %d instead of %d\n", comps[i], comps[i].edges.Len(), (((comp.edges.Len()) * (d.BalFactor - 1)) / d.BalFactor))
				return false
			}
		}
	} else {
		for i := range comps {
			if comps[i].edges.Len() == comp.edges.Len() { // not made any progres
				// log.Println("No progress made")
				return false
			}
			if len(comps[i].low) == 0 {
				continue
			}
			if comps[i].length > (((comp.edges.Len()) * (d.BalFactor - 1)) / d.BalFactor) {
				// log.Printf("Using component %+v has weight %d instead of %d\n", comps[i], comps[i].edges.Len(), (((comp.edges.Len()) * (d.BalFactor - 1)) / d.BalFactor))
				return false
			}
		}
		// if len(Inter(sep.Vertices(), comp.edges.Vertices())) == 0 { //make some progress
		// 	return false
		// }
	}

	return true
}

//TODO check if this kind of manipulation actually works outside of current scope
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

// TODO: Remove the blind special edge here in the post process step
func (d DivideKDecomp) baseCase(comp divideComp) Decomp {
	log.Println("Using base case")
	det := DetKDecomp{Graph: d.Graph, BalFactor: d.BalFactor, SubEdge: false, Divide: true}
	det.cache = make(map[uint32]*CompCache)
	var H Graph

	H.Edges = NewEdges(comp.edges.Slice())

	log.Println("up", PrintVertices(comp.up))
	log.Println("H", H.Edges)
	if len(comp.low) == 0 {
		return det.findDecomp(d.K, H, comp.up, []Special{})
	} else {
		return det.findDecomp(d.K, H, comp.up, []Special{Special{Vertices: comp.low}})

	}
}

func (d DivideKDecomp) decomposable(comp divideComp) Decomp {

	// if !Subset(comp.low, comp.edges.Vertices()) {
	// 	log.Println("comp ", comp)
	// 	log.Panicln("low set not inside edges")
	// }
	if !Subset(comp.up, comp.edges.Vertices()) {
		log.Println("comp ", comp)
		log.Panicln("up set not inside edges")
	}

	log.Printf("\n\nCurrent SubGraph: %v\n", comp)

	//base case: size of comp <= K

	if comp.edges.Len() <= 2*d.K {
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

	conn = Inter(comp.low, comp.up) // What happens if this is empty?
	edges := FilterVertices(d.Graph.Edges, comp.edges.Vertices())
	genCov = NewCover(d.K, conn, edges, comp.edges)

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
			gen.Confirm()
			balsep := GetSubset(edges, gen.Combination)
			//extend balsep with sep above
			if !firstPass {
				balsep = NewEdges(append(balsep.Slice(), sep.Slice()...))

			}
			firstPass = false

			comps, valid := comp.getComponents(balsep)

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
				child := d.decomposable(comps[i])
				if reflect.DeepEqual(child, Decomp{}) {
					log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{Edges: balsep}, comps[i])
					log.Printf("\n\nCurrent SubGraph: %v\n", comp)
					continue OUTER
				}

				log.Printf("Produced Decomp: %v\n", child)

				if comps[i].upConnecting {
					upconnecting = true
					parent = child.Root
					parent.Bag = comps[i].edges.Vertices()
					parent.Up = comps[i].up
					parent.Low = comps[i].low

				} else {
					subtrees = append(subtrees, child.Root)
				}
			}

			var output Node
			upFlag := false
			if len(comp.low) > 0 && Subset(comp.low, balsep.Vertices()) {
				upFlag = true
			}

			var bag []int

			if upconnecting {

				for i := range subtrees {
					childComp := append(subtrees[i].Bag, subtrees[i].Up...)
					bag = append(bag, childComp...)
				}

			} else {
				bag = Diff(comp.edges.Vertices(), comp.up)
			}

			SubtreeRootedAtS := Node{LowConnecting: upFlag, Up: comp.up, Low: comp.low, Cover: balsep, Bag: bag, Children: subtrees}

			if reflect.DeepEqual(parent, Node{}) && (!Subset(comp.up, balsep.Vertices())) {

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

			output.Up = make([]int, len(comp.up))
			copy(output.Up, output.Up)
			output.Low = make([]int, len(comp.low))
			copy(output.Low, comp.low)

			return Decomp{Graph: d.Graph, Root: output}
		}
	}

	log.Println("REJECT: Couldn't find a sep for ", comp)

	return Decomp{} // using empty decomp as reject case

}

func (d DivideKDecomp) FindDecomp(K int) Decomp {
	output := d.decomposable(divideComp{edges: d.Graph.Edges})
	output.Root = postProcess(output.Root, []int{})

	return output
}

func (d DivideKDecomp) Name() string {
	return "DivideK"
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
