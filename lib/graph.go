package lib

import (
	"bytes"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spakin/disjoint"
)

// A Graph is a collection of edges
type Graph struct {
	Edges    Edges
	vertices []int
}

func (g Graph) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, e := range g.Edges.Slice() {
		buffer.WriteString(e.String())
		if i != g.Edges.Len()-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (this Graph) equal(other Graph) bool {
	return cmp.Equal(this, other, cmpopts.IgnoreUnexported(this), cmp.Comparer(equalEdges))
}

// produces the union of all vertices from all edges of the graph
func (g *Graph) Vertices() []int {
	if len(g.vertices) > 0 {
		return g.vertices
	}
	var output []int
	for _, otherE := range g.Edges.Slice() {
		output = append(output, otherE.Vertices...)
	}
	g.vertices = RemoveDuplicates(output)
	return g.vertices
}

func GetSubset(edges Edges, s []int) Edges {
	var output []Edge
	for _, i := range s {
		output = append(output, edges.Slice()[i])
	}
	return NewEdges(output)
}

func (g Graph) GetSubset(s []int) Edges {
	return GetSubset(g.Edges, s)
}

// Uses Disjoint Set data structure to compute connected components
func (g Graph) GetComponents(sep Edges, Sp []Special) ([]Graph, [][]Special, map[int]*disjoint.Element) {
	var outputG []Graph
	var outputS [][]Special

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Special)

	balsepVert := sep.Vertices()
	balSepCache := make(map[int]bool, len(balsepVert))
	//fmt.Println("Current separator ", Edge{Vertices: balsepVert})
	for _, v := range balsepVert {
		balSepCache[v] = true
	}

	//  Set up the disjoint sets for each node
	for _, i := range append(g.Edges.Vertices(), VerticesSpecial(Sp)...) {
		vertices[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for _, e := range g.Edges.Slice() {
		//	fmt.Println("Have edge ", Edge{Vertices: e.Vertices})
		// actualVertices := Diff(e.Vertices, balsepVert)
		// for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
		// 	disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		// }
		for i := 0; i < len(e.Vertices); i++ {
			if balSepCache[e.Vertices[i]] {

				continue
			}
			for j := i + 1; j < len(e.Vertices); j++ {
				if balSepCache[e.Vertices[j]] {
					continue
				}
				//			fmt.Println("Union of ", m[e.Vertices[i]], "and ", m[e.Vertices[j]])
				disjoint.Union(vertices[e.Vertices[i]], vertices[e.Vertices[j]])
				// j = i-1
				break
			}
		}
	}

	for _, s := range Sp {
		//actualVertices := Diff(s.Vertices, balsepVert)
		// for i := 0; i < len(actualVertices)-1 && i+1 < len(actualVertices); i++ {
		// 	disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		// }

		for i := 0; i < len(s.Vertices)-1; i++ {
			if balSepCache[s.Vertices[i]] {
				continue
			}
			for j := i + 1; j < len(s.Vertices); j++ {
				if balSepCache[s.Vertices[j]] {
					continue
				}
				disjoint.Union(vertices[s.Vertices[i]], vertices[s.Vertices[j]])
				// j = i-1
				break
			}
		}

	}

	//sort each edge and special edge to a corresponding component
	for _, e := range g.Edges.Slice() {
		//actualVertices := Diff(e.Vertices, balsepVert)
		// if len(actualVertices) > 0 {
		// 	edges := comps[vertices[actualVertices[0]].Find()]
		// 	edges.append(e) // = append(comps[vertices[actualVertices[0]].Find()], e)
		// }
		var vertexRep int
		found := false
		for _, v := range e.Vertices {
			if balSepCache[v] {
				continue
			}
			vertexRep = v
			found = true
			break
		}
		if !found {
			continue
		}

		slice, ok := comps[vertices[vertexRep].Find()]
		if !ok {
			newslice := make([]Edge, 0, g.Edges.Len())
			comps[vertices[vertexRep].Find()] = newslice
			slice = newslice
		}

		comps[vertices[vertexRep].Find()] = append(slice, e)

	}

	var isolatedSp []Special
	for _, s := range Sp {
		// actualVertices := Diff(s.Vertices, balsepVert)
		// if len(actualVertices) > 0 {
		// 	compsSp[vertices[actualVertices[0]].Find()] = append(compsSp[vertices[actualVertices[0]].Find()], s)
		// } else {
		// 	isolatedSp = append(isolatedSp, s)
		// }
		var vertexRep int
		found := false
		for _, v := range s.Vertices {
			if balSepCache[v] {
				continue
			}
			vertexRep = v
			found = true
			break
		}
		if !found {
			isolatedSp = append(isolatedSp, s)
			continue
		}

		//compsSp[vertices[vertexRep].Find()] = append(compsSp[vertices[vertexRep].Find()], s)

		slice, ok := compsSp[vertices[vertexRep].Find()]
		if !ok {
			newslice := make([]Special, 0, len(Sp))
			compsSp[vertices[vertexRep].Find()] = newslice
			slice = newslice
		}

		compsSp[vertices[vertexRep].Find()] = append(slice, s)

	}

	// Store the components as graphs
	for k, _ := range comps {
		g := Graph{Edges: Edges{slice: comps[k]}}
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

// Uses Disjoint Set data structure to compute connected components
func (g Graph) GetComponentsIsolated(sep Edges, sepBoth Edges, up []int, low []int) ([]Graph, int, int, []Edge) {
	var outputG []Graph
	var outputEdgesinS []Edge

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	upCompIndex := -1
	lowCompIndex := -1
	var upComp *disjoint.Element
	var lowComp *disjoint.Element

	balsepVert := Diff(sep.Vertices(), append(up, low...))
	balSepCache := make(map[int]bool, len(balsepVert))
	// fmt.Println("Current separator ", PrintVertices(balsepVert))
	for _, v := range balsepVert {
		balSepCache[v] = true
	}

	//  Set up the disjoint sets for each node
	for _, i := range g.Edges.Vertices() {
		vertices[i] = disjoint.NewElement()
	}

	//up
	for i := 0; i < len(up)-1; i++ {
		for j := i + 1; j < len(up); j++ {
			disjoint.Union(vertices[up[i]], vertices[up[j]])
			break
		}
	}
	//low
	for i := 0; i < len(low)-1; i++ {
		for j := i + 1; j < len(low); j++ {
			disjoint.Union(vertices[low[i]], vertices[low[j]])
			break
		}
	}

	// Merge together the connected components
	for e := range g.Edges.Slice() {
		for i := 0; i < len(g.Edges.Slice()[e].Vertices); i++ {
			if balSepCache[g.Edges.Slice()[e].Vertices[i]] {
				continue
			}
			for j := i + 1; j < len(g.Edges.Slice()[e].Vertices); j++ {
				if balSepCache[g.Edges.Slice()[e].Vertices[j]] {
					continue
				}
				disjoint.Union(vertices[g.Edges.Slice()[e].Vertices[i]], vertices[g.Edges.Slice()[e].Vertices[j]])
				break
			}
		}
	}

	//sort each edge to a corresponding component
	for i := range g.Edges.Slice() {
		var vertexRep int
		found := false
		for _, v := range g.Edges.Slice()[i].Vertices {
			if balSepCache[v] {
				continue
			}
			vertexRep = v
			found = true
			break
		}
		if !found {
			outputEdgesinS = append(outputEdgesinS, g.Edges.Slice()[i])
			continue
		}

		slice, ok := comps[vertices[vertexRep].Find()]
		if !ok {
			newslice := make([]Edge, 0, g.Edges.Len())
			comps[vertices[vertexRep].Find()] = newslice
			slice = newslice
		}

		comps[vertices[vertexRep].Find()] = append(slice, g.Edges.Slice()[i])

	}

	if len(up) > 0 {
		upComp = vertices[up[0]].Find()
	}
	if len(low) > 0 {
		lowComp = vertices[low[0]].Find()
	}

	currentIndex := 0
	// Store the components as graphs
	for k, _ := range comps {
		if k == upComp {
			upCompIndex = currentIndex
		}
		if k == lowComp {
			lowCompIndex = currentIndex
		}

		g := Graph{Edges: Edges{slice: comps[k]}}
		outputG = append(outputG, g)
		currentIndex++
	}

	return outputG, upCompIndex, lowCompIndex, outputEdgesinS
}

func FilterVertices(edges Edges, vertices []int) Edges {
	var output []Edge

	for _, e := range edges.Slice() {
		if len(Inter(e.Vertices, vertices)) > 0 {
			output = append(output, e)
		}
	}

	return NewEdges(output)

}

func FilterVerticesStrict(edges Edges, vertices []int) Edges {
	var output []Edge

	for _, e := range edges.Slice() {
		if Subset(e.Vertices, vertices) {
			output = append(output, e)
		}
	}

	return NewEdges(output)

}

func CutEdges(edges Edges, vertices []int) Edges {
	var output []Edge

	for i := range edges.Slice() {
		inter := Inter(edges.Slice()[i].Vertices, vertices)
		if len(inter) > 0 {
			name := edges.Slice()[i].Name
			// if len(inter) < len(edges.Slice()[i].Vertices) {

			// 	var mux sync.Mutex
			// 	mux.Lock() // ensure that hash is computed only on one gorutine at a time
			// 	name = encode
			// 	m[encode] = m[edges.Slice()[i].Name] + "'"
			// 	encode++
			// 	mux.Unlock()

			// }
			output = append(output, Edge{Name: name, Vertices: inter})
		}
	}

	return NewEdges(output)

}

func (g Graph) CheckBalancedSep(sep Edges, sp []Special, balancedFactor int) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, compSps, _ := g.GetComponents(sep, sp)
	// log.Printf("Components of sep %+v\n", comps)
	for i := range comps {
		if comps[i].Edges.Len()+len(compSps[i]) > (((g.Edges.Len() + len(sp)) * (balancedFactor - 1)) / balancedFactor) {
			//log.Printf("Using %+v component %+v has weight %d instead of %d\n", sep, comps[i], comps[i].Edges.Len()+len(compSps[i]), ((g.Edges.Len() + len(sp)) / 2))
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
		if reflect.DeepEqual(s.Vertices, sep.Vertices()) {
			//log.Println("Special edge %+v\n used again", s)
			return false
		}
	}

	return true
}

func (g Graph) CheckNextSep(sep Edges, oldSep Edges, Sp []Special) bool {

	verticesCurrent := append(g.Vertices(), VerticesSpecial(Sp)...)

	// check if balsep covers the intersection of oldsep and H
	if !Subset(Inter(oldSep.Vertices(), verticesCurrent), sep.Vertices()) {
		return false
	}
	//check if balsep "makes some progress" into separating H
	if len(Inter(sep.Vertices(), Diff(verticesCurrent, oldSep.Vertices()))) == 0 {
		return false
	}

	return true

}

func (g Graph) ComputeSubEdges(K int) Graph {
	var output = g.Edges.Slice()

	for _, e := range g.Edges.Slice() {
		edgesWihoutE := DiffEdges(g.Edges, e)
		gen := GetCombin(edgesWihoutE.Len(), K)
		for gen.HasNext() {
			subset := GetSubset(edgesWihoutE, gen.Combination)
			var tuple = subset.Vertices()
			output = append(output, Edge{Vertices: Inter(e.Vertices, tuple)}.subedges()...)
			gen.Confirm()
		}
	}

	return Graph{Edges: removeDuplicateEdges(output)}
}

func (g Graph) GetBIP() int {
	var output int

	edges := g.Edges.Slice()

	for i := range edges {
		for j := range edges {
			if j <= i {
				continue
			}
			tmp := len(Inter(edges[i].Vertices, edges[j].Vertices))
			if tmp > output {
				output = tmp
			}
		}
	}

	return output
}
