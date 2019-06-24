// Decomposition method based on "Balanced Decompositions" Akatov et al.
package main

import (
	"log"
	"reflect"
)

type balKDecomp struct {
	graph Graph
}

func (b balKDecomp) findDecomp(K int, H Graph) Decomp {

	log.Printf("\n\nCurrent Subgraph: %v\n", H)
	gen := getCombin(len(H.edges), K)

OUTER:
	for gen.hasNext() {
		balsep := getSubset(H.edges, gen.combination)

		log.Printf("Testing: %v\n", Graph{edges: balsep})
		gen.confirm()
		if !H.checkBalancedSep(balsep, []Special{}) {
			continue
		}

		log.Printf("Balanced Sep chosen: %v\n", Graph{edges: balsep})

		comps, _, _ := H.getComponents(balsep, []Special{})

		sepVertices := Vertices(balsep)

		var subtrees []Node
		for _, c := range comps {

			var newEdges []Edge

			//culling edges to be vertex-distinct from sep
			for _, e := range c.edges {
				tempVertices := diff(e.vertices, sepVertices)

				if len(tempVertices) > 0 {
					m[encode] = m[e.name] + "'"
					nuName := encode
					encode++

					newEdges = append(newEdges, Edge{name: nuName, vertices: tempVertices})
				}
			}

			decomp := b.findDecomp(K, Graph{edges: newEdges})
			if reflect.DeepEqual(decomp, Decomp{}) {
				log.Printf("REJECTING %v: couldn't decompose %v \n", Graph{edges: balsep}, c)
				log.Printf("\n\nCurrent Subgraph: %v\n", H)
				continue OUTER
			}

			log.Printf("Produced Decomp: %v\n", decomp)
			subtrees = append(subtrees, decomp.root)
		}

		node := Node{bag: sepVertices, cover: balsep, children: subtrees}
		return Decomp{graph: H, root: node}

	}

	log.Printf("REJECT: Couldn't find balsep for H %v\n", H)
	return Decomp{} // using empty decomp as failure
}

func (b balKDecomp) findBD(K int) Decomp {
	return b.findDecomp(K, b.graph)
}
