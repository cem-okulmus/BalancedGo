package algorithms

import (
	"log"
	"reflect"
	"runtime"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// BalSepGlobal implements the global Balanced Separator algorithm.
// This requires all subedges to be added explicitly to the input lib.Graph.
type BalSepGlobal struct {
	K         int
	Graph     lib.Graph
	BalFactor int
}

// SetWidth sets the current width parameter of the algorithm
func (b *BalSepGlobal) SetWidth(K int) {
	b.K = K
}

func (b BalSepGlobal) findGHD() lib.Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecomp finds a decomp
func (b BalSepGlobal) FindDecomp() lib.Decomp {
	return b.findDecomp(b.Graph)
}

// FindDecompGraph finds a decomp, for an explicit lib.Graph
func (b BalSepGlobal) FindDecompGraph(G lib.Graph) lib.Decomp {
	return b.findDecomp(G)
}

// Name returns the name of the algorithm
func (b BalSepGlobal) Name() string {
	return "BalSep Global"
}

func baseCaseSmart(g lib.Graph, H lib.Graph) lib.Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", H.Special.Len() )
	var output lib.Decomp

	if H.Edges.Len() <= 2 && len(H.Special) == 0 {
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: H.Vertices(), Cover: H.Edges}}
	} else if H.Edges.Len() == 1 && len(H.Special) == 1 {
		sp1 := H.Special[0]
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: H.Edges.Vertices(), Cover: H.Edges,
				Children: []lib.Node{lib.Node{Bag: sp1.Vertices(), Cover: sp1}}}}
	} else {
		return baseCase(g, H)
	}
	return output
}

func baseCase(g lib.Graph, H lib.Graph) lib.Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", H.Special.Len())
	var output lib.Decomp
	switch len(H.Special) {
	case 0:
		output = lib.Decomp{Graph: g} // use g here to avoid reject
	case 1:
		sp1 := H.Special[0]
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: sp1.Vertices(), Cover: sp1}}
	case 2:
		sp1 := H.Special[0]
		sp2 := H.Special[1]
		output = lib.Decomp{Graph: H,
			Root: lib.Node{Bag: sp1.Vertices(), Cover: sp1,
				Children: []lib.Node{lib.Node{Bag: sp2.Vertices(), Cover: sp2}}}}
	}
	return output
}

func earlyTermination(H lib.Graph) lib.Decomp {
	//We assume that H as less than K edges, and only one special edge
	return lib.Decomp{Graph: H,
		Root: lib.Node{Bag: H.Edges.Vertices(), Cover: H.Edges,
			Children: []lib.Node{lib.Node{Bag: H.Special[0].Vertices(), Cover: H.Special[0]}}}}
}

func rerooting(H lib.Graph, balsep lib.Edges, subtrees []lib.Decomp) lib.Decomp {
	//Create a new GHD for H
	rerootNode := lib.Node{Bag: balsep.Vertices(), Cover: balsep}
	output := lib.Node{Bag: balsep.Vertices(), Cover: balsep}

	for _, s := range subtrees {
		// fmt.Println("H ", H, "balsep ", balsep, "comp ", s.Graph)
		s.Root = s.Root.Reroot(rerootNode)
		log.Printf("Rerooted Decomp: %v\n", s)
		output.Children = append(output.Children, s.Root.Children...)
	}
	// log.Println("H: ", H, "output: ", output)
	return lib.Decomp{Graph: H, Root: output}
}

func (b BalSepGlobal) findDecomp(H lib.Graph) lib.Decomp {
	// log.Printf("Current SubGraph: %+v\n", H)

	//stop if there are at most two special edges left
	if H.Len() <= 2 {
		return baseCaseSmart(b.Graph, H)
	}

	//Early termination
	if H.Edges.Len() <= b.K && len(H.Special) == 1 {
		return earlyTermination(H)
	}

	var balsep lib.Edges

	edges := lib.FilterVerticesStrict(b.Graph.Edges, append(H.Vertices()))
	generators := lib.SplitCombin(edges.Len(), b.K, runtime.GOMAXPROCS(-1), false)
	parallelSearch := lib.Search{H: &H, Edges: &edges, BalFactor: b.BalFactor, Generators: generators}
	pred := lib.BalancedCheck{}
	parallelSearch.FindNext(pred) // initial Search

OUTER:
	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {
		balsep = lib.GetSubset(edges, parallelSearch.Result)

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

		comps, _, _ := H.GetComponents(balsep)

		// log.Printf("Comps of Sep: %+v\n", comps)

		SepSpecial := lib.NewEdges(balsep.Slice())

		var subtrees []lib.Decomp
		ch := make(chan lib.Decomp)

		for i := range comps {
			go func(i int, comps []lib.Graph, SepSpecial lib.Edges) {
				comps[i].Special = append(comps[i].Special, SepSpecial)
				ch <- b.findDecomp(comps[i])
			}(i, comps, SepSpecial)
		}

		for i := 0; i < len(comps); i++ {
			decomp := <-ch
			if reflect.DeepEqual(decomp, lib.Decomp{}) {
				// log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: balsep}, comps[i],
				//  append(compsSp[i], SepSpecial))
				subtrees = []lib.Decomp{}
				//log.Printf("\n\nCurrent SubGraph: %v\n", H)
				//log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}
			// log.Printf("Produced Decomp: %+v\n", decomp)

			subtrees = append(subtrees, decomp)
		}

		return rerooting(H, balsep, subtrees)
	}

	// log.Printf("REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
	return lib.Decomp{} // empty Decomp signifiyng reject
}
