// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to
// DetKDecomp

package algorithms

import (
	"container/heap"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strconv"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

type JCostBalDetKDecomp struct {
	Graph     Graph
	BalFactor int
	Depth     int // how many rounds of balSep are used
	JCosts    EdgesCostMap
}

/*func decrease(count int) int {
	output := count - 1
	if output < 1 {
		return 0
	} else {
		return output
	}
}*/

func (b JCostBalDetKDecomp) findDecompBalSep(K int, currentDepth int, H Graph, Sp []Special) Decomp {
	// log.Println("Current Depth: ", (b.Depth - currentDepth))
	// log.Printf("Current SubGraph: %+v\n", H)
	// log.Printf("Current Special Edges: %+v\n\n", Sp)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(b.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	var balsep Edges

	//find a balanced separator
	var decomposed = false
	edges := CutEdges(b.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), true)
	var subtrees []Decomp

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	// collect separators
	var seps [][]int
	var found []int
	parallelSearch(H, Sp, edges, &found, generators, b.BalFactor)
	if len(found) == 0 { // meaning that the search above never found anything
		log.Printf("balDet REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
		return Decomp{}
	}
	for i := 0; len(found) != 0; i++ {
		for _, g := range generators {
			fmt.Print(*g, " ")
		}
		fmt.Println()
		fmt.Println("found=", found)
		seps = append(seps, make([]int, len(found)))
		copy(seps[i], found)
		parallelSearch(H, Sp, edges, &found, generators, b.BalFactor)
	}

	// populate heap
	jh := make(JoinHeap, len(seps))
	for i, s := range seps {
		cost := b.JCosts.Cost(s) // TODO is s in the correct encoding?
		jh[i] = &Separator{
			EdgeComb: s,
			Cost:     cost,
		}
	}
	heap.Init(&jh)
	for jh.Len() > 0 {
		fmt.Println(heap.Pop(&jh).(*Separator).EdgeComb, heap.Pop(&jh).(*Separator).Cost)
	}
	return Decomp{}

	//find a balanced separator
OUTER:
	for !decomposed {
		if jh.Len() > 0 {
			found = heap.Pop(&jh).(*Separator).EdgeComb
		} else { // meaning that the search above never found anything
			log.Printf("balDet REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
			return Decomp{}
		}

		//wait until first worker finds a balanced sep
		balsep = GetSubset(edges, found)
		//  balsepOrig := balsep
		var sepSub *SepSub

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

	INNER:
		for !decomposed {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			// log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			ch := make(chan Decomp)
			//var outDecomp []Decomp
			for i := range comps {

				if currentDepth > 0 {
					go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {
						ch <- b.findDecompBalSep(K, decrease(currentDepth), comps[i], append(compsSp[i], SepSpecial))
					}(K, i, comps, compsSp, SepSpecial)
				} else {
					go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special) {

						// Base case handling

						Sp := append(compsSp[i], SepSpecial)
						//stop if there are at most two special edges left
						if comps[i].Edges.Len()+len(Sp) <= 2 {
							ch <- baseCaseSmart(b.Graph, comps[i], Sp)
							//outDecomp = append(outDecomp, baseCaseSmart(b.Graph, comps[i], Sp))
							return
						}

						//Early termination
						if comps[i].Edges.Len() <= K && len(Sp) == 1 {
							ch <- earlyTermination(comps[i], Sp[0])
							//outDecomp = append(outDecomp, earlyTermination(comps[i], Sp[0]))
							return
						}

						det := DetKDecomp{Graph: b.Graph, BalFactor: b.BalFactor, SubEdge: true}

						// edgesFromSpecial := EdgesSpecial(Sp)
						// comps[i].Edges.Append(edgesFromSpecial...)

						det.cache = make(map[uint64]*CompCache)
						result := det.findDecomp(K, comps[i], balsep.Vertices(), compsSp[i])
						if !reflect.DeepEqual(result, Decomp{}) && currentDepth == 0 {
							result.SkipRerooting = true
						} else {
							// res2 := b.findDecompBalSep(K, 1000, comps[i], append(compsSp[i], SepSpecial))
							// if !reflect.DeepEqual(res2, Decomp{}) {
							//  fmt.Println("Result, ", res2)
							//  fmt.Println("H: ", comps[i], "Sp ", compsSp, "balsep ", balsep)
							//  log.Panicln("Something is rotten in the state of this program")

							// }
						}
						ch <- result
						// outDecomp = append(outDecomp, result)
					}(K, i, comps, compsSp, SepSpecial)
				}

			}

			for i := 0; i < len(comps); i++ {
				//  decomp := outDecomp[i]
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					// log.Printf("balDet REJECTING %v: couldn't decompose a component of H %v \n",
					//        Graph{Edges: balsep}, H)
					// log.Println("\n\nCurrent Depth: ", (b.Depth - currentDepth))
					// log.Printf("Current SubGraph: %+v\n", H)
					// log.Printf("Current Special Edges: %+v\n\n", Sp)

					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(b.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
							// 	Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							// 	log.Println(s.Combination)
							// }
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if H.CheckBalancedSep(balsep, Sp, b.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							//log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					//      log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()),
					//         " of ", balsepOrig, " , ", Sp)
					continue INNER
				}

				subtrees = append(subtrees, decomp)
			}

			decomposed = true
		}
	}

	output := Node{Bag: balsep.Vertices(), Cover: balsep}

	for _, s := range subtrees {
		//TODO: Reroot only after all subtrees received
		if currentDepth == 0 && s.SkipRerooting {
			// log.Println("\nFrom detK on", decomp.Graph, ":\n", decomp)
			// local := BalSepGlobal{Graph: b.Graph, BalFactor: b.BalFactor}
			// decomp_deux := local.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
			// fmt.Println("Output from Balsep: ", decomp_deux)
		} else {
			s.Root = s.Root.Reroot(Node{Bag: balsep.Vertices(), Cover: balsep})
			s.Root = s.Root.Children[0]
			// log.Printf("Produced Decomp (with balsep %v): %+v\n", balsep, decomp)
		}

		output.Children = append(output.Children, s.Root)
	}

	return Decomp{Graph: H, Root: output}

}

func (b JCostBalDetKDecomp) FindGHD(K int, currentGraph Graph, Sp []Special) Decomp {
	return b.findDecompBalSep(K, b.Depth, currentGraph, Sp)
}

func (b JCostBalDetKDecomp) FindDecomp(K int) Decomp {
	return b.FindGHD(K, b.Graph, []Special{})
}

func (b JCostBalDetKDecomp) FindDecompGraph(G Graph, K int) Decomp {
	return b.FindGHD(K, G, []Special{})
}

func (b JCostBalDetKDecomp) FindDecompUpdate(K int, currentGraph Graph, savedScenes map[uint32]SceneValue) Decomp {
	// b.cache = make(map[uint32]*CompCache)

	if log.Flags() == 0 {
		counterMap = make(map[string]int)
		defer func(map[string]int) {

			fmt.Println("Counter Map:")

			for k, v := range counterMap {
				fmt.Println("Scene: ", k, "\nTimes Used: ", v)
			}

		}(counterMap)
	}

	return b.findDecompBalSepUpdate(K, b.Depth, currentGraph, []Special{}, savedScenes)
}

func (b JCostBalDetKDecomp) findDecompBalSepUpdate(K int, currentDepth int, H Graph, Sp []Special,
	savedScenes map[uint32]SceneValue) Decomp {
	// log.Println("Current Depth: ", (b.Depth - currentDepth))
	// log.Printf("Current SubGraph: %+v\n", H)
	// log.Printf("Current Special Edges: %+v\n\n", Sp)
	verticesCurrent := append(H.Vertices(), VerticesSpecial(Sp)...)

	//stop if there are at most two special edges left
	if H.Edges.Len()+len(Sp) <= 2 {
		return baseCaseSmart(b.Graph, H, Sp)
	}

	//Early termination
	if H.Edges.Len() <= K && len(Sp) == 1 {
		return earlyTermination(H, Sp[0])
	}

	var balsep Edges

	//find a balanced separator
	var decomposed = false
	edges := CutEdges(b.Graph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

	generators := SplitCombin(edges.Len(), K, runtime.GOMAXPROCS(-1), true)
	var subtrees []Decomp

	var cache map[uint32]struct{}
	cache = make(map[uint32]struct{})

	//find a balanced separator
OUTER:
	for !decomposed {

		hash := IntHash(verticesCurrent) // save hash to avoid recomputing it below
		val, ok := savedScenes[hash]

		if !val.Perm { // delete one-time cached scene from map
			delete(savedScenes, hash)
		}
		// if !Subset(conn, val.Sep.Vertices()) {
		//  ok = false // ignore this choice of separator if it breaks connectedness
		// }

		if !ok {
			var found []int

			//g.startSearchSimple(&found, &generator, result, input, &wg)
			parallelSearch(H, Sp, edges, &found, generators, b.BalFactor)

			if len(found) == 0 { // meaning that the search above never found anything
				log.Printf("balDet REJECT: Couldn't find balsep for H %v SP %v\n", H, Sp)
				return Decomp{}
			}

			//wait until first worker finds a balanced sep
			balsep = GetSubset(edges, found)
			//  balsepOrig := balsep

		} else {
			balsep = val.Sep
			// log.Println("Using scene: ", val)

			if log.Flags() == 0 {
				if counter, ok := counterMap[val.String()]; ok {
					counterMap[val.String()] = counter + 1
				} else {
					counterMap[val.String()] = 1
				}
			}

		}

		var sepSub *SepSub

		// log.Printf("Balanced Sep chosen: %+v\n", Graph{Edges: balsep})

	INNER:
		for !decomposed {
			comps, compsSp, _, _ := H.GetComponents(balsep, Sp)

			// log.Printf("Comps of Sep: %+v\n", comps)

			SepSpecial := Special{Edges: balsep, Vertices: balsep.Vertices()}

			ch := make(chan Decomp)
			//var outDecomp []Decomp
			for i := range comps {

				if currentDepth > 0 {
					go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special,
						savedScenes map[uint32]SceneValue) {
						ch <- b.findDecompBalSepUpdate(K, decrease(currentDepth), comps[i], append(compsSp[i],
							SepSpecial), savedScenes)
					}(K, i, comps, compsSp, SepSpecial, savedScenes)
				} else {
					go func(K int, i int, comps []Graph, compsSp [][]Special, SepSpecial Special,
						savedScenes map[uint32]SceneValue) {

						// Base case handling

						Sp := append(compsSp[i], SepSpecial)
						//stop if there are at most two special edges left
						if comps[i].Edges.Len()+len(Sp) <= 2 {
							ch <- baseCaseSmart(b.Graph, comps[i], Sp)
							//outDecomp = append(outDecomp, baseCaseSmart(b.Graph, comps[i], Sp))
							return
						}

						//Early termination
						if comps[i].Edges.Len() <= K && len(Sp) == 1 {
							ch <- earlyTermination(comps[i], Sp[0])
							//outDecomp = append(outDecomp, earlyTermination(comps[i], Sp[0]))
							return
						}

						det := DetKDecomp{Graph: b.Graph, BalFactor: b.BalFactor, SubEdge: true}

						// edgesFromSpecial := EdgesSpecial(Sp)
						// comps[i].Edges.Append(edgesFromSpecial...)

						det.cache = make(map[uint64]*CompCache)
						result := det.findDecompUpdate(K, comps[i], balsep.Vertices(), compsSp[i], savedScenes)
						if !reflect.DeepEqual(result, Decomp{}) {
							result.SkipRerooting = true
						} else {
							// res2 := b.findDecompBalSep(K, 1000, comps[i], append(compsSp[i], SepSpecial))
							// if !reflect.DeepEqual(res2, Decomp{}) {
							//  fmt.Println("Result, ", res2)
							//  fmt.Println("H: ", comps[i], "Sp ", compsSp, "balsep ", balsep)
							//  log.Panicln("Something is rotten in the state of this program")

							// }
						}
						ch <- result
						// outDecomp = append(outDecomp, result)
					}(K, i, comps, compsSp, SepSpecial, savedScenes)
				}

			}

			for i := 0; i < len(comps); i++ {
				//  decomp := outDecomp[i]
				decomp := <-ch
				if reflect.DeepEqual(decomp, Decomp{}) {
					// log.Printf("balDet REJECTING %v: couldn't decompose a component of H %v \n",
					//        Graph{Edges: balsep}, H)
					// log.Println("\n\nCurrent Depth: ", (b.Depth - currentDepth))
					// log.Printf("Current SubGraph: %+v\n", H)
					// log.Printf("Current Special Edges: %+v\n\n", Sp)

					subtrees = []Decomp{}
					if sepSub == nil {
						sepSub = GetSepSub(b.Graph.Edges, balsep, K)
					}
					nextBalsepFound := false
				thisLoop:
					for !nextBalsepFound {
						if sepSub.HasNext() {
							balsep = sepSub.GetCurrent()
							// log.Printf("Testing SSep: %v of %v , Special Edges %v \n", Graph{Edges: balsep},
							//        Graph{Edges: balsepOrig}, Sp)
							// log.Println("SubSep: ")
							// for _, s := range sepSub.Edges {
							//  log.Println(s.Combination)
							// }
							_, ok := cache[IntHash(balsep.Vertices())]
							if ok { //skip since already seen
								continue thisLoop
							}
							if H.CheckBalancedSep(balsep, Sp, b.BalFactor) {
								cache[IntHash(balsep.Vertices())] = Empty
								nextBalsepFound = true
							}
						} else {
							//      log.Printf("No SubSep found for %v with Sp %v  \n", Graph{Edges: balsepOrig}, Sp)
							continue OUTER
						}
					}
					// log.Println("Sub Sep chosen: ", balsep, "Vertices: ", PrintVertices(balsep.Vertices()), " of ",
					//     balsepOrig, " , ", Sp)
					continue INNER
				}

				//TODO: Reroot only after all subtrees received
				if currentDepth == 0 && decomp.SkipRerooting {
					// log.Println("\nFrom detK on", decomp.Graph, ":\n", decomp)
					// local := BalSepGlobal{Graph: b.Graph, BalFactor: b.BalFactor}
					// decomp_deux := local.findDecomp(K, comps[i], append(compsSp[i], SepSpecial))
					// fmt.Println("Output from Balsep: ", decomp_deux)
				} else {
					decomp.Root = decomp.Root.Reroot(Node{Bag: balsep.Vertices(), Cover: balsep})
					decomp.Root = decomp.Root.Children[0]
					// log.Printf("Produced Decomp (with balsep %v): %+v\n", balsep, decomp)
				}

				subtrees = append(subtrees, decomp)
			}

			decomposed = true
		}
	}

	output := Node{Bag: balsep.Vertices(), Cover: balsep}

	for _, s := range subtrees {
		output.Children = append(output.Children, s.Root)
	}

	return Decomp{Graph: H, Root: output}

}

func (b JCostBalDetKDecomp) Name() string {
	return "BalSep / DetK - Hybrid with Depth " + strconv.Itoa(b.Depth+1) + " + Join Costs"
}
