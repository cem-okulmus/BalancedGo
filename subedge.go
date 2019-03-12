package main

import "fmt"

type Subset struct {
	source  []int
	current Combin
	m       map[int]string
}

func getSubsetIterator(vertices []int, m map[int]string) *Subset {
	var output Subset

	//fmt.Println("Vertices", Edge{nodes: vertices, m: edges[0].m})
	output = Subset{source: vertices, current: getCombin(len(vertices), len(vertices)), m: m}
	return &output
}

func (s *Subset) hasNext() bool {
	return s.current.hasNext()
}

func getEdge(nodes []int, s []int, m map[int]string) Edge {
	var output Edge
	output.m = m
	for _, i := range s {
		output.nodes = append(output.nodes, nodes[i])
	}
	output.nodes = removeDuplicates(output.nodes)
	return output
}

func (s *Subset) getCurrent() Edge {
	s.current.confirm()

	return getEdge(s.source, s.current.combination, s.m)
}

//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------

type SubEdges struct {
	k             int
	initial       Edge
	source        Edges
	current       Edge
	gen           *CombinationGenerator
	combination   []int
	currentSubset *Subset
}

func getSubEdgeIterator(edges Edges, e Edge, k int) SubEdges {
	var h_edges Edges

	for _, j := range edges {
		inter := inter(j.nodes, e.nodes)
		if len(inter) > 0 && len(inter) < len(e.nodes) {
			h_edges.append(Edge{nodes: inter, m: e.m})
		}
	}
	// TODO: Sort h_edges by size
	//fmt.Println("h_edges", h_edges)

	var output SubEdges

	output.source = h_edges
	if k > len(output.source) {
		k = len(output.source)
	}
	// fmt.Println("k", k)
	output.gen = NewCombinationGenerator(len(output.source), k)
	output.current = e
	output.initial = e
	output.k = k
	output.combination = make([]int, k)

	return output
}

func (s *SubEdges) reset() {
	// fmt.Println("Reset")
	s.gen = NewCombinationGenerator(len(s.source), s.k)
	s.currentSubset = nil
	s.current = s.initial
}

// This checks whether the current edge has a more tuples to intersect with,
// and create a new vertex set
func (s *SubEdges) hasNextCombination() bool {
	if !s.gen.Next() {
		return false
	}

	s.gen.Combination(s.combination)

	return true
}

func (s *SubEdges) hasNext() bool {
	if s.currentSubset == nil || !s.currentSubset.hasNext() {
		if s.hasNextCombination() {
			// fmt.Println("We need a new subset")
			// fmt.Println("current:", getSubset(s.source, s.combination))
			edges := getSubset(s.source, s.combination)
			vertices := removeDuplicates(Vertices(edges))
			s.currentSubset = getSubsetIterator(vertices, s.initial.m)
			s.currentSubset.hasNext()
		} else {
			return false
		}
	}

	s.current = s.currentSubset.getCurrent()
	return true
}

func (s SubEdges) getCurrent() Edge {
	return s.current
}

//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------

type SepSub struct {
	edges []SubEdges
}

func getSepSub(edges Edges, sep Edges, k int) *SepSub {
	var output SepSub

	for _, e := range sep {
		output.edges = append(output.edges, getSubEdgeIterator(edges, e, k))
	}

	return &output
}

func (sep *SepSub) hasNext() bool {
	i := 0

	// fmt.Println("len", len(sep.edges))
	for i < len(sep.edges) {
		if sep.edges[i].hasNext() {
			//fmt.Println("increased subedge ", i)
			return true
		} else {
			sep.edges[i].reset()
			i++
		}

		// fmt.Println("i", i)
	}

	return false
}

func (sep SepSub) getCurrent() []Edge {
	var output Edges

	for _, s := range sep.edges {
		output.append(s.getCurrent())
	}

	return output
}

// TEST

func test() {

	// fmt.Println("Subset test: ")
	// fmt.Println("========================================")

	// test := getSubsetIterator([]Edge{Edge{nodes: []int{1, 2, 3, 4}}})

	// for test.hasNext() {
	// 	fmt.Println(test.getCurrent())
	// }

	// fmt.Println("Subegde test: ")
	// fmt.Println("========================================")

	// test2 := getSubEdgeIterator([]Edge{Edge{nodes: []int{1, 2, 3, 4}}, Edge{nodes: []int{1, 2, 5, 6}}}, Edge{nodes: []int{5, 8, 2, 9}}, 2)

	// fmt.Println("begin", test2.getCurrent())
	// for test2.hasNext() {
	// 	fmt.Println("now", test2.getCurrent())
	// }
	// test2.reset()
	// fmt.Println("begin", test2.getCurrent())
	// for test2.hasNext() {
	// 	fmt.Println("now", test2.getCurrent())
	// }

	fmt.Println("SepSub test: ")
	fmt.Println("========================================")

	test3 := getSepSub([]Edge{Edge{nodes: []int{5, 8, 2, 9}}, Edge{nodes: []int{1, 2, 3, 4}}, Edge{nodes: []int{1, 2, 5, 6}}, Edge{nodes: []int{9, 12, 15, 16}}, Edge{nodes: []int{16, 112, 115, 116}}}, []Edge{Edge{nodes: []int{5, 8, 2, 9}}, Edge{nodes: []int{9, 12, 15, 16}}}, 2)

	fmt.Println("begin", test3.getCurrent())
	for test3.hasNext() {
		fmt.Println("now", test3.getCurrent())
	}

}
