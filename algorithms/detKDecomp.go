package algorithms

import (
	"fmt"
	"log"
	"reflect"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type DetKDecomp struct {
	K         int
	Graph     Graph
	BalFactor int
	SubEdge   bool
	Cache     Cache
}

func (d *DetKDecomp) SetWidth(K int) {
	d.K = K
}

func (d *DetKDecomp) FindHD(currentGraph Graph) Decomp {

	// d.Cache = make(map[uint64]*CompCache)
	d.Cache.Init()
	return d.findDecomp(currentGraph, []int{})
}

func (d *DetKDecomp) FindDecomp() Decomp {
	return d.FindHD(d.Graph)
}

func (d *DetKDecomp) Name() string {
	if d.SubEdge {
		return "DetK with local BIP"
	} else {
		return "DetK"
	}
}

func (d *DetKDecomp) FindDecompGraph(G Graph) Decomp {
	return d.FindHD(G)
}

var counterMap map[string]int

func (d *DetKDecomp) FindDecompUpdate(graph Graph, savedScenes map[uint32]SceneValue, savedCache Cache) Decomp {
	// d.Cache = make(map[uint64]*CompCache)
	// d.Cache.Init()
	d.Cache = savedCache // use provided cache

	fmt.Println("Cache Size at start:", d.Cache.Len())

	if log.Flags() == 0 {
		counterMap = make(map[string]int)
		defer func(map[string]int) {

			fmt.Println("Counter Map:")

			for k, v := range counterMap {
				fmt.Println("Scene: ", k, "\nTimes Used: ", v)
			}

		}(counterMap)
	}

	return d.findDecompUpdate(graph, []int{}, savedScenes)
}

func (d *DetKDecomp) GetCache() Cache {
	return d.Cache
}

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
func baseCaseDetK(H Graph) Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children Node

	switch len(H.Special) {
	case 0:
		return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges}}

	case 1:
		sp1 := H.Special[0]
		children = Node{Bag: sp1.Vertices(), Cover: sp1}
	}

	if H.Edges.Len() == 0 {
		return Decomp{Graph: H, Root: children}
	}
	return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges, Children: []Node{children}}}
}

func (d *DetKDecomp) findDecomp(H Graph, oldSep []int) Decomp {
	verticesCurrent := append(H.Vertices())
	verticesExtended := append(verticesCurrent, oldSep...)
	conn := Inter(oldSep, verticesCurrent)
	compVertices := Diff(verticesCurrent, oldSep)
	bound := FilterVertices(d.Graph.Edges, conn)

	// log.Printf("\n\nD Current oldSep: %v, Conn: %v\n", PrintVertices(oldSep), PrintVertices(conn))
	// log.Printf("D Current SubGraph: %v ( %v hash) \n", H, H.Edges.Hash())
	// log.Printf("D Current SubGraph: %v ( %v edges) (hash: %v )\n", H, H.Edges.Len(), H.Edges.Hash())
	// log.Printf("D Current Special Edges: %v\n\n", Sp)
	// log.Println("D Hedges ", H)
	// log.Println("D Comp Vertices: ", PrintVertices(compVertices))

	// Base case if H <= K
	if H.Edges.Len() == 0 && len(H.Special) <= 1 {
		return baseCaseDetK(H)
	}

	gen := NewCover(d.K, conn, bound, H.Edges.Vertices())

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

		// if !Subset(conn, sep.Vertices()) {
		//  log.Panicln("Cover messed up! 137")
		// }

		// log.Println("Next Cover ", sep)

		addEdges := false

		//check if sep "makes some progress" into separating H

		if len(Inter(sep.Vertices(), compVertices)) == 0 {
			addEdges = true
		}

		if !addEdges || d.K-sep.Len() > 0 {
			i_add := 0

		addingEdges:
			for !addEdges || i_add < H.Edges.Len() {
				var sepActual Edges

				if addEdges {
					sepActual = NewEdges(append(sep.Slice(), H.Edges.Slice()[i_add]))
				} else {
					sepActual = sep
				}

				// sepActualOrigin := sepActual
				var sepSub *SepSub
				var sepConst []Edge
				var sepChanging []Edge
				if d.SubEdge {
					for i, v := range gen.Subset {
						if gen.InComp[v] {
							sepChanging = append(sepChanging, sep.Slice()[i])
						} else {
							sepConst = append(sepConst, sep.Slice()[i])
						}
					}
					if addEdges {
						sepChanging = append(sepChanging, H.Edges.Slice()[i_add])
					}
				}

			subEdges:
				for true {

					// log.Println("Sep chosen ", sepActual, " out ", out)
					comps, _, _ := H.GetComponents(sepActual)

					//check chache for previous encounters
					if d.Cache.CheckNegative(sepActual, comps) {
						if addEdges {
							i_add++
							continue addingEdges
						} else {
							continue OUTER
						}
					}

					// log.Printf("Comps of Sep: %v, len: %v\n", comps, len(comps))

					var subtrees []Node
					bag := Inter(sepActual.Vertices(), verticesExtended)

					// log.Println("sep", sep, "\nsepActual", sepActual, "\n B of SepActual",
					//      PrintVertices(sepActual.Vertices()), "\noldSep", PrintVertices(oldSep), "\nvertices of C",
					//      PrintVertices(verticesCurrent), "\n\nunion o both", PrintVertices(verticesExtended),
					//      "\n bag: ", PrintVertices(bag))

					// for i := range sepActual.Vertices() {
					//  if Mem(verticesCurrent, sepActual.Vertices()[i]) && !Mem(bag, sepActual.Vertices()[i]) {

					//      fmt.Println("Another union: ", PrintVertices(append(oldSep, verticesCurrent...)))

					//      fmt.Println("Another intersect: ", PrintVertices(Inter(sepActual.Vertices(),
					//                 verticesExtended)))

					//      fmt.Println("sep", sep, "\nsepActual", sepActual, "\n B of SepActual",
					//            PrintVertices(sepActual.Vertices()), "\noldSep ", PrintVertices(oldSep),
					//          "\nvertices of C", PrintVertices(verticesCurrent), "\n\nunion o both",
					//                PrintVertices(verticesExtended), "\n bag: ", PrintVertices(bag))

					//      log.Panicln("something is not right in the state of this program!")
					//  }
					// }

					for i := range comps {
						decomp := d.findDecomp(comps[i], bag)
						if reflect.DeepEqual(decomp, Decomp{}) {

							d.Cache.AddNegative(sepActual, comps[i])
							// log.Printf("detK REJECTING %v: couldn't decompose %v with SP %v \n",
							// Graph{Edges: sepActual}, comps[i], compsSp[i])
							// log.Printf("\n\nCurrent oldSep: %v\n", PrintVertices(oldSep))
							// log.Printf("Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len(), H.Edges.Hash())
							// log.Printf("Current Special Edges: %v\n\n", Sp)

							if d.SubEdge {
								if sepSub == nil {
									sepSub = GetSepSub(d.Graph.Edges, NewEdges(sepChanging), d.K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										sepActual = NewEdges(append(sepActual.Slice(), sepConst...))
										// log.Printf("Testing SSep: %v of %v , Special Edges %v \n",
										// Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
										//log.Println("Sep const: ", sepConst, "sepChang ", sepChanging)
										// log.Println("SubSep: ")
										// for _, s := range sepSub.Edges {
										//  log.Println(s.Combination)
										// }
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										// log.Printf("No SubSep found for %v with Sp %v  \n",
										//  Graph{Edges: sepActualOrigin}, Sp)
										if addEdges {
											i_add++
											continue addingEdges
										} else {
											continue OUTER
										}
									}
								}
								// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: sepActual},
								//    Graph{Edges: sepActualOrigin}, Sp)
								continue subEdges
							}

							if addEdges {
								i_add++
								continue addingEdges
							} else {
								continue OUTER
							}
						}

						//d.Cache.AddPositive(sepActual, comps[i])

						// log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}

					return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}

func (d *DetKDecomp) findDecompUpdate(H Graph, oldSep []int, savedScenes map[uint32]SceneValue) Decomp {
	//Check current scenario for saved scene
	// usingScene := false
	// usingSep := Edges{}
	// for i := range savedScenes {
	//  if Equiv(savedScenes[i].Sub.Vertices(), H.Vertices()) {
	//      usingScene = true
	//      usingBag = savedScenes[i].Sep
	//      log.Println("Using saved scene!")
	//      break
	//  }
	// }

	verticesCurrent := append(H.Vertices())
	verticesExtended := append(verticesCurrent, oldSep...)
	conn := Inter(oldSep, verticesCurrent)
	compVertices := Diff(verticesCurrent, oldSep)
	bound := FilterVertices(d.Graph.Edges, conn)

	// log.Printf("\n\nDU Current oldSep: %v, Conn: %v\n", PrintVertices(oldSep), PrintVertices(conn))
	// log.Printf("DU Current SubGraph: %v ( %v hash) \n", H, H.Edges.Hash())
	// log.Printf("DU Current SubGraph: %v ( %v edges) (hash: %v )\n", H, H.Edges.Len(), H.Edges.Hash())
	// log.Printf("D Current Special Edges: %v\n\n", Sp)
	// log.Println("DU Hedges ", H)
	// log.Println("DU Comp Vertices: ", PrintVertices(compVertices))

	// Base case if H <= K
	if H.Edges.Len() == 0 && len(H.Special) <= 1 {

		return baseCaseDetK(H)
	}

	gen := NewCover(d.K, conn, bound, H.Edges.Vertices())

OUTER:
	for gen.HasNext {

		hash := IntHash(verticesCurrent) // save hash to avoid recomputing it below
		val, ok := savedScenes[hash]

		if !val.Perm { // delete one-time cached scene from map
			delete(savedScenes, hash)
		}
		if !Subset(conn, val.Sep.Vertices()) {
			ok = false // ignore this choice of separator if it breaks connectedness
		}

		var sep Edges
		addEdges := false

		if !ok {
			out := gen.NextSubset()

			if out == -1 {
				if gen.HasNext {
					log.Panicln(" -1 but hasNext not false!")
				}
				continue
			}
			sep = GetSubset(bound, gen.Subset)

			//check if sep "makes some progress" into separating H
			if len(Inter(sep.Vertices(), compVertices)) == 0 {
				addEdges = true
			}

			if !Subset(conn, sep.Vertices()) {
				log.Panicln("Cover messed up! 137")
			}

		} else {
			sep = val.Sep
			// log.Println("Using scene: ", val)

			if log.Flags() == 0 {
				if counter, ok := counterMap[val.String()]; ok {
					counterMap[val.String()] = counter + 1
				} else {
					counterMap[val.String()] = 1
				}
			}

		}

		if !addEdges || d.K-sep.Len() > 0 {
			i_add := 0

		addingEdges:
			for !addEdges || i_add < H.Edges.Len() {
				var sepActual Edges

				if addEdges {
					sepActual = NewEdges(append(sep.Slice(), H.Edges.Slice()[i_add]))
				} else {
					sepActual = sep
				}

				// sepActualOrigin := sepActual
				var sepSub *SepSub
				var sepConst []Edge
				var sepChanging []Edge
				if d.SubEdge {
					for i, v := range gen.Subset {
						if gen.InComp[v] {
							sepChanging = append(sepChanging, sep.Slice()[i])
						} else {
							sepConst = append(sepConst, sep.Slice()[i])
						}
					}
					if addEdges {
						sepChanging = append(sepChanging, H.Edges.Slice()[i_add])
					}
				}

			subEdges:
				for true {

					// log.Println("Sep chosen ", sepActual)

					// if usingScene {
					//  sep := NewEdges([]Edge{Edge{Vertices: usingBag}})
					//  comps, _, _ = H.GetComponents(sep, []Special{})
					// } else {
					//
					// }
					comps, _, _ := H.GetComponents(sepActual)

					//check chache for previous encounters
					if d.Cache.CheckNegative(sep, comps) {
						if addEdges {
							i_add++
							continue addingEdges
						} else {
							continue OUTER
						}
					}

					var subtrees []Node
					bag := Inter(sepActual.Vertices(), verticesExtended)

					for i := range comps {

						decomp := d.findDecompUpdate(comps[i], bag, savedScenes)
						if reflect.DeepEqual(decomp, Decomp{}) {

							d.Cache.AddNegative(sepActual, comps[i])
							// log.Printf("DU detK REJECTING %v: couldn't decompose %v  \n",
							//        Graph{Edges: sepActual}, comps[i])
							// log.Printf("\n\nDU Current oldSep: %v\n", PrintVertices(oldSep))
							// log.Printf("DU Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len(), H.Edges.Hash())
							// log.Printf("Current Special Edges: %v\n\n", Sp)

							if d.SubEdge {
								if sepSub == nil {
									sepSub = GetSepSub(d.Graph.Edges, NewEdges(sepChanging), d.K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										sepActual = NewEdges(append(sepActual.Slice(), sepConst...))
										// log.Printf("Testing SSep: %v of %v , Special Edges %v \n",
										//        Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
										//log.Println("Sep const: ", sepConst, "sepChang ", sepChanging)
										// log.Println("SubSep: ")
										// for _, s := range sepSub.Edges {
										//  log.Println(s.Combination)
										// }
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										// log.Printf("No SubSep found for %v with Sp %v  \n",
										//        Graph{Edges: sepActualOrigin}, Sp)
										if addEdges {
											i_add++
											continue addingEdges
										} else {
											continue OUTER
										}
									}
								}
								// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: sepActual},
								//        Graph{Edges: sepActualOrigin}, Sp)
								continue subEdges
							}

							if addEdges {
								i_add++
								continue addingEdges
							} else {
								continue OUTER
							}
						}

						//d.Cache.AddPositive(sepActual, comps[i])

						// log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}

					return Decomp{Graph: H, Root: Node{Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}
