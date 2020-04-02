// functions to compute a hinge-tree decomposition of a hypergraph, and methods to use it to speed up the computation of a GHD

package lib

import (
	"bytes"
	"log"
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
			buffer.WriteString("\n" + indent(i+1) + "sepEdge: " + c.e.String())
			buffer.WriteString(c.h.stringIdent(i + 1))
		}
		buffer.WriteString(indent(i) + "]\n")
	}

	return buffer.String()
}

func (h hingetree) String() string {
	return h.stringIdent(0)
}

func getHingeTree(g Graph) hingetree {

	initialTree := hingetree{hinge: g}

	isUsed := make(map[int]bool)

	for _, e := range g.Edges.Slice() {
		isUsed[e.Name] = false
	}

	// // fmt.Println("Initial Hingetree \n")
	// // fmt.Println(initialTree.String())

	output := initialTree.expandHingeTree(isUsed, -1)

	// fmt.Println("Proudced Hingetree \n")
	// fmt.Println(output.String())

	return output
}

// Following Gyssens, 1994, implemented re
func (h hingetree) expandHingeTree(isUsed map[int]bool, parentE int) hingetree {

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
					isUsed[i.Name] = true // set selected edge to used
					break
				}
			}
		}

		if e == nil {
			h.minimal = true

			// fmt.Println("Setting hinge ", h.hinge, " to minimal")
			continue
		}

		if parentE != -1 {
			// fmt.Println("Next unused edge", *e, "with parent Edge: ", m[parentE])

		} else {

			// fmt.Println("Next unused edge", *e, " at root")
		}

		sepEdge := NewEdges([]Edge{*e})
		hinges, _, gamma := h.hinge.GetComponents(sepEdge, []Special{})

		// fmt.Printf("Hinges of Sep: %v\n", hinges)

		// Skip reordering step if only single component
		if len(hinges) == 1 {
			// fmt.Println("skipping sepEdge since no progress")
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
			i := gamma[hingedge.e.Name]
			htrees[i].children = append(htrees[i].children, hingedge)
		}

		// finally add hingetrees above to newchildren
		if parentE == -1 {
			h = htrees[0]
		} else {
			h = htrees[gamma[parentE]]

			found := false

			for _, e := range h.hinge.Edges.Slice() {
				if e.Name == parentE {
					found = true
				}
			}

			if !found {
				log.Panic(m[parentE], "does not occur in", htrees[gamma[parentE]].hinge)
			}

		}

		for i := range htrees {
			if (parentE == -1 && i == 0) || (parentE != -1 && i == gamma[parentE]) {
				continue
			}
			h.children = append(h.children, hingeEdge{e: *e, h: htrees[i]})
		}

		// fmt.Println("Current Hingetree \n")
		// fmt.Println(h.String())

	}

	// fmt.Println("going over children")

	// recursively repeat procedure over all children of h

	for i := range h.children {
		h.children[i].h = h.children[i].h.expandHingeTree(isUsed, h.children[i].e.Name)
	}

	return h
}
