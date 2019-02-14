package main

import (
	"bytes"
	"fmt"
)

type Special struct {
	vertices []int
	edges    []Edge
}

func (s Special) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, e := range s.vertices {
		if s.edges[0].m == nil {
			buffer.WriteString(fmt.Sprintf("%d", e))
		} else {
			buffer.WriteString(s.edges[0].m[e])
		}
		if i != len(s.vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

func VerticesSpecial(sp []Special) []int {
	var output []int

	for _, s := range sp {
		output = append(output, s.vertices...)
	}

	return removeDuplicates(output)
}

func (s Special) areSNeighbours(o Edge, sep []int) bool {

OUTER:
	for _, a := range o.nodes {
		for _, ss := range sep { // don't consider sep vertices for neighbouring edges
			if ss == a {
				continue OUTER
			}
		}
		if Contains(s.edges, a) {
			return true
		}
	}
	return false
}
