package algorithms

import (
	"math"

	"github.com/cem-okulmus/BalancedGo/lib"
)

type SplitDecomp struct {
	k     int
	Graph lib.Graph
}

func (d *SplitDecomp) Name() string {
	return "SplitDecomp"
}

func (d *SplitDecomp) FindDecomp() lib.Decomp {
	return d.FindDecompGraph(d.Graph)
}

func (d *SplitDecomp) FindDecompGraph(G lib.Graph) lib.Decomp {
	d.SetWidth(0) // ignores 0 anyway

	childCover, childBag := split(G.Edges.Slice()[:d.k], len(G.Vertices()))
	child := lib.Node{Bag: childBag, Cover: childCover}

	rootCover, rootBag := split(G.Edges.Slice()[d.k:], len(G.Vertices()))
	root := lib.Node{Bag: rootBag, Cover: rootCover, Children: []lib.Node{child}}

	return lib.Decomp{Graph: G, Root: root}
}

func split(edges []lib.Edge, maxV int) (lib.Edges, []int) {
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

func (d *SplitDecomp) SetWidth(K int) {
	d.k = int(math.Ceil(float64(d.Graph.Edges.Len()) / 2))
}
