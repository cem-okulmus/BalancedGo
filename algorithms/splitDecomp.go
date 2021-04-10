package algorithms

import (
	"github.com/cem-okulmus/BalancedGo/lib"
)

// SplitDecomp is a special algorihm that only tries to find a decomposition by splitting the
// hypergraph in two. This is only useful as a first step for the approximation method for finding
// decomposition.
type SplitDecomp struct {
	K     int
	Graph lib.Graph
}

// Name returns the name of the algorithm
func (d *SplitDecomp) Name() string {
	return "SplitDecomp"
}

// FindDecomp finds a decomp
func (d *SplitDecomp) FindDecomp() lib.Decomp {
	return d.FindDecompGraph(d.Graph)
}

// SetWidth sets the current width parameter of the algorithm
func (d *SplitDecomp) SetWidth(K int) {
	d.K = K
}

// FindDecompGraph finds a decomp, for an explicit lib.Graph
func (d *SplitDecomp) FindDecompGraph(G lib.Graph) lib.Decomp {
	childCover, childBag := split(G.Edges.Slice()[:d.K])
	child := lib.Node{Bag: childBag, Cover: childCover}

	rootCover, rootBag := split(G.Edges.Slice()[d.K:])
	root := lib.Node{Bag: rootBag, Cover: rootCover, Children: []lib.Node{child}}

	return lib.Decomp{Graph: G, Root: root}
}

func split(edges []lib.Edge) (lib.Edges, []int) {
	cover := lib.NewEdges(edges)
	bag := make([]int, 0)
	bagSet := make(map[int]bool)

	for _, e := range cover.Slice() {
		for _, v := range e.Vertices {
			if _, ok := bagSet[v]; !ok {
				bagSet[v] = true
				bag = append(bag, v)
			}
		}
	}
	return cover, bag
}
