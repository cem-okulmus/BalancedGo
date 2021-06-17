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

	output = subSet{source: vertices, current: getCombin(len(vertices), len(vertices))}
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

	return getEdge(s.source, s.current.combination)
}

type subEdges struct {
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

func getSubEdgeIterator(edges Edges, e Edge, k int) subEdges {
	var HEdges []Edge

	for j := range edges.Slice() {
		inter := Inter(edges.Slice()[j].Vertices, e.Vertices)
		if len(inter) > 0 && len(inter) < len(e.Vertices) {
			HEdges = append(HEdges, Edge{Vertices: inter})
		}
	}

	source := removeDuplicateEdges(HEdges)

	sort.Slice(HEdges, func(i, j int) bool { return len(HEdges[i].Vertices) > len(HEdges[j].Vertices) })
	var output subEdges
	output.cache = make(map[uint32]struct{})

	output.source = source
	if k > output.source.Len() {
		k = output.source.Len()
	}
	tmp := getCombinUnextend(output.source.Len(), k)
	output.gen = &tmp
	output.current = e
	output.initial = e
	output.k = k
	output.combination = make([]int, k)
	output.cache[IntHash(edges.Vertices())] = Empty // initial cache

	return output
}

func (s *subEdges) reset() {
	// fmt.Println("Reset")
	tmp := getCombinUnextend(s.source.Len(), s.k)
	s.gen = &tmp

	s.currentSubset = nil
	s.current = s.initial
	s.emptyReturned = false
}

// This checks whether the current edge has a more tuples to intersect with,
// and create a new vertex set
func (s *subEdges) hasNextCombination() bool {

	if !s.gen.HasNext() {
		return false
	}
	s.gen.Confirm()
	copy(s.combination, s.gen.combination)

	return true
}

func (s subEdges) existsSubset(b []int) bool {
	_, ok := s.cache[IntHash(b)]

	return ok
}

func (s *subEdges) hasNext() bool {
	newSelected := false
	if s.currentSubset == nil || !s.currentSubset.hasNext() {
		for s.hasNextCombination() {
			edges := GetSubset(s.source, s.combination)
			vertices := edges.Vertices()
			if len(vertices) == 0 || len(vertices) == s.source.Len() || s.existsSubset(vertices) {
				continue //skip
			} else {
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

func (s subEdges) getCurrent() Edge {
	if s.emptyReturned {
		return Edge{Vertices: []int{}}
	}
	return s.current
}

// SepSub is used to iterate over all subedge variants of a separator
type SepSub struct {
	edges []subEdges
}

// GetSepSub is a constructor for SepSub
func GetSepSub(edges Edges, sep Edges, k int) *SepSub {
	var output SepSub
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
		output.edges = append(output.edges, getSubEdgeIterator(edges, newSep.Slice()[i], k))
	}

	return &output
}

// HasNext checks if SepSub has more subedge variants to produce
func (sep *SepSub) HasNext() bool {
	for i := 0; i < len(sep.edges); i++ {
		if sep.edges[i].hasNext() {
			return true
		}
		if i < len(sep.edges)-1 {
			sep.edges[i].reset()
		}
	}

	return false
}

// GetCurrent is used to extract the current subedge variant
func (sep SepSub) GetCurrent() Edges {
	var output []Edge

	for _, s := range sep.edges {
		output = append(output, s.getCurrent())
	}

	return NewEdges(output)
}
