package lib

import (
	"bytes"
	"reflect"

	"github.com/spakin/disjoint"
)

// A Graph is a collection of edges
type Graph struct {
	Edges Edges
}

func (g Graph) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, e := range g.Edges {
		buffer.WriteString(e.String())
		if i != len(g.Edges)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

// produces the union of all vertices from all edges of the graph
func (g Graph) Vertices() []int {
	var output []int
	for _, otherE := range g.Edges {
		output = append(output, otherE.Vertices...)
	}
	return RemoveDuplicates(output)
}

func GetSubset(edges []Edge, s []int) []Edge {
	var output []Edge
	for _, i := range s {
		output = append(output, edges[i])
	}
	return output
}

func (g Graph) GetSubset(s []int) []Edge {
	return GetSubset(g.Edges, s)
}

// Uses Disjoint Set data structure to compute connected components
func (g Graph) GetComponents(sep []Edge, Sp []Special) ([]Graph, [][]Special, map[int]*disjoint.Element) {
	var outputG []Graph
	var outputS [][]Special

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Special)

	balsepVert := Vertices(sep)

	//  Set up the disjoint sets for each node
	for _, i := range append(Vertices(g.Edges), VerticesSpecial(Sp)...) {
		vertices[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for _, e := range g.Edges {
		actualVertices := Diff(e.Vertices, balsepVert)
		for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
			disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		}
	}

	for _, s := range Sp {
		actualVertices := Diff(s.Vertices, balsepVert)
		for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
			disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		}
	}

	//sort each edge and special edge to a corresponding component
	for _, e := range g.Edges {
		actualVertices := Diff(e.Vertices, balsepVert)
		if len(actualVertices) > 0 {
			comps[vertices[actualVertices[0]].Find()] = append(comps[vertices[actualVertices[0]].Find()], e)
		}
	}
	var isolatedSp []Special
	for _, s := range Sp {
		actualVertices := Diff(s.Vertices, balsepVert)
		if len(actualVertices) > 0 {
			compsSp[vertices[actualVertices[0]].Find()] = append(compsSp[vertices[actualVertices[0]].Find()], s)
		} else {
			isolatedSp = append(isolatedSp, s)
		}
	}

	// Store the components as graphs
	for k, _ := range comps {
		g := Graph{Edges: comps[k]}
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

func FilterVertices(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if len(Inter(e.Vertices, vertices)) > 0 {
			output = append(output, e)
		}
	}

	return output

}

func FilterVerticesStrict(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		if Subset(e.Vertices, vertices) {
			output = append(output, e)
		}
	}

	return output

}

func CutEdges(edges []Edge, vertices []int) []Edge {
	var output []Edge

	for _, e := range edges {
		inter := Inter(e.Vertices, vertices)
		if len(inter) > 0 {
			output = append(output, Edge{Vertices: inter})
		}
	}

	return output

}

func (g Graph) CheckBalancedSep(sep []Edge, sp []Special, balancedFactor int) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, compSps, _ := g.GetComponents(sep, sp)
	// log.Printf("Components of sep %+v\n", comps)
	for i := range comps {
		if len(comps[i].Edges)+len(compSps[i]) > (((len(g.Edges) + len(sp)) * (balancedFactor - 1)) / balancedFactor) {
			//	log.Printf("Using %+v component %+v has weight %d instead of %d\n", sep, comps[i], len(comps[i].Edges)+len(compSps[i]), ((len(g.Edges) + len(sp)) / 2))
			return false
		}
	}

	// Check if subset of V(H) + Vertices of Sp
	// var allowedVertices = append(g.Vertices(), VerticesSpecial(sp)...)
	// if !Subset(Vertices(sep), allowedVertices) {
	// 	// log.Println("Subset condition violated")
	// 	return false
	// }

	// Make sure that "special seps can never be used as separators"
	for _, s := range sp {
		if reflect.DeepEqual(s.Vertices, Vertices(sep)) {
			//	log.Println("Special edge %+v\n used again", s)
			return false
		}
	}

	return true
}

func (g Graph) CheckNextSep(sep []Edge, oldSep []Edge, Sp []Special) bool {

	verticesCurrent := append(g.Vertices(), VerticesSpecial(Sp)...)

	// check if balsep covers the intersection of oldsep and H
	if !Subset(Inter(Vertices(oldSep), verticesCurrent), Vertices(sep)) {
		return false
	}
	//check if balsep "makes some progress" into separating H
	if len(Inter(Vertices(sep), Diff(verticesCurrent, Vertices(oldSep)))) == 0 {
		return false
	}

	return true

}

func (g Graph) ComputeSubEdges(K int) Graph {
	var output = g

	for _, e := range g.Edges {
		edgesWihoutE := DiffEdges(g.Edges, e)
		gen := GetCombin(len(edgesWihoutE), K)
		for gen.HasNext() {
			var tuple = Vertices(GetSubset(edgesWihoutE, gen.Combination))
			output.Edges = append(output.Edges, Edge{Vertices: Inter(e.Vertices, tuple)}.subedges()...)
			gen.Confirm()
		}
	}

	output.Edges = removeDuplicateEdges(output.Edges)
	return output
}
