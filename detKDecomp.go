package main

import (
	"log"
	"reflect"
)

type detKDecomp struct {
	graph Graph
}

// TODO add caching to this
func (d detKDecomp) findDecomp(K int, H Graph, oldSep []Edge, Sp []Special) Decomp {

	// Base case if H <= K
	if len(H.edges) <= K && len(Sp) <= 2 {
		var children []Node

		if len(Sp) == 1 {
			children = []Node{Node{bag: Sp[0].vertices, cover: Sp[0].edges}}
		} else if len(Sp) == 2 {
			children = []Node{Node{bag: Sp[0].vertices, cover: Sp[0].edges,
				children: []Node{Node{bag: Sp[1].vertices, cover: Sp[1].edges}}}}
		}

		return Decomp{graph: H, root: Node{bag: H.Vertices(), cover: H.edges, children: children}}
	}

	//TODO: think about whether filtering here is allowed, and if it should be strict or not
	edges := filterVerticesStrict(d.graph.edges, append(H.Vertices(), VerticesSpecial(Sp)...))
	gen := getCombin(len(edges), K)

OUTER:
	for gen.hasNext() {
		gen.confirm()
		balsep := getSubset(edges, gen.combination)

		verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
		// check if balsep covers the intersection of oldsep and H
		if !subset(inter(Vertices(oldSep), verticesCurrent), Vertices(balsep)) {
			continue
		}
		//check if balsep "makes some progress" into separating H
		if len(inter(Vertices(balsep), diff(verticesCurrent, Vertices(oldSep)))) == 0 {
			continue
		}

		comps, compsSp, _ := H.getComponents(balsep, Sp)

		var subtrees []Node
		for i := range comps {
			decomp := d.findDecomp(K, comps[i], balsep, compsSp[i])
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{edges: balsep}, comps[i], compsSp[i])
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				log.Printf("Current Special Edges: %v\n\n", Sp)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		bag := inter(Vertices(balsep), append(Vertices(oldSep), H.Vertices()...))
		return Decomp{graph: H, root: Node{bag: bag, cover: balsep, children: subtrees}}
	}

	return Decomp{} // Reject if no separator could be found
}
