package main

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
)

type Edge struct {
	name     string
	vertices []int // use integers for vertices
}

func (e Edge) String() string {
	if e.name != "" {
		return e.name
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i, n := range e.vertices {
		var s string
		if m == nil {
			s = fmt.Sprintf("%v", n)
		} else {
			s = m[n]
		}
		buffer.WriteString(s)
		if i != len(e.vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

type Edges []Edge

func (s Edges) Len() int {
	return len(s)
}
func (s Edges) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Edges) Less(i, j int) bool {
	if len(s[i].vertices) < len(s[j].vertices) {
		return true
	} else if len(s[i].vertices) > len(s[j].vertices) {
		return false
	} else {
		for k := 0; k < len(s[i].vertices); k++ {
			if s[i].vertices[k] < s[j].vertices[k] {
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
		if reflect.DeepEqual(elements[j].vertices, elements[i].vertices) {
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

	powerSetSize := int(math.Pow(2, float64(len(e.vertices))))
	var index int
	for index < powerSetSize {
		var subSet []int

		for j, elem := range e.vertices {
			if index&(1<<uint(j)) > 0 {
				subSet = append(subSet, elem)
			}
		}
		output = append(output, Edge{vertices: subSet})
		index++
	}

	return output
}

func getDegree(edges Edges, node int) int {
	var output int

	for _, e := range edges {
		if mem(e.vertices, node) {
			output++
		}
	}

	return output
}

func (e Edge) intersect(l []Edge) Edge {
	var output Edge

OUTER:
	for _, n := range e.vertices {
		for _, o := range l { // skip all vertices n that are not conained in ALL edges of l
			if !o.contains(n) {
				continue OUTER
			}
		}
		output.vertices = append(output.vertices, n)
	}

	return output
}

func Contains(l []Edge, v int) bool {

	for _, e := range l {
		for _, a := range e.vertices {
			if a == v {
				return true
			}
		}
	}
	return false
}

func (e Edge) contains(v int) bool {
	for _, a := range e.vertices {
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
	for _, a := range o.vertices {
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
	for _, a := range o.vertices {
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

func Vertices(e []Edge) []int {
	var output []int
	for _, otherE := range e {
		output = append(output, otherE.vertices...)
	}
	return removeDuplicates(output)
}
