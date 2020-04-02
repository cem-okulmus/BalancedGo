// functions to compute a hinge-tree decomposition of a hypergraph, and methods to use it to speed up the computation of a GHD

package lib

import (
	"bytes"
	"fmt"
)

type hingeEdge struct {
	h hingetree
	e Edge
}

type hingetree struct {
	hinge    Graph
	minimal  bool
	children []hingeEdge
}

func (h hingetree) stringIdent(i int) string {
	var buffer bytes.Buffer

	buffer.WriteString("\n" + indent(i) + h.hinge.String() + "\n")

	if len(h.children) > 0 {
		buffer.WriteString(indent(i) + "Children:\n" + indent(i) + "[")
		for _, c := range h.children {
			buffer.WriteString("sepEdge: " + c.e.String() + "\n")
			buffer.WriteString(indent(i) + c.h.stringIdent(i+1))
		}
		buffer.WriteString(indent(i) + "]\n")
	}

	return buffer.String()
}

func (h hingetree) String() string {
	return h.stringIdent(0)
}

func GetHingeTree(g Graph) hingetree {

	initialTree := hingetree{hinge: g}

	isUsed := make(map[int]bool)

	for _, e := range g.Edges.Slice() {
		isUsed[e.Name] = false
	}

	output := initialTree.expandHingeTree(isUsed)

	fmt.Println("Proudced Hingetree \n")
	fmt.Println(output.String())

	return output
}

func (h hingetree) expandHingeTree(isUsed map[int]bool) hingetree {

	// keep expanding the current node until no more new children can be generated
	for !h.minimal {
		var e *Edge

		{ // maybe a separate function for this?
			//select next unused edge
			for _, i := range h.hinge.Edges.Slice() {
				if isUsed[i.Name] {
					continue
				} else {
					e = &i
					isUsed[e.Name] = true // set selected edge to used
				}
			}
		}

		if e == nil {
			h.minimal = true
			continue
		}
		sepEdge := NewEdges([]Edge{*e})
		hinges, _, gamma := h.hinge.GetComponents(sepEdge, []Special{})

		// Skip reordering step if only single component
		if len(hinges) == 1 {
			continue
		}

		// hinges into hingetrees, preserving order
		var htrees []hingetree

		for i := range hinges {
			edges := hinges[i].Edges.Slice()
			extendedHinge := Graph{Edges: NewEdges(append(edges, *e))}
			htrees = append(htrees, hingetree{hinge: extendedHinge})
		}

		// then assign the children to each hingetree
		for _, hingedge := range h.children {
			i := gamma[hingedge.e.Vertices[0]]
			htrees[i].children = append(htrees[i].children, hingedge)
		}

		// finally add hingetrees above to newchildren
		h = htrees[0]

		for i := range htrees[1:] {
			h.children = append(h.children, hingeEdge{e: *e, h: htrees[i]})
		}

	}

	// recursively repeat procedure over all children of h

	for i := range h.children {
		h.children[i].h = h.children[i].h.expandHingeTree(isUsed)
	}

	return h
}
