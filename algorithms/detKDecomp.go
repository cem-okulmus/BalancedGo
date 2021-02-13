package algorithms

import (
	"log"
	"reflect"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// DetKDecomp computes for a graph and some width K a HD of width K if it exists
type DetKDecomp struct {
	K         int
	Graph     lib.Graph
	BalFactor int
	SubEdge   bool
	cache     lib.Cache
}

// SetWidth sets the current width parameter of the algorithm
func (d *DetKDecomp) SetWidth(K int) {
	d.K = K
}

func (d *DetKDecomp) findHD(currentGraph lib.Graph) lib.Decomp {
	d.cache.Init()
	return d.findDecomp(currentGraph, []int{})
}

// FindDecomp finds a decomp
func (d *DetKDecomp) FindDecomp() lib.Decomp {
	return d.findHD(d.Graph)
}

// Name returns the name of the algorithm
func (d *DetKDecomp) Name() string {
	if d.SubEdge {
		return "DetK with local BIP"
	}
	return "DetK"
}

// FindDecompGraph finds a decomp, for an explicit graph
func (d *DetKDecomp) FindDecompGraph(G lib.Graph) lib.Decomp {
	return d.findHD(G)
}

func connectingSep(sep []int, conn []int, comp []int) bool {
	if !lib.Subset(conn, sep) {
		return false
	}

	if len(lib.Inter(sep, comp)) == 0 {
		return false
	}

	return true
}

//Note: as implemented this breaks Special Condition (bag must be limited by oldSep)
func baseCaseDetK(H lib.Graph) lib.Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children lib.Node

	switch len(H.Special) {
	case 0:
		return lib.Decomp{Graph: H, Root: lib.Node{Bag: H.Vertices(), Cover: H.Edges}}

	case 1:
		sp1 := H.Special[0]
		children = lib.Node{Bag: sp1.Vertices(), Cover: sp1}
	}

	if H.Edges.Len() == 0 {
		return lib.Decomp{Graph: H, Root: children}
	}
	return lib.Decomp{Graph: H, Root: lib.Node{Bag: H.Vertices(), Cover: H.Edges, Children: []lib.Node{children}}}
}

func (d *DetKDecomp) findDecomp(H lib.Graph, oldSep []int) lib.Decomp {
	verticesCurrent := append(H.Vertices())
	verticesExtended := append(verticesCurrent, oldSep...)
	conn := lib.Inter(oldSep, verticesCurrent)
	compVertices := lib.Diff(verticesCurrent, oldSep)
	bound := lib.FilterVertices(d.Graph.Edges, conn)

	// log.Printf("\n\nD Current oldSep: %v, Conn: %v\n", lib.PrintVertices(oldSep), lib.PrintVertices(conn))
	// log.Printf("D Current SubGraph: %v ( %v hash) \n", H, H.Edges.Hash())
	// log.Printf("D Current SubGraph: %v ( %v edges) (hash: %v )\n", H, H.Edges.Len(), H.Edges.Hash())
	// log.Println("D Hedges ", H)
	// log.Println("D Comp Vertices: ", lib.PrintVertices(compVertices))

	// Base case if H <= K
	if H.Edges.Len() == 0 && len(H.Special) <= 1 {
		return baseCaseDetK(H)
	}

	gen := lib.NewCover(d.K, conn, bound, H.Edges.Vertices())

OUTER:
	for gen.HasNext {
		out := gen.NextSubset()

		if out == -1 {
			if gen.HasNext {
				log.Panicln(" -1 but hasNext not false!")
			}
			continue
		}

		var sep lib.Edges
		sep = lib.GetSubset(bound, gen.Subset)

		// if !Subset(conn, sep.Vertices()) {
		//  log.Panicln("Cover messed up! 137")
		// }
		// log.Println("Next Cover ", sep)

		addEdges := false

		//check if sep "makes some progress" into separating H

		if len(lib.Inter(sep.Vertices(), compVertices)) == 0 {
			addEdges = true
		}

		if !addEdges || d.K-sep.Len() > 0 {
			iAdd := 0

		addingEdges:
			for !addEdges || iAdd < H.Edges.Len() {
				var sepActual lib.Edges

				if addEdges {
					sepActual = lib.NewEdges(append(sep.Slice(), H.Edges.Slice()[iAdd]))
				} else {
					sepActual = sep
				}

				// sepActualOrigin := sepActual
				var sepSub *lib.SepSub
				var sepConst []lib.Edge
				var sepChanging []lib.Edge
				if d.SubEdge {
					for i, v := range gen.Subset {
						if gen.InComp[v] {
							sepChanging = append(sepChanging, sep.Slice()[i])
						} else {
							sepConst = append(sepConst, sep.Slice()[i])
						}
					}
					if addEdges {
						sepChanging = append(sepChanging, H.Edges.Slice()[iAdd])
					}
				}

			subEdges:
				for true {

					// log.Println("Sep chosen ", sepActual, " out ", out)
					comps, _, _ := H.GetComponents(sepActual)

					//check cache for previous encounters
					if d.cache.CheckNegative(sepActual, comps) {
						// log.Println("Skipping sep", sepActual, "due to cache.")
						if addEdges {
							iAdd++
							continue addingEdges
						} else {
							continue OUTER
						}
					}

					// log.Printf("Comps of Sep: %v, len: %v\n", comps, len(comps))

					var subtrees []lib.Node
					bag := lib.Inter(sepActual.Vertices(), verticesExtended)

					for i := range comps {
						decomp := d.findDecomp(comps[i], bag)
						if reflect.DeepEqual(decomp, lib.Decomp{}) {

							d.cache.AddNegative(sepActual, comps[i])
							// log.Printf("detK REJECTING %v: couldn't decompose %v  \n",
							// 	Graph{Edges: sepActual}, comps[i])
							// log.Printf("\n\nCurrent oldSep: %v\n", PrintVertices(oldSep))
							// log.Printf("Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len(), H.Edges.Hash())

							if d.SubEdge {
								if sepSub == nil {
									sepSub = lib.GetSepSub(d.Graph.Edges, lib.NewEdges(sepChanging), d.K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										sepActual = lib.NewEdges(append(sepActual.Slice(), sepConst...))
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										// log.Printf("No SubSep found for %v  \n", Graph{Edges: sepActualOrigin})
										if addEdges {
											iAdd++
											continue addingEdges
										} else {
											continue OUTER
										}
									}
								}
								// log.Printf("Sub Sep chosen: %vof %v \n", Graph{Edges: sepActual},
								// 	Graph{Edges: sepActualOrigin})
								continue subEdges
							}

							if addEdges {
								iAdd++
								continue addingEdges
							} else {
								continue OUTER
							}
						}
						//d.Cache.AddPositive(sepActual, comps[i])
						// log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}

					return lib.Decomp{Graph: H, Root: lib.Node{Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}
		}
	}

	return lib.Decomp{} // Reject if no separator could be found
}
