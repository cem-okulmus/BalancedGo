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
	if G.Edges.Len() < d.K {
		d.K = G.Edges.Len()
	}

	rootCover := lib.NewEdges(G.Edges.Slice()[:d.K])
	root := lib.Node{Bag: rootCover.Vertices(), Cover: rootCover}

	if G.Edges.Len() > d.K {
		childCover := lib.NewEdges(G.Edges.Slice()[d.K:])
		child := lib.Node{Bag: childCover.Vertices(), Cover: childCover}
		root.Children = []lib.Node{child}
	}

	return lib.Decomp{Graph: G, Root: root}
}
