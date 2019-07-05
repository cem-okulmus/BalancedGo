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
type Edges []Edge

func (s Edges) Len() int {
	return len(s)
}
func (s Edges) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Edges) Less(i, j int) bool {
	if len(s[i].Vertices) < len(s[j].Vertices) {
		return true
	} else if len(s[i].Vertices) > len(s[j].Vertices) {
		return false
	} else {
		for k := 0; k < len(s[i].Vertices); k++ {
			if s[i].Vertices[k] < s[j].Vertices[k] {
				return true
			}
		}
	}
	return false
}

func (s *Edges) append(e Edge) {
	*s = append(*s, e)
}

//using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
func removeDuplicateEdges(elements []Edge) []Edge {
	sort.Sort(Edges(elements))

	j := 0
	for i := 1; i < len(elements); i++ {
		if reflect.DeepEqual(elements[j].Vertices, elements[i].Vertices) {
			continue
		}
		j++

		// only set what is required
		elements[j] = elements[i]
	}

	return elements[:j+1]
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

	for _, e := range edges {
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

func (e Edge) numNeighboursOrder(l []Edge, remaining []bool) int {
	output := 0

	for i := range l {
		if remaining[i] && e.areNeighbours(l[i]) {
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
func Vertices(e []Edge) []int {
	var output []int
	for _, otherE := range e {
		output = append(output, otherE.Vertices...)
	}
	return RemoveDuplicates(output)
}
