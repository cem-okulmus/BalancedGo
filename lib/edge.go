package lib

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

// An Edge (used here for hyperedge) consists of a collection of vertices and a name
type Edge struct {
	Name     int
	Vertices []int // use integers for vertices
}

// FullString always prints the list of vertices of an edge, even if the edge is named
func (e Edge) FullString() string {
	var buffer bytes.Buffer
	mutex.RLock()
	defer mutex.RUnlock()
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
	mutex.RLock()
	defer mutex.RUnlock()
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

// Edges struct is a slice of Edge, defined for the use of the sort interface,
// as well as various other optimisations which are only possible on the slice level
type Edges struct {
	slice         []Edge
	vertices      []int
	hash          *uint64
	hashMux       *sync.Mutex
	duplicateFree bool
}

// NewEdges is a constructor for Edges
func NewEdges(slice []Edge) Edges {
	var hashMux sync.Mutex

	return Edges{slice: slice, hashMux: &hashMux}
}

// RemoveDuplicates removes duplicate edges from an Edges struct
func (e *Edges) RemoveDuplicates() {
	if e.duplicateFree {
		return
	}
	output := removeDuplicateEdges(e.slice)
	e.slice = output.slice
	e.duplicateFree = true
}

// FullString always prints the list of vertices of an edge, even if named
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

	if this.Hash() != other.Hash() {
		return false
	}

	return reflect.DeepEqual(this.slice, other.slice)
}

// Slice returns the internal slice of an Edges struct
func (e Edges) Slice() []Edge {
	return e.slice
}

// Len returns the length of the internal slice
func (e Edges) Len() int {
	return len(e.slice)
}

// Swap as used for the sort interface
func (e Edges) Swap(i, j int) {
	e.slice[i], e.slice[j] = e.slice[j], e.slice[i]
}

//lexicographic order on each edge
func (e Edges) Less(i, j int) bool {
	if len(e.slice[i].Vertices) < len(e.slice[j].Vertices) {
		return true
	} else if len(e.slice[i].Vertices) > len(e.slice[j].Vertices) {
		return false
	} else {
		for k := 0; k < len(e.slice[i].Vertices); k++ {
			ki := e.slice[i].Vertices[k]
			kj := e.slice[j].Vertices[k]

			if ki != kj {
				return ki < kj
			}
		}
	}
	return false //edges at i and j identical
}

//using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
func removeDuplicateEdges(elementsSlice []Edge) Edges {
	if len(elementsSlice) == 0 {
		return NewEdges([]Edge{})
	}
	elements := NewEdges(elementsSlice)
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

	return NewEdges(elements.slice[:j+1])
}

// subedges computes all subedges for an Edges slice.
// TODO: Unnecessarily adds empty edge
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

// getDegree returns the degree of a given node in the Edges slice
func getDegree(edges Edges, node int) int {
	var output int

	for _, e := range edges.Slice() {
		if mem(e.Vertices, node) {
			output++
		}
	}

	return output
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
		if o.Name == e.Name {
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

func (e Edge) numNeighboursOrder(l Edges, remaining []bool) int {
	output := 0

	for i := range l.Slice() {
		if remaining[i] && e.areNeighbours(l.Slice()[i]) {
			output++
		}
	}

	return output
}

// Vertices produces the union of all vertices from a slice of Edge
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

//Diff computes the set difference of edges based on names
func (e *Edges) Diff(other Edges) Edges {
	var output []Edge

	encounteredOther := make(map[int]struct{})

	for i := range other.slice {
		encounteredOther[other.slice[i].Name] = Empty
	}

	for j := range e.slice {
		if _, ok := encounteredOther[e.slice[j].Name]; !ok {
			output = append(output, e.slice[j])
		}
	}

	return NewEdges(output)
}

// FullString always prints the list of vertices of an edge, even if the edge is named
func (e Edge) FullStringInt() string {
	var buffer bytes.Buffer
	mutex.RLock()
	defer mutex.RUnlock()
	if e.Name > 0 {
		buffer.WriteString("E" + strconv.Itoa(e.Name))
	}
	buffer.WriteString(" (")
	for i, n := range e.Vertices {
		var s string
		s = fmt.Sprintf("V%v", n)
		buffer.WriteString(s)
		if i != len(e.Vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}
