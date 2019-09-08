package lib

import (
	"fmt"
	"io/ioutil"
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
	cache         [][]int
	emptyReturned bool
}

func getSubEdgeIterator(edges Edges, e Edge, k int) SubEdges {
	var h_edges Edges

	for _, j := range edges.Slice() {
		inter := Inter(j.Vertices, e.Vertices)
		if len(inter) > 0 && len(inter) < len(e.Vertices) {
			h_edges.Append(Edge{Vertices: inter})
		}
	}
	// TODO: Sort h_edges by size
	//fmt.Println("h_edges", h_edges)

	var output SubEdges

	//sort.Slice(h_edges, func(i, j int) bool { return len(h_edges[i].Vertices) > len(h_edges[j].Vertices) })
	output.source = h_edges
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
	for _, e := range s.cache {
		if Subset(b, e) && Subset(e, b) {
			return true
		}
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
			if s.existsSubset(vertices) || len(vertices) == 0 { // || len(vertices) == len(Vertices(s.source))
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
	if s.existsSubset(s.current.Vertices) {
		return s.hasNext()
	} else {
		s.cache = append(s.cache, s.current.Vertices)
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
}

func GetSepSub(edges Edges, sep Edges, k int) *SepSub {
	var output SepSub

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
			//fmt.Println("increased subedge ", i)
			return true
		} else {
			sep.Edges[i].reset()
			i++
		}

		// fmt.Println("i", i)
	}

	return false
}

func (sep SepSub) GetCurrent() Edges {
	var output Edges

	for _, s := range sep.Edges {
		output.Append(s.getCurrent())
	}

	return output
}

// TEST

func check(e error) {
	if e != nil {
		panic(e)
	}

}

func test() {
	dat, err := ioutil.ReadFile("/home/cem/Desktop/scripts/BalancedGo/hypergraphs/adlerexample_badorder.hg")
	check(err)

	parsedGraph, parse := GetGraph(string(dat))

	e1 := parse.GetEdge("E1 (V1, V2, V9)")
	fmt.Println(e1.FullString(), "\n\n")

	for _, e := range parsedGraph.Edges.Slice() {
		fmt.Println(e.FullString())
	}

	test := GetSepSub(parsedGraph.Edges, NewEdges([]Edge{e1}), 2)

	fmt.Println(test.GetCurrent())
	for test.HasNext() {
		fmt.Println(test.GetCurrent())
	}
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
