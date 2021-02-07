package lib

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spakin/disjoint"
)

// A Graph is a collection of (special) edges
type Graph struct {
	Edges    Edges
	Special  []Edges
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

	if len(g.Special) > 0 {
		buffer.WriteString(" & Special Edges [")
		for i := range g.Special {
			buffer.WriteString(g.Special[i].String())
			if i != len(g.Special)-1 {
				buffer.WriteString(", ")
			}
		}
		buffer.WriteString(" ]")
	}

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
	for i := range g.Special {
		output = append(output, g.Special[i].Vertices()...)
	}
	g.vertices = RemoveDuplicates(output)
	return g.vertices
}

func (g Graph) Len() int {
	return g.Edges.Len() + len(g.Special)
}

// func GetSubsetMap(edges Edges, s []int, Map []int) Edges {
// 	var output []Edge
// 	for _, i := range s {
// 		output = append(output, edges.Slice()[Map[i]])
// 	}
// 	return NewEdges(output)
// }

// Special edges are ignored here, since they should never be
// considered when choosing a separator
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
func (g Graph) GetComponents(sep Edges) ([]Graph, map[int]int, []Edge) {
	var outputG []Graph

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]Edge)
	var compsSp = make(map[*disjoint.Element][]Edges)

	balsepVert := sep.Vertices()
	balSepCache := make(map[int]bool, len(balsepVert))
	//fmt.Println("Current separator ", Edge{Vertices: balsepVert})
	for _, v := range balsepVert {
		balSepCache[v] = true
	}

	//  Set up the disjoint sets for each node
	for _, i := range g.Vertices() {
		vertices[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for k := range g.Edges.Slice() {
		//  fmt.Println("Have edge ", Edge{Vertices: e.Vertices})
		// actualVertices := Diff(e.Vertices, balsepVert)
		// for i := 0; i < len(actualVertices)-1 && i+1 < len(vertices); i++ {
		//  disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		// }
		for i := 0; i < len(g.Edges.Slice()[k].Vertices); i++ {
			if balSepCache[g.Edges.Slice()[k].Vertices[i]] {

				continue
			}
			for j := i + 1; j < len(g.Edges.Slice()[k].Vertices); j++ {
				if balSepCache[g.Edges.Slice()[k].Vertices[j]] {
					continue
				}
				// fmt.Println("Union of ", m[g.Edges.Slice()[k].Vertices[i]], "and ",
				//      m[g.Edges.Slice()[k].Vertices[j]])
				disjoint.Union(vertices[g.Edges.Slice()[k].Vertices[i]], vertices[g.Edges.Slice()[k].Vertices[j]])
				// j = i-1
				break
			}
		}
	}

	for k := range g.Special {
		//actualVertices := Diff(s.Vertices, balsepVert)
		// for i := 0; i < len(actualVertices)-1 && i+1 < len(actualVertices); i++ {
		//  disjoint.Union(vertices[actualVertices[i]], vertices[actualVertices[i+1]])
		// }

		for i := 0; i < len(g.Special[k].Vertices())-1; i++ {
			if balSepCache[g.Special[k].Vertices()[i]] {
				continue
			}
			for j := i + 1; j < len(g.Special[k].Vertices()); j++ {
				if balSepCache[g.Special[k].Vertices()[j]] {
					continue
				}
				disjoint.Union(vertices[g.Special[k].Vertices()[i]], vertices[g.Special[k].Vertices()[j]])
				// j = i-1
				break
			}
		}

	}

	var isolatedEdges []Edge

	//sort each edge and special edge to a corresponding component
	for i := range g.Edges.Slice() {
		//actualVertices := Diff(e.Vertices, balsepVert)
		// if len(actualVertices) > 0 {
		//  edges := comps[vertices[actualVertices[0]].Find()]
		//  edges.append(e) // = append(comps[vertices[actualVertices[0]].Find()], e)
		// }
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
			isolatedEdges = append(isolatedEdges, g.Edges.Slice()[i])
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

	var isolatedSp []Edges
	for i := range g.Special {
		// actualVertices := Diff(s.Vertices, balsepVert)
		// if len(actualVertices) > 0 {
		//  compsSp[vertices[actualVertices[0]].Find()] = append(compsSp[vertices[actualVertices[0]].Find()], s)
		// } else {
		//  isolatedSp = append(isolatedSp, s)
		// }
		var vertexRep int
		found := false
		for _, v := range g.Special[i].Vertices() {
			if balSepCache[v] {
				continue
			}
			vertexRep = v
			found = true
			break
		}
		if !found {
			isolatedSp = append(isolatedSp, g.Special[i])
			continue
		}

		//compsSp[vertices[vertexRep].Find()] = append(compsSp[vertices[vertexRep].Find()], s)

		slice, ok := compsSp[vertices[vertexRep].Find()]
		if !ok {
			newslice := make([]Edges, 0, len(g.Special))
			compsSp[vertices[vertexRep].Find()] = newslice
			slice = newslice
		}

		compsSp[vertices[vertexRep].Find()] = append(slice, g.Special[i])
	}

	edgeToComp := make(map[int]int)

	// Store the components as graphs
	for k, _ := range comps {
		slice := comps[k]
		for i := range slice {
			edgeToComp[slice[i].Name] = len(outputG)
		}
		g := Graph{Edges: NewEdges(slice), Special: compsSp[k]}
		outputG = append(outputG, g)
		// outputS = append(outputS, compsSp[k])

	}

	for k, _ := range compsSp {
		_, ok := comps[k]
		if ok {
			continue
		}
		g := Graph{Special: compsSp[k]}
		outputG = append(outputG, g)
		// outputS = append(outputS, compsSp[k])
	}

	for i := range isolatedSp {
		g := Graph{Special: []Edges{isolatedSp[i]}}
		outputG = append(outputG, g)
		// outputS = append(outputS, []Special{s})
	}

	return outputG, edgeToComp, isolatedEdges
}

// // Uses Disjoint Set data structure to compute connected components
// func (g Graph) GetComponentsIsolated(sep Edges, up []int, low []int) ([]Graph, int, int, []Edge) {
//  var outputG []Graph
//  var isolatedEdges []Edge

//  var vertices = make(map[int]*disjoint.Element)
//  var comps = make(map[*disjoint.Element][]Edge)

//  upCompIndex := -1
//  lowCompIndex := -1

//  var upComp *disjoint.Element
//  var lowComp *disjoint.Element

//  balsepVert := Diff(sep.Vertices(), append(up, low...))
//  balSepCache := make(map[int]bool, len(balsepVert))
//  // fmt.Println("Current separator ", PrintVertices(balsepVert))
//  for _, v := range balsepVert {
//      balSepCache[v] = true
//  }

//  //  Set up the disjoint sets for each node
//  for _, i := range g.Edges.Vertices() {
//      vertices[i] = disjoint.NewElement()
//  }

//  // //low
//  // for i := 0; i < len(low)-1; i++ {
//  //  for j := i + 1; j < len(low); j++ {
//  //      disjoint.Union(vertices[low[i]], vertices[low[j]])
//  //      break
//  //  }
//  // }

//  // Merge together the connected components
//  for e := range g.Edges.Slice() {
//      for i := 0; i < len(g.Edges.Slice()[e].Vertices); i++ {
//          if balSepCache[g.Edges.Slice()[e].Vertices[i]] {
//              continue
//          }
//          for j := i + 1; j < len(g.Edges.Slice()[e].Vertices); j++ {
//              if balSepCache[g.Edges.Slice()[e].Vertices[j]] {
//                  continue
//              }
//              disjoint.Union(vertices[g.Edges.Slice()[e].Vertices[i]], vertices[g.Edges.Slice()[e].Vertices[j]])
//              break
//          }
//      }
//  }

//  //merge any components that share components of up
//  for i := 0; i < len(up)-1; i++ {
//      if balSepCache[up[i]] {
//          continue
//      }
//      for j := i + 1; j < len(up); j++ {
//          if balSepCache[up[j]] {
//              continue
//          }
//          disjoint.Union(vertices[up[i]], vertices[up[j]])
//          break
//      }
//  }

//  //sort each edge to a corresponding component
//  for i := range g.Edges.Slice() {
//      var vertexRep int
//      found := false
//      for _, v := range g.Edges.Slice()[i].Vertices {
//          if balSepCache[v] {
//              continue
//          }
//          vertexRep = v
//          found = true
//          break
//      }
//      if !found {
//          isolatedEdges = append(isolatedEdges, g.Edges.Slice()[i])
//          continue
//      }

//      slice, ok := comps[vertices[vertexRep].Find()]
//      if !ok {
//          newslice := make([]Edge, 0, g.Edges.Len())
//          comps[vertices[vertexRep].Find()] = newslice
//          slice = newslice
//      }

//      comps[vertices[vertexRep].Find()] = append(slice, g.Edges.Slice()[i])

//  }

//  if len(up) > 0 {
//      upComp = vertices[up[0]].Find()
//  }
//  if len(low) > 0 {
//      elementStart := vertices[low[0]].Find()

//      for _, i := range low[1:] {
//          if elementStart != vertices[i].Find() {
//              lowCompIndex = -1
//          }
//      }
//      if lowCompIndex != -1 {
//          lowComp = elementStart
//      }
//  }

//  currentIndex := 0
//  // Store the components as graphs
//  for k, _ := range comps {
//      if k == upComp {
//          upCompIndex = currentIndex
//      }
//      if k == lowComp {
//          lowCompIndex = currentIndex
//      }

//      g := Graph{Edges: Edges{slice: comps[k]}}
//      outputG = append(outputG, g)
//      currentIndex++
//  }

//  return outputG, upCompIndex, lowCompIndex, isolatedEdges
// }

// func CreateOrderingMap(edges Edges, vertices []int) []int {
// 	tmp := make([]int, edges.Len())
// 	tmp2 := make([]int, edges.Len())

// 	for i, e := range edges.Slice() {
// 		tmp[i] = len(Inter(e.Vertices, vertices))
// 		tmp2[i] = i
// 	}

// 	// log.Println("map b4", tmp2)
// 	// log.Println("values b4", tmp)
// 	// log.Println("values b4", edges)

// 	two := TwoSlicesInt{main_slice: tmp2, other_slice: tmp}
// 	sort.Sort(SortByOtherInt(two))

// 	// log.Println("map afterwards", tmp2)

// 	return tmp2
// }

// func FilterVerticesSort(edges Edges, vertices []int) []int {
// 	tmp := make([]int, edges.Len())

// 	for i, e := range edges.Slice() {
// 		tmp[i] = len(Inter(e.Vertices, vertices))
// 	}

// 	return tmp
// }

// ignoring special edges here, as this concerns search for separators,
// which may not use special edges
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

			//  var mux sync.Mutex
			//  mux.Lock() // ensure that hash is computed only on one gorutine at a time
			//  name = encode
			//  m[encode] = m[edges.Slice()[i].Name] + "'"
			//  encode++
			//  mux.Unlock()

			// }
			output = append(output, Edge{Name: name, Vertices: inter})
		}
	}

	return NewEdges(output)

}

func (g Graph) CheckNextSep(sep Edges, oldSep Edges) bool {

	verticesCurrent := append(g.Vertices())

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

// special edges not relevant for subedges
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
