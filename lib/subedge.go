package lib

import (
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
	//  fmt.Println(h_edges[i].FullString())
	// }
	// fmt.Println("\n")

	// TODO: Sort h_edges by size

	source := removeDuplicateEdges(h_edges)

	sort.Slice(h_edges, func(i, j int) bool { return len(h_edges[i].Vertices) > len(h_edges[j].Vertices) })
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
	_, ok := s.cache[IntHash(b)]

	return ok
}

func (s *SubEdges) hasNext() bool {
	newSelected := false
	if s.currentSubset == nil || !s.currentSubset.hasNext() {
		for s.hasNextCombination() {
			// fmt.Println("We need a new subset")
			// fmt.Println("current:", GetSubset(s.source, s.Combination))
			edges := GetSubset(s.source, s.combination)
			vertices := edges.Vertices()
			if len(vertices) == 0 || len(vertices) == s.source.Len() || s.existsSubset(vertices) {
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

	s.cache[IntHash(s.current.Vertices)] = Empty // add used combination to cache

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

	// cache map[uint32]struct{}
}

func GetSepSub(edges Edges, sep Edges, k int) *SepSub {
	var output SepSub
	// output.cache = make(map[uint32]struct{})

	encountered := make(map[int]struct{})
	var Empty struct{}

	var sepIntersectFree []Edge

	for i := range sep.Slice() {
		tmpEdge := make([]int, 0, len(sep.Slice()[i].Vertices))

		for _, j := range sep.Slice()[i].Vertices {

			if _, ok := encountered[j]; !ok {

				encountered[j] = Empty
				tmpEdge = append(tmpEdge, j)

			}
		}

		sepIntersectFree = append(sepIntersectFree, Edge{Name: sep.Slice()[i].Name, Vertices: tmpEdge})

	}

	newSep := NewEdges(sepIntersectFree)

	for i := range newSep.Slice() {
		output.Edges = append(output.Edges, getSubEdgeIterator(edges, newSep.Slice()[i], k))
	}

	return &output
}

func (sep *SepSub) HasNext() bool {
	i := 0

	// fmt.Println("len", len(sep.Edges))
	for i < len(sep.Edges) {
		if sep.Edges[i].hasNext() {
			// if sep.alreadyChecked() {
			//  return sep.HasNext()
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

func (sep SepSub) GetCurrent() Edges {
	var output []Edge

	for _, s := range sep.Edges {
		output = append(output, s.getCurrent())
	}

	return NewEdges(output)
}
