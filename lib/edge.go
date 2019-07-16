package lib

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
)

// An Edge (used here for hyperedge) consists of a collection of vertices and a name
type Edge struct {
	Name     int
	Vertices []int // use integers for vertices
}

func (e Edge) String() string {
	if e.Name > 0 {
		return m[e.Name]
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i, n := range e.Vertices {
		var s string
		if m == nil {
			s = fmt.Sprintf("%v", n)
		} else {
			s = m[n]
		}
		buffer.WriteString(s)
		if i != len(e.Vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

// A slice of Edge, defined for the use of the sort interface
type Edges struct {
	Slice    []Edge
	vertices []int
}

func (s Edges) Len() int {
	return len(s.Slice)
}
func (s Edges) Swap(i, j int) {
	s.Slice[i], s.Slice[j] = s.Slice[j], s.Slice[i]
}

func (s Edges) Less(i, j int) bool {
	if len(s.Slice[i].Vertices) < len(s.Slice[j].Vertices) {
		return true
	} else if len(s.Slice[i].Vertices) > len(s.Slice[j].Vertices) {
		return false
	} else {
		for k := 0; k < len(s.Slice[i].Vertices); k++ {
			if s.Slice[i].Vertices[k] < s.Slice[j].Vertices[k] {
				return true
			}
		}
	}
	return false
}

func (s *Edges) append(e Edge) {
	s.Slice = append(s.Slice, e)
}

//using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
func removeDuplicateEdges(elementsSlice []Edge) Edges {
	elements := Edges{Slice: elementsSlice}
	sort.Sort(elements)

	j := 0
	for i := 1; i < len(elements.Slice); i++ {
		if reflect.DeepEqual(elements.Slice[j].Vertices, elements.Slice[i].Vertices) {
			continue
		}
		j++

		// only set what is required
		elements.Slice[j] = elements.Slice[i]
	}

	return Edges{Slice: elements.Slice[:j+1]}
}

// Unnessarily adds empty edge
func (e Edge) subedges() []Edge {
	var output []Edge

	powerSetSize := int(math.Pow(2, float64(len(e.Vertices))))
	var index int
	for index < powerSetSize {
		var subSet []int

		for j, elem := range e.Vertices {
			if index&(1<<uint(j)) > 0 {
				subSet = append(subSet, elem)
			}
		}
		output = append(output, Edge{Vertices: subSet})
		index++
	}

	return output
}

func getDegree(edges Edges, node int) int {
	var output int

	for _, e := range edges.Slice {
		if Mem(e.Vertices, node) {
			output++
		}
	}

	return output
}

func (e Edge) intersect(l []Edge) Edge {
	var output Edge

OUTER:
	for _, n := range e.Vertices {
		for _, o := range l { // skip all vertices n that are not conained in ALL edges of l
			if !o.contains(n) {
				continue OUTER
			}
		}
		output.Vertices = append(output.Vertices, n)
	}

	return output
}

// checks if vertex i is contained in a slice of edges

func (l Edges) Contains(v int) bool {

	for _, e := range l.Slice {
		for _, a := range e.Vertices {
			if a == v {
				return true
			}
		}
	}
	return false
}

func Contains(l []Edge, v int) bool {

	for _, e := range l {
		for _, a := range e.Vertices {
			if a == v {
				return true
			}
		}
	}
	return false
}

func (e Edge) contains(v int) bool {
	for _, a := range e.Vertices {
		if a == v {
			return true
		}
	}
	return false
}

func (e Edge) containedIn(l []Edge) bool {
	for _, o := range l {
		if reflect.DeepEqual(o, e) {
			return true
		}
	}
	return false
}

func (e Edge) areNeighbours(o Edge) bool {
	for _, a := range o.Vertices {
		if e.contains(a) {
			return true
		}
	}
	return false
}

func (e Edge) numIndicent(l []Edge) int {
	output := 0

	for i := range l {
		if e.areNeighbours(l[i]) {
			output++
		}
	}

	return output
}

func (e Edge) numNeighboursOrder(l Edges, remaining []bool) int {
	output := 0

	for i := range l.Slice {
		if remaining[i] && e.areNeighbours(l.Slice[i]) {
			output++
		}
	}

	return output
}

func (e Edge) areSNeighbours(o Edge, sep []int) bool {
	if reflect.DeepEqual(e, o) {
		return true
	}

OUTER:
	for _, a := range o.Vertices {
		for _, s := range sep { // don't consider sep vertices for neighbouring edges
			if s == a {
				continue OUTER
			}
		}
		if e.contains(a) {
			return true
		}
	}
	return false
}

// produces the union of all vertices from a slice of Edge
func (e *Edges) Vertices() []int {
	if len(e.vertices) == 0 {
		var output []int
		for _, otherE := range e.Slice {
			output = append(output, otherE.Vertices...)
		}
		e.vertices = RemoveDuplicates(output)
	}

	return e.vertices
}
