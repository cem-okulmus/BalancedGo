// Computes a hinge tree of a given hypergraph

package main

// import "github.com/spakin/disjoint"

// type LabelledEdge struct {
// 	used bool
// 	edge Edge
// }

// func VerticesLabelled(e []*LabelledEdge) []int {
// 	var output []int
// 	for _, otherE := range e {
// 		output = append(output, otherE.edge.vertices...)
// 	}
// 	return removeDuplicates(output)
// }

// func toEdges(ls []*LabelledEdge) []Edge {
// 	var output Edges

// 	for _, e := range ls {
// 		output.append(e.edge)
// 	}

// 	return output
// }

// func getHingeComponents(edges []*LabelledEdge, sep *LabelledEdge) [][]*LabelledEdge {
// 	var output [][]*LabelledEdge

// 	var vertices = make(map[int]*disjoint.Element)
// 	var comps = make(map[*disjoint.Element][]*LabelledEdge)

// 	balsepVert := sep.edge.vertices

// 	//  Set up the disjoint sets for each node
// 	for _, i := range VerticesLabelled(edges) {
// 		vertices[i] = disjoint.NewElement()
// 	}

// 	// Merge together the connected components
// 	for _, e := range edges {
// 		actualVertices := diff(e.edge.vertices, balsepVert)
// 		for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
// 			disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
// 		}
// 	}

// 	//sort each edge to its corresponding component
// 	for _, e := range edges {
// 		actualVertices := diff(e.edge.vertices, balsepVert)
// 		if len(actualVertices) > 0 {
// 			comps[vertices[actualVertices[0]].Find()] = append(comps[vertices[actualVertices[0]].Find()], e)
// 		}
// 	}

// 	// Store the components as graphs
// 	for k := range comps {
// 		output = append(output, comps[k])
// 	}

// 	return output
// }

// type HingetreeArc struct {
// 	edge  *LabelledEdge
// 	child HingetreeConstruct
// }

// type HingetreeConstruct struct {
// 	minimal  bool
// 	node     []*LabelledEdge
// 	children []HingetreeArc
// }

// func (h *HingetreeConstruct) nextNonminimal() (HingetreeConstruct, bool) {
// 	if !h.minimal {
// 		return h, true
// 	}

// 	for i, c := range h.children {
// 		node, found, path := c.child.nextNonminimal()
// 		if found {
// 			return node, found
// 		}
// 	}

// 	return HingetreeConstruct{}, false
// }

// func (h *HingetreeConstruct) toHinge() Hingetree {
// 	var output Hingetree

// 	// copy all edges
// 	for _, e := range h.node {
// 		output.node.append(e.edge)
// 	}

// 	for _, c := range h.children {
// 		output.children = append(output.children, c.child.toHinge())
// 	}

// 	return output
// }

// type Hingetree struct {
// 	node     Edges
// 	children []Hingetree
// }

// func nextUnused(edges []*LabelledEdge) (*LabelledEdge, bool) {

// 	for _, e := range edges {
// 		if !e.used {
// 			return e, true
// 		}
// 	}

// 	return LabelledEdge{}, false
// }

// func pickComponent(edge *LabelledEdge, comps [][]*LabelledEdge) ([]*LabelledEdge, bool) {

// 	for _, comp := range comps {
// 		for _, l := range comps {
// 			if edge == l {
// 				return comp, true
// 			}
// 		}
// 	}

// 	return []*LabelledEdge{}, false

// }

// // based on "Algorithm 3.9" from Gyssens et al. 1994
// func (g Graph) getHinge() Hingetree {
// 	var output *HingetreeConstruct

// 	//initialise the Hingetree
// 	for _, e := range g.edges {
// 		output.node = append(output.node, LabelledEdge{edge: e})
// 	}

// 	// repeat until non mininal nodes exist in Hingetree
// OUTER:
// 	for {
// 		chosen, found, path := output.nextNonminimal()

// 		if !found {
// 			break // stop computation
// 		}

// 		var components [][]LabelledEdge
// 	INNER:
// 		for {
// 			// Find next unused edge, mark node as minimal if none exists
// 			edge, existUnused := nextUnused(chosen.node)

// 			if !(existUnused) {
// 				chosen.minimal = true
// 				continue OUTER // look for next non minimal node
// 			}
// 			edge.used = true

// 			components = getHingeComponents(chosen.node, edge)

// 			if len(components) == 1 {
// 				continue INNER
// 			}

// 		}

// 		// for _,c := range

// 		// chosen.node = components[0] // set initial comp to old node

// 		// for i, c := range components {
// 		// 	if i == 0 {
// 		// 		continue
// 		// 	}
// 		// 	chosen.node.children = append(chosen.node.children, HingetreeConstruct{node: c})

// 		// }

// 	}

// 	return output.toHinge()
// }
