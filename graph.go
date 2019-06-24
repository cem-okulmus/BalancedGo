package main

import (
	"bytes"
	"reflect"

	"github.com/spakin/disjoint"
)

// A Graph is a collection of edges
type Graph struct {
	edges []Edge
}

func (g Graph) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, e := range g.edges {
		buffer.WriteString(e.String())
		if i != len(g.edges)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (g Graph) Vertices() []int {
	var output []int
	for _, otherE := range g.edges {
		output = append(output, otherE.vertices...)
	}
	return removeDuplicates(output)
}

func getSubset(edges []Edge, s []int) []Edge {
	var output []Edge
	for _, i := range s {
		output = append(output, edges[i])
	}
	return output
}

func (g Graph) getSubset(s []int) []Edge {
	return getSubset(g.edges, s)
}

// Uses Disjoint Set data structure to compute connected components
func (g Graph) getComponents(sep []Edge, Sp []Special) ([]Graph, [][]Special, map[int]*disjoint.Element) {
	var outputG []Graph
	var outputS [][]Special

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Special)

	balsepVert := Vertices(sep)

	//  Set up the disjoint sets for each node
	for _, i := range append(Vertices(g.edges), VerticesSpecial(Sp)...) {
		vertices[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for _, e := range g.edges {
		actualVertices := diff(e.vertices, balsepVert)
		for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
			disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		}
	}

	for _, s := range Sp {
		actualVertices := diff(s.vertices, balsepVert)
		for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
			disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		}
	}

	//sort each edge and special edge to a corresponding component
	for _, e := range g.edges {
		actualVertices := diff(e.vertices, balsepVert)
		if len(actualVertices) > 0 {
			comps[vertices[actualVertices[0]].Find()] = append(comps[vertices[actualVertices[0]].Find()], e)
		}
	}
	var isolatedSp []Special
	for _, s := range Sp {
		actualVertices := diff(s.vertices, balsepVert)
		if len(actualVertices) > 0 {
			compsSp[vertices[actualVertices[0]].Find()] = append(compsSp[vertices[actualVertices[0]].Find()], s)
		} else {
			isolatedSp = append(isolatedSp, s)
		}
	}

	// Store the components as graphs
	for k, _ := range comps {
		g := Graph{edges: comps[k]}
		outputG = append(outputG, g)
		outputS = append(outputS, compsSp[k])

	}

	for k, _ := range compsSp {
		_, ok := comps[k]
		if ok {
			continue
		}
		g := Graph{}
		outputG = append(outputG, g)
		outputS = append(outputS, compsSp[k])
	}

	for _, s := range isolatedSp {
		g := Graph{}
		outputG = append(outputG, g)
		outputS = append(outputS, []Special{s})
	}

	return outputG, outputS, vertices
}

func filterVertices(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if len(inter(e.vertices, vertices)) > 0 {
			output = append(output, e)
		}
	}

	return output

}

func filterVerticesStrict(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if subset(e.vertices, vertices) {
			output = append(output, e)
		}
	}

	return output

}

func cutEdges(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		inter := inter(e.vertices, vertices)
		if len(inter) > 0 {
			output = append(output, Edge{vertices: inter})
		}
	}

	return output

}

func (g Graph) checkBalancedSep(sep []Edge, sp []Special) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, compSps, _ := g.getComponents(sep, sp)
	// log.Printf("Components of sep %+v\n", comps)
	for i := range comps {
		if len(comps[i].edges)+len(compSps[i]) > (((len(g.edges) + len(sp)) * (BalancedFactor - 1)) / BalancedFactor) {
			//	log.Printf("Using %+v component %+v has weight %d instead of %d\n", sep, comps[i], len(comps[i].edges)+len(compSps[i]), ((len(g.edges) + len(sp)) / 2))
			return false
		}
	}

	// Check if subset of V(H) + Vertices of Sp
	// var allowedVertices = append(g.Vertices(), VerticesSpecial(sp)...)
	// if !subset(Vertices(sep), allowedVertices) {
	// 	// log.Println("Subset condition violated")
	// 	return false
	// }

	// Make sure that "special seps can never be used as separators"
	for _, s := range sp {
		if reflect.DeepEqual(s.vertices, Vertices(sep)) {
			//	log.Println("Special edge %+v\n used again", s)
			return false
		}
	}

	return true
}

func (g Graph) computeSubEdges(K int) Graph {
	var output = g

	for _, e := range g.edges {
		edges_wihout_e := diffEdges(g.edges, e)
		gen := getCombin(len(edges_wihout_e), K)
		for gen.hasNext() {
			var tuple = Vertices(getSubset(edges_wihout_e, gen.combination))
			output.edges = append(output.edges, Edge{vertices: inter(e.vertices, tuple)}.subedges()...)
			gen.confirm()
		}
	}

	output.edges = removeDuplicateEdges(output.edges)
	return output
}
