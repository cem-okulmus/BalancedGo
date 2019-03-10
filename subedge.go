package main

type Subset struct {
	source  []int
	current Combin
}

func getSubsetIterator(edges Edges) *Subset {
	var output Subset
	vertices := removeDuplicates(Vertices(edges))
	output = Subset{source: vertices, current: getCombin(len(vertices), len(vertices))}
	return &output
}

func (s Subset) hasNext() bool {
	return s.current.hasNext()
}

func getEdge(nodes []int, s []int) Edge {
	var output Edge
	for _, i := range s {
		output.nodes = append(output.nodes, nodes[i])
	}
	output.nodes = removeDuplicates(output.nodes)
	return output
}

func (s Subset) getCurrent() Edge {
	s.current.confirm()

	return getEdge(s.source, s.current.combination)
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
		if len(inter) > 0 {
			h_edges.append(Edge{nodes: inter})
		}
	}

	var output SubEdges

	output.source = h_edges
	if k > len(output.source) {
		k = len(output.source)
	}
	output.gen = NewCombinationGenerator(len(output.source), k)
	output.current = e
	output.initial = e
	output.k = k

	return output
}

func (s *SubEdges) reset() {
	(*s).gen = NewCombinationGenerator(len(s.source), s.k)
	(*s).currentSubset = nil
	(*s).current = s.initial
}

// This checks whether the current edge has a more tuples to intersect with,
// and create a new vertex set
func (s SubEdges) hasNextCombination() bool {
	if !s.gen.Next() {
		return false
	}
	s.gen.Combination(s.combination)
	return true
}

func (s SubEdges) hasNext() bool {
	if s.currentSubset == nil || !s.currentSubset.hasNext() {
		if s.hasNextCombination() {
			s.currentSubset = getSubsetIterator(getSubset(s.source, s.combination))
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

func (sep SepSub) hasNext() bool {

	i := 0
	for i < len(sep.edges) {
		if sep.edges[i].hasNext() {
			return true
		} else {
			sep.edges[i].reset()
			i++
		}
	}

	return false
}

func (sep SepSub) current() []Edge {
	var output Edges

	for _, s := range sep.edges {
		output.append(s.getCurrent())
	}

	return output
}
