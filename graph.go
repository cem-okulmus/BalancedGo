package main

import (
	"bytes"
	"github.com/spakin/disjoint"
	"reflect"
)

type Graph struct {
	edges []Edge
	m     map[int]string
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
		output = append(output, otherE.nodes...)
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

func (g Graph) getComponents(sep []Edge, Sp []Special) ([]Graph, [][]Special, map[int]*disjoint.Element) {
	vertices := append(Vertices(g.edges), VerticesSpecial(Sp)...)

	return g.getCompGeneral(vertices, sep, Sp)
}

// Uses Disjoint Set data structure to compute connected components
func (g Graph) getCompGeneral(vs []int, sep []Edge, Sp []Special) ([]Graph, [][]Special, map[int]*disjoint.Element) {
	var outputG []Graph
	var outputS [][]Special

	var nodes = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Special)

	balsepVert := Vertices(sep)

	//  Set up the disjoint sets for each node
	for _, i := range vs {
		nodes[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for _, e := range g.edges {
		actualNodes := diff(e.nodes, balsepVert)
		for i := 0; i < len(actualNodes)-1; i++ {
			disjoint.Union(nodes[actualNodes[i]], nodes[actualNodes[i+1]])
		}
	}

	for _, s := range Sp {
		actualNodes := diff(s.vertices, balsepVert)
		for i := 0; i < len(actualNodes)-1; i++ {
			disjoint.Union(nodes[actualNodes[i]], nodes[actualNodes[i+1]])
		}
	}

	//sort each edge and special edge to a corresponding component
	for _, e := range g.edges {
		actualNodes := diff(e.nodes, balsepVert)
		if len(actualNodes) > 0 {
			comps[nodes[actualNodes[0]].Find()] = append(comps[nodes[actualNodes[0]].Find()], e)
		}
	}
	var isolatedSp []Special
	for _, s := range Sp {
		actualNodes := diff(s.vertices, balsepVert)
		if len(actualNodes) > 0 {
			compsSp[nodes[actualNodes[0]].Find()] = append(compsSp[nodes[actualNodes[0]].Find()], s)
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

	return outputG, outputS, nodes
}

func filterVertices(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if subset(e.nodes, vertices) {
			output = append(output, e)
		}
	}

	return output

}

func cutEdges(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		inter := inter(e.nodes, vertices)
		if len(inter) > 0 {
			output = append(output, Edge{nodes: inter, m: e.m})
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
		if len(comps[i].edges)+len(compSps[i]) > (((len(g.edges) + len(sp)) * (BALANCED_FACTOR - 1)) / BALANCED_FACTOR) {
			// log.Printf("Component %+v has weight%d instead of %d\n", comps[i], len(comps[i].edges)+len(compSps[i]), ((len(g.edges) + len(sp)) / 2))
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
			// log.Println("Special edge %+v\n used again", s)
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
			output.edges = append(output.edges, Edge{nodes: inter(e.nodes, tuple), m: e.m}.subedges()...)
			gen.confirm()
		}
	}

	output.edges = removeDuplicateEdges(output.edges)
	return output
}
