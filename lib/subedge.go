package lib

import (
	"fmt"
	"io/ioutil"
	"sort"
)

type subSet struct {
	source  []int
	current CombinationIterator
}

func getSubsetIterator(vertices []int) *subSet {
	var output subSet

	//fmt.Println("Vertices", Edge{Vertices: vertices})
	output = subSet{source: vertices, current: GetCombin(len(vertices), len(vertices))}
	return &output
}

func (s *subSet) hasNext() bool {
	return s.current.HasNext()
}

func getEdge(vertices []int, s []int) Edge {
	var output Edge

	for _, i := range s {
		output.Vertices = append(output.Vertices, vertices[i])
	}

	return output
}

func (s *subSet) getCurrent() Edge {
	s.current.Confirm()

	return getEdge(s.source, s.current.Combination)
}

//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------

type SubEdges struct {
	k             int
	initial       Edge
	source        Edges
	current       Edge
	gen           *CombinationIterator
	combination   []int
	currentSubset *subSet
	cache         map[uint32]struct{}
	emptyReturned bool
}

func getSubEdgeIterator(edges Edges, e Edge, k int) SubEdges {
	var h_edges []Edge

	for j := range edges.Slice() {
		inter := Inter(edges.Slice()[j].Vertices, e.Vertices)
		if len(inter) > 0 && len(inter) < len(e.Vertices) {
			h_edges = append(h_edges, Edge{Vertices: inter})
		}
	}

	// fmt.Println("Neighbourhood: ")
	// for i := range h_edges {
	// 	fmt.Println(h_edges[i].FullString())
	// }
	// fmt.Println("\n")

	// TODO: Sort h_edges by size
	sort.Slice(h_edges, func(i, j int) bool { return len(h_edges[i].Vertices) > len(h_edges[j].Vertices) })

	source := removeDuplicateEdges(h_edges)
	//fmt.Println("h_edges", h_edges)
	var output SubEdges
	output.cache = make(map[uint32]struct{})

	//sort.Slice(h_edges, func(i, j int) bool { return len(h_edges[i].Vertices) > len(h_edges[j].Vertices) })
	output.source = source
	if k > output.source.Len() {
		k = output.source.Len()
	}
	// fmt.Println("k", k)
	tmp := GetCombinUnextend(output.source.Len(), k)
	output.gen = &tmp
	output.current = e
	output.initial = e
	output.k = k
	output.combination = make([]int, k)
	output.cache[IntHash(edges.Vertices())] = Empty // initial cache
	//output.cache = append(output.cache, Vertices(edges))

	return output
}

func (s *SubEdges) reset() {
	// fmt.Println("Reset")
	tmp := GetCombinUnextend(s.source.Len(), s.k)
	s.gen = &tmp

	s.currentSubset = nil
	s.current = s.initial
	s.emptyReturned = false
}

// This checks whether the current edge has a more tuples to intersect with,
// and create a new vertex set
func (s *SubEdges) hasNextCombination() bool {

	if !s.gen.HasNext() {
		return false
	}
	s.gen.Confirm()
	copy(s.combination, s.gen.Combination)

	return true
}

func (s SubEdges) existsSubset(b []int) bool {
	hashOfB := IntHash(b)
	if _, ok := s.cache[hashOfB]; ok {
		return true
	}

	return false
}

func (s *SubEdges) hasNext() bool {
	newSelected := false
	if s.currentSubset == nil || !s.currentSubset.hasNext() {
		for s.hasNextCombination() {
			// fmt.Println("We need a new subset")
			// fmt.Println("current:", GetSubset(s.source, s.Combination))
			edges := GetSubset(s.source, s.combination)
			vertices := edges.Vertices()
			if len(vertices) == 0 || s.existsSubset(vertices) {
				continue //skip
			} else {
				//s.cache = append(s.cache, vertices)
				s.currentSubset = getSubsetIterator(vertices)
				s.currentSubset.hasNext()
				newSelected = true
				break
			}

		}
		if !newSelected && !s.hasNextCombination() {
			if !s.emptyReturned {
				s.emptyReturned = true
				return true
			}
			return false
		}
	}

	s.current = s.currentSubset.getCurrent()
	if s.existsSubset(s.current.Vertices) || len(s.current.Vertices) == s.source.Len() {
		return s.hasNext()
	} else {
		s.cache[IntHash(s.current.Vertices)] = Empty // add used combination to cache
	}

	return true
}

func (s SubEdges) getCurrent() Edge {
	if s.emptyReturned {
		return Edge{Vertices: []int{}}
	}
	return s.current
}

//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------

type SepSub struct {
	Edges []SubEdges

	cache map[uint32]struct{}
}

func GetSepSub(edges Edges, sep Edges, k int) *SepSub {
	var output SepSub
	output.cache = make(map[uint32]struct{})

	for _, e := range sep.Slice() {
		output.Edges = append(output.Edges, getSubEdgeIterator(edges, e, k))
	}

	return &output
}

func (sep *SepSub) HasNext() bool {
	i := 0

	// fmt.Println("len", len(sep.Edges))
	for i < len(sep.Edges) {
		if sep.Edges[i].hasNext() {
			// if sep.alreadyChecked() {
			// 	return sep.HasNext()
			// }
			return true
		} else {
			sep.Edges[i].reset()
			i++
		}

		// fmt.Println("i", i)
	}

	return false
}

func (sep *SepSub) alreadyChecked() bool {
	currentEdges := sep.GetCurrent()
	currentVertices := currentEdges.Vertices()
	hashOfV := IntHash(currentVertices)
	if _, ok := sep.cache[hashOfV]; ok {
		return true
	}
	sep.cache[hashOfV] = Empty // add new vertex set to cache
	return false

}

func (sep SepSub) GetCurrent() Edges {
	var output []Edge

	for _, s := range sep.Edges {
		output = append(output, s.getCurrent())
	}

	return NewEdges(output)
}

// TEST

func check(e error) {
	if e != nil {
		panic(e)
	}

}

func test() {
	dat, err := ioutil.ReadFile("/home/cem/Desktop/scripts/BalancedGo/hypergraphs/rand_q0037.hg")
	check(err)

	parsedGraph, parse := GetGraph(string(dat))

	e1 := parse.GetEdge("54020(1,2,5,6,7,8,9,10,12,14,17,18,19,22,23,25)")
	fmt.Println(e1.FullString(), "\n\n")

	for _, e := range parsedGraph.Edges.Slice() {
		fmt.Println(e.FullString())
	}

	test := GetSepSub(parsedGraph.Edges, NewEdges([]Edge{e1}), 2)
	count := 1
	fmt.Println(test.GetCurrent())
	for test.HasNext() {
		// fmt.Println(test.GetCurrent())
		count++
	}

	fmt.Println("\n\n Tested ", count, " many subedges")
	return

	// fmt.Println("Subset test: ")
	// fmt.Println("========================================")

	// test := getSubsetIterator([]Edge{Edge{Vertices: []int{1, 2, 3, 4}}})

	// for test.HasNext() {
	// 	fmt.Println(test.getCurrent())
	// }

	// fmt.Println("Subegde test: ")
	// fmt.Println("========================================")

	// test2 := getSubEdgeIterator([]Edge{Edge{Vertices: []int{1, 2, 3, 4}}, Edge{Vertices: []int{1, 2, 5, 6}}}, Edge{Vertices: []int{5, 8, 2, 9}}, 2)

	// fmt.Println("begin", test2.getCurrent())
	// for test2.HasNext() {
	// 	fmt.Println("now", test2.getCurrent())
	// }
	// test2.reset()
	// fmt.Println("begin", test2.getCurrent())
	// for test2.HasNext() {
	// 	fmt.Println("now", test2.getCurrent())
	// }

	// fmt.Println("SepSub test: ")
	// fmt.Println("========================================")

	// test3 := GetSepSub([]Edge{Edge{Vertices: []int{5, 8, 2, 9}}, Edge{Vertices: []int{1, 2, 3, 4}}, Edge{Vertices: []int{1, 2, 5, 6}}, Edge{Vertices: []int{9, 12, 15, 16}}, Edge{Vertices: []int{16, 112, 115, 116}}}, []Edge{Edge{Vertices: []int{5, 8, 2, 9}}, Edge{Vertices: []int{9, 12, 15, 16}}}, 2)

	// fmt.Println("begin", test3.GetCurrent())
	// for test3.HasNext() {
	// 	fmt.Println("now", test3.GetCurrent())
	// }

}
