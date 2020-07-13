package algorithms

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type CompCache struct {
	Succ []uint32
	Fail []uint32
}

type DetKDecomp struct {
	Graph     Graph
	BalFactor int
	SubEdge   bool
	Divide    bool
	cache     map[uint32]*CompCache
	cacheMux  sync.RWMutex
}

func (d *DetKDecomp) addPositive(sep Edges, comp Graph) {
	d.cacheMux.Lock()
	d.cache[sep.Hash()].Succ = append(d.cache[sep.Hash()].Succ, comp.Edges.Hash())
	d.cacheMux.Unlock()
}

func (d *DetKDecomp) addNegative(sep Edges, comp Graph) {
	d.cacheMux.Lock()
	d.cache[sep.Hash()].Fail = append(d.cache[sep.Hash()].Fail, comp.Edges.Hash())
	d.cacheMux.Unlock()
}

func (d *DetKDecomp) checkNegative(sep Edges, comp Graph) bool {
	d.cacheMux.RLock()
	defer d.cacheMux.RUnlock()

	compCachePrev, _ := d.cache[sep.Hash()]
	for i := range compCachePrev.Fail {
		if comp.Edges.Hash() == compCachePrev.Fail[i] {
			//	log.Println("Comp ", comp, "(hash ", comp.Edges.Hash(), ")  known as negative for sep ", sep)
			return true
		}

	}

	return false
}

func (d *DetKDecomp) checkPositive(sep Edges, comp Graph) bool {
	d.cacheMux.RLock()
	defer d.cacheMux.RUnlock()

	compCachePrev, _ := d.cache[sep.Hash()]
	for i := range compCachePrev.Fail {
		if comp.Edges.Hash() == compCachePrev.Succ[i] {
			//	log.Println("Comp ", comp, " known as negative for sep ", sep)
			return true
		}

	}

	return false
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
func baseCaseDetK(H Graph, Sp []Special) Decomp {
	// log.Printf("Base case reached. Number of Special Edges %d\n", len(Sp))
	var children Node

	switch len(Sp) {
	case 0:
		return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges}}

	case 1:
		children = Node{Bag: Sp[0].Vertices, Cover: Sp[0].Edges}
	}

	if H.Edges.Len() == 0 {
		return Decomp{Graph: H, Root: children}
	}
	return Decomp{Graph: H, Root: Node{Bag: H.Vertices(), Cover: H.Edges, Children: []Node{children}}}
}

func (d *DetKDecomp) findDecomp(K int, H Graph, oldSep []int, Sp []Special) Decomp {
	verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
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
	if H.Edges.Len() == 0 && len(Sp) <= 1 {
		if d.Divide {
			out := baseCaseDetK(H, []Special{})
			out.Root.LowConnecting = true

			return out
		}
		return baseCaseDetK(H, Sp)
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

		if !Subset(conn, sep.Vertices()) {
			log.Panicln("Cover messed up! 137")
		}

		// log.Println("Next Cover ", sep)

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
					comps, compsSp, _ := H.GetComponents(sepActual, Sp)

					//check chache for previous encounters
					d.cacheMux.RLock()
					_, ok := d.cache[sepActual.Hash()]
					d.cacheMux.RUnlock()
					if !ok {
						var newCache CompCache
						d.cacheMux.Lock()
						d.cache[sepActual.Hash()] = &newCache
						d.cacheMux.Unlock()

					} else {
						for j := range comps {
							if d.checkNegative(sepActual, comps[j]) { //TODO: Add positive check and cutNodes
								//fmt.Println("Skipping a sep", sepActual)
								if addEdges {
									i_add++
									continue addingEdges
								} else {
									continue OUTER
								}
							}
						}
					}

					// log.Printf("Comps of Sep: %v, len: %v\n", comps, len(comps))

					var subtrees []Node
					bag := Inter(sepActual.Vertices(), verticesExtended)

					// log.Println("sep", sep, "\nsepActual", sepActual, "\n B of SepActual", PrintVertices(sepActual.Vertices()), "\noldSep ", PrintVertices(oldSep),
					// 	"\nvertices of C", PrintVertices(verticesCurrent), "\n\nunion o both", PrintVertices(verticesExtended), "\n bag: ", PrintVertices(bag))

					// for i := range sepActual.Vertices() {
					// 	if Mem(verticesCurrent, sepActual.Vertices()[i]) && !Mem(bag, sepActual.Vertices()[i]) {

					// 		fmt.Println("Another union: ", PrintVertices(append(oldSep, verticesCurrent...)))

					// 		fmt.Println("Another intersect: ", PrintVertices(Inter(sepActual.Vertices(), verticesExtended)))

					// 		fmt.Println("sep", sep, "\nsepActual", sepActual, "\n B of SepActual", PrintVertices(sepActual.Vertices()), "\noldSep ", PrintVertices(oldSep),
					// 			"\nvertices of C", PrintVertices(verticesCurrent), "\n\nunion o both", PrintVertices(verticesExtended), "\n bag: ", PrintVertices(bag))

					// 		log.Panicln("something is not right in the state of this program!")
					// 	}
					// }

					lowFlag := false
					for i := range comps {
						if comps[i].Edges.Len() == 0 && d.Divide {
							lowFlag = true //since special Edge would connect to current sep, if accepting
							continue
						}
						decomp := d.findDecomp(K, comps[i], bag, compsSp[i])
						if reflect.DeepEqual(decomp, Decomp{}) {
							//cache[sepActual.Hash()].Fail = append(cache[sepActual.Hash()].Fail, comps[i].Edges.Hash())
							d.addNegative(sepActual, comps[i])
							// log.Printf("detK REJECTING %v: couldn't decompose %v with SP %v \n", Graph{Edges: sepActual}, comps[i], compsSp[i])
							// log.Printf("\n\nCurrent oldSep: %v\n", PrintVertices(oldSep))
							// log.Printf("Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len(), H.Edges.Hash())
							// log.Printf("Current Special Edges: %v\n\n", Sp)

							if d.SubEdge {
								if sepSub == nil {
									sepSub = GetSepSub(d.Graph.Edges, NewEdges(sepChanging), K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										sepActual = NewEdges(append(sepActual.Slice(), sepConst...))
										// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
										//log.Println("Sep const: ", sepConst, "sepChang ", sepChanging)
										// log.Println("SubSep: ")
										// for _, s := range sepSub.Edges {
										// 	log.Println(s.Combination)
										// }
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: sepActualOrigin}, Sp)
										if addEdges {
											i_add++
											continue addingEdges
										} else {
											continue OUTER
										}
									}
								}
								// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
								continue subEdges
							}

							if addEdges {
								i_add++
								continue addingEdges
							} else {
								continue OUTER
							}
						}
						//cache[sepActual.Hash()].Succ = append(cache[sepActual.Hash()].Succ, comps[i].Edges.Hash())
						//d.addPositive(sepActual, comps[i])

						// log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}

					return Decomp{Graph: H, Root: Node{LowConnecting: lowFlag, Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}

func (d DetKDecomp) FindHD(K int, currentGraph Graph, Sp []Special) Decomp {
	d.cache = make(map[uint32]*CompCache)
	return d.findDecomp(K, currentGraph, []int{}, Sp)
}

func (d DetKDecomp) FindDecomp(K int) Decomp {
	return d.FindHD(K, d.Graph, []Special{})
}

func (d DetKDecomp) Name() string {
	if d.SubEdge {
		return "DetK with local BIP"
	} else {
		return "DetK"
	}
}

func (d DetKDecomp) FindDecompGraph(G Graph, K int) Decomp {
	return d.FindHD(K, G, []Special{})
}

var counterMap map[string]int

func (d DetKDecomp) FindDecompUpdate(K int, currentGraph Graph, savedScenes map[uint32]SceneValue) Decomp {
	d.cache = make(map[uint32]*CompCache)

	if log.Flags() == 0 {
		counterMap = make(map[string]int)
		defer func(map[string]int) {

			fmt.Println("Counter Map:")

			for k, v := range counterMap {
				fmt.Println("Scene: ", k, "\nTimes Used: ", v)
			}

		}(counterMap)
	}

	return d.findDecompUpdate(K, currentGraph, []int{}, []Special{}, savedScenes)
}

func (d *DetKDecomp) findDecompUpdate(K int, H Graph, oldSep []int, Sp []Special, savedScenes map[uint32]SceneValue) Decomp {

	//Check current scenario for saved scene
	// usingScene := false
	// usingSep := Edges{}
	// for i := range savedScenes {
	// 	if Equiv(savedScenes[i].Sub.Vertices(), H.Vertices()) {
	// 		usingScene = true
	// 		usingBag = savedScenes[i].Sep
	// 		log.Println("Using saved scene!")
	// 		break
	// 	}
	// }

	verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)
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
	if H.Edges.Len() == 0 && len(Sp) <= 1 {
		if d.Divide {
			out := baseCaseDetK(H, []Special{})
			out.Root.LowConnecting = true

			return out
		}
		return baseCaseDetK(H, Sp)
	}

	gen := NewCover(K, conn, bound, H.Edges)

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
					// 	sep := NewEdges([]Edge{Edge{Vertices: usingBag}})
					// 	comps, _, _ = H.GetComponents(sep, []Special{})
					// } else {
					//
					// }
					comps, compsSp, _ := H.GetComponents(sepActual, Sp)

					//check chache for previous encounters
					d.cacheMux.RLock()
					_, ok := d.cache[sepActual.Hash()]
					d.cacheMux.RUnlock()
					if !ok {
						var newCache CompCache
						d.cacheMux.Lock()
						d.cache[sepActual.Hash()] = &newCache
						d.cacheMux.Unlock()

					} else {
						for j := range comps {
							if d.checkNegative(sepActual, comps[j]) { //TODO: Add positive check and cutNodes
								//fmt.Println("Skipping a sep", sepActual)
								if addEdges {
									i_add++
									continue addingEdges
								} else {
									continue OUTER
								}
							}
						}
					}

					// log.Printf("Comps of Sep: %v, len: %v\n", comps, len(comps))

					var subtrees []Node
					bag := Inter(sepActual.Vertices(), verticesExtended)

					lowFlag := false
					for i := range comps {
						if comps[i].Edges.Len() == 0 && d.Divide {
							lowFlag = true //since special Edge would connect to current sep, if accepting
							continue
						}
						decomp := d.findDecompUpdate(K, comps[i], bag, compsSp[i], savedScenes)
						if reflect.DeepEqual(decomp, Decomp{}) {
							//cache[sepActual.Hash()].Fail = append(cache[sepActual.Hash()].Fail, comps[i].Edges.Hash())
							d.addNegative(sepActual, comps[i])
							// log.Printf("DU detK REJECTING %v: couldn't decompose %v  \n", Graph{Edges: sepActual}, comps[i])
							// log.Printf("\n\nDU Current oldSep: %v\n", PrintVertices(oldSep))
							// log.Printf("DU Current SubGraph: %v ( %v edges)\n", H, H.Edges.Len(), H.Edges.Hash())
							// log.Printf("Current Special Edges: %v\n\n", Sp)

							if d.SubEdge {
								if sepSub == nil {
									sepSub = GetSepSub(d.Graph.Edges, NewEdges(sepChanging), K)
								}

								nextBalsepFound := false

								for !nextBalsepFound {
									if sepSub.HasNext() {
										sepActual = sepSub.GetCurrent()
										sepActual = NewEdges(append(sepActual.Slice(), sepConst...))
										// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
										//log.Println("Sep const: ", sepConst, "sepChang ", sepChanging)
										// log.Println("SubSep: ")
										// for _, s := range sepSub.Edges {
										// 	log.Println(s.Combination)
										// }
										if connectingSep(sepActual.Vertices(), conn, compVertices) {
											nextBalsepFound = true
										}
									} else {
										// log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: sepActualOrigin}, Sp)
										if addEdges {
											i_add++
											continue addingEdges
										} else {
											continue OUTER
										}
									}
								}
								// log.Printf("Sub Sep chosen: %vof %v , %v \n", Graph{Edges: sepActual}, Graph{Edges: sepActualOrigin}, Sp)
								continue subEdges
							}

							if addEdges {
								i_add++
								continue addingEdges
							} else {
								continue OUTER
							}
						}
						//cache[sepActual.Hash()].Succ = append(cache[sepActual.Hash()].Succ, comps[i].Edges.Hash())
						//d.addPositive(sepActual, comps[i])

						// log.Printf("Produced Decomp: %v\n", decomp)
						subtrees = append(subtrees, decomp.Root)
					}

					return Decomp{Graph: H, Root: Node{LowConnecting: lowFlag, Bag: bag, Cover: sepActual, Children: subtrees}}
				}
			}

		}

	}

	return Decomp{} // Reject if no separator could be found
}
