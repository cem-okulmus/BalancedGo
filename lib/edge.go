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

func (e Edge) FullString() string {
	var buffer bytes.Buffer

	if e.Name > 0 {
		buffer.WriteString(m[e.Name])
	}
	buffer.WriteString(" (")
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
	slice    []Edge
	vertices []int
	hash     *uint32
}

func NewEdges(slice []Edge) Edges {
	return Edges{slice: slice}
}

func (e *Edges) Clear() {
	e.vertices = e.vertices[:0]
}

func (e Edges) FullString() string {

	var buffer bytes.Buffer
	buffer.WriteString("{")

	for i, e2 := range e.slice {
		buffer.WriteString(e2.FullString())
		if i != len(e.slice)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("}")
	return buffer.String()
}

func (e Edges) String() string {

	var buffer bytes.Buffer
	buffer.WriteString("{")

	for i, e2 := range e.slice {
		buffer.WriteString(e2.String())
		if i != len(e.slice)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("}")
	return buffer.String()
}

func equalEdges(this, other Edges) bool {
	return reflect.DeepEqual(this.slice, other.slice)
}

func (e Edges) Slice() []Edge {
	return e.slice
}

func (s Edges) Len() int {
	return len(s.slice)
}
func (s Edges) Swap(i, j int) {
	s.slice[i], s.slice[j] = s.slice[j], s.slice[i]
}

//lexicographic order on each edge
func (s Edges) Less(i, j int) bool {
	if len(s.slice[i].Vertices) < len(s.slice[j].Vertices) {
		return true
	} else if len(s.slice[i].Vertices) > len(s.slice[j].Vertices) {
		return false
	} else {
		for k := 0; k < len(s.slice[i].Vertices); k++ {
			k_i := s.slice[i].Vertices[k]
			k_j := s.slice[j].Vertices[k]

			if k_i != k_j {
				return k_i < k_j
			}
		}
	}
	return false //edges at i and j identical
}

func (s *Edges) append(es ...Edge) {
	for _, e := range es {
		s.slice = append(s.slice, e)
	}
	if len(s.vertices) > 0 {
		s.vertices = s.vertices[:0] // do this to preserve allocated memory
	}
	if s.hash != nil {
		s.hash = nil
	}
}

//using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
func removeDuplicateEdges(elementsSlice []Edge) Edges {
	elements := Edges{slice: elementsSlice}
	sort.Sort(elements)

	j := 0
	for i := 1; i < len(elements.Slice()); i++ {
		if reflect.DeepEqual(elements.slice[j].Vertices, elements.slice[i].Vertices) {
			continue
		}
		j++

		// only set what is required
		elements.slice[j] = elements.slice[i]
	}

	return Edges{slice: elements.slice[:j+1]}
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

	for _, e := range edges.Slice() {
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

	for _, e := range l.Slice() {
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

	for i := range l.Slice() {
		if remaining[i] && e.areNeighbours(l.Slice()[i]) {
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
		for _, otherE := range e.Slice() {
			output = append(output, otherE.Vertices...)
		}
		e.vertices = RemoveDuplicates(output)
	}

	return e.vertices
}
