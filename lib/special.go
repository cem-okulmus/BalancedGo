package lib

import (
	"bytes"
	"fmt"
)

// A Special Edge is a collection of edges, seen as one edge
type Special struct {
	Vertices []int
	Edges    []Edge
}

func (s Special) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, v := range s.Vertices {
		if m == nil {
			buffer.WriteString(fmt.Sprintf("%d", v))
		} else {
			buffer.WriteString(m[v])
		}
		if i != len(s.Vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

func VerticesSpecial(sp []Special) []int {
	var output []int

	for _, s := range sp {
		output = append(output, s.Vertices...)
	}

	return RemoveDuplicates(output)
}

func (s Special) areSNeighbours(o Edge, sep []int) bool {

OUTER:
	for _, a := range o.Vertices {
		for _, ss := range sep { // don't consider sep vertices for neighbouring edges
			if ss == a {
				continue OUTER
			}
		}
		if Contains(s.Edges, a) {
			return true
		}
	}
	return false
}
