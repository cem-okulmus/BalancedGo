package main

type Subset struct {
	source  []int
	current Combin
}

func getSubsetIterator(edges Edges) Subset {
	var output Subset

	output.source = Vertices(edges)
	output.current = getCombin(len(output.source), len(output.source))
	return output
}

func (s Subset) hasNext() bool {
	return s.current.hasNext()
}

func getEdge(nodes []int, s []int) Edge {
	var output Edge
	for _, i := range s {
		output.nodes = append(output.nodes, nodes[i])
	}
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
	k              int
	source         Edges
	current        *CombinationGenerator
	combination    []int
	currrentSubset *Subset
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
	output.current = NewCombinationGenerator(len(output.source), k)
	output.k = k

	return output
}

func (s *SubEdges) reset() {
	(*s).current = NewCombinationGenerator(len(s.source), s.k)
	s.hasNext()
}

// This checks whether the current edge has a more tuples to intersect with,
// and create a new vertex set
func (s SubEdges) hasNextCombination() bool {
	if !s.current.Next() {
		return false
	}
	s.current.Combination(s.combination)
	return true
}

func (s SubEdges) hasNext() bool {
	if s.currrentSubset == nil || !s.currrentSubset.hasNext() {
		if s.hasNextCombination() {
			*s.currrentSubset = getSubsetIterator(getSubset(s.source, s.combination))
			return s.currrentSubset.hasNext()
		} else {
			return false
		}
	}

	return true
}

func (s SubEdges) getCurrent() Edge {
	return s.currrentSubset.getCurrent()
}

//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------
//   ----------------------------------------------------------------------------

type SepSub struct {
	edges []SubEdges
}

func getSepSub(h Graph, Sp []Special, edges Edges, sep Edges, k int) SepSub {
	var output SepSub

	V := append(VerticesSpecial(Sp), h.Vertices()...)

	// Remove those edges that don't intersect with the current subgraph
	// And reduce those that do to their biggest allowed subedge
	var h_edges Edges
	for _, j := range edges {
		inter := inter(j.nodes, V)
		if len(inter) > 0 {
			h_edges.append(Edge{nodes: inter})
		}
	}

	for _, e := range sep {
		output.edges = append(output.edges, getSubEdgeIterator(h_edges, e, k))
	}

	return output
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
