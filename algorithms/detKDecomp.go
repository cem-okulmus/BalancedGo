package algorithms

import (
	"log"
	"reflect"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type DetKDecomp struct {
	Graph     Graph
	BalFactor int
	SubEdge   bool
}

type CompCache struct {
	Succ []uint32
	Fail []uint32
}

var cache map[uint32]*CompCache

func connectingSep(sep []int, conn []int, comp []int) bool {
	if !Subset(conn, sep) {
		return false
	}

	if len(Inter(sep, comp)) == 0 {
		return false
	}

	return true
}

//Note: as implemented this breaks Special Condition (bag must be limited by oldSep)
func baseCaseDetK(g Graph, H Graph, Sp []Special) Decomp {
	log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children Node

	switch len(Sp) {
	case 0:
		return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges}}
	case 1:
		children = Node{Bag: Sp[0].Vertices, Cover: Sp[0].Edges}
	case 2:
		children = Node{Bag: Sp[0].Vertices, Cover: Sp[0].Edges,
			Children: []Node{Node{Bag: Sp[1].Vertices, Cover: Sp[1].Edges}}}

	}

	if H.Edges.Len() == 0 {
		return Decomp{Graph: H, Root: children}
	}
	return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges, Children: []Node{children}}}
}

// TODO add caching to this
func (d DetKDecomp) findDecomp(K int, H Graph, oldSep []int, Sp []Special) Decomp {
	verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
	conn := Inter(oldSep, verticesCurrent)
	compVertices := Diff(verticesCurrent, oldSep)
	bound := FilterVertices(d.Graph.Edges, conn)

	log.Printf("\n\nCurrent oldSep: %v, Conn: %v\n", PrintVertices(oldSep), PrintVertices(conn))
	log.Printf("Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len())
	log.Printf("Current Special Edges: %v\n\n", Sp)

	// Base case if H <= K
	if H.Edges.Len() <= K && len(Sp) <= 1 {
		return baseCaseDetK(d.Graph, H, Sp)
	}

	gen := NewCover(K, conn, bound, H.Edges)

OUTER:
	for gen.HasNext {
		out := gen.NextSubset()

		if out == -1 {
			if gen.HasNext {
				log.Panicln(" -1 but hasNext not false!")
			}
			continue
		}

		var sep Edges
		sep = GetSubset(bound, gen.Subset)
		log.Println("Next Cover ", sep)

		addEdges := false

		//check if sep "makes some progress" into separating H

		if len(Inter(sep.Vertices(), compVertices)) == 0 {
			addEdges = true
		}

		if !addEdges || K-sep.Len() > 0 {
			i_add := 0

		addingEdges:
			for !addEdges || i_add < H.Edges.Len() {
				var sepActual Edges

				if addEdges {
					sepActual = NewEdges(append(sep.Slice(), H.Edges.Slice()[i_add]))
				} else {
					sepActual = sep
				}

				sepActualOrigin := sepActual
				var sepSub *SepSub
			subEdges:
				for true {

					log.Println("Sep chosen ", sepActual, " out ", out)
					comps, compsSp, _ := H.GetComponents(sepActual, Sp)

					log.Printf("Comps of Sep: %v\n", comps)

					var subtrees []Node
					bag := Inter(sepActual.Vertices(), append(oldSep, verticesCurrent...))

					for i := range comps {
						decomp := d.findDecomp(K, comps[i], bag, compsSp[i])
						if reflect.DeepEqual(decomp, Decomp{}) {
							log.Printf("REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sepActual}, comps[i], compsSp[i])
							log.Printf("\n\nCurrent oldSep: %v\n", oldSep)
							log.Printf("Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len())
							log.Printf("Current Special Edges: %v\n\n", Sp)

							if d.SubEdge {
								if sepSub == nil {
									sepSub = GetSepSub(d.Graph.Edges, sepActual, K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
										// log.Println("SubSep: ")
										// for _, s := range sepSub.Edges {
										// 	log.Println(s.Combination)
										// }
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: sepActualOrigin}, Sp)
										continue OUTER
									}
								}
								log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
								continue subEdges
							}

							//	cache[sepActual.Hash()].Fail = append(cache[sepActual.Hash()].Fail, H.Edges.Hash())
							if addEdges {
								i_add++
								continue addingEdges
							} else {
								continue OUTER
							}
						}

						log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}
					//cache[sepActual.Hash()].Succ = append(cache[sepActual.Hash()].Succ, H.Edges.Hash())

					return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}

func (d DetKDecomp) FindHD(K int, Sp []Special) Decomp {
	cache = make(map[uint32]*CompCache)
	return d.findDecomp(K, d.Graph, []int{}, Sp)
}
