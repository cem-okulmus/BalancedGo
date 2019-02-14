package main

import (
	"bytes"
	"fmt"
	"reflect"
)

type Edge struct {
	nodes []int // use integers for nodes
	m     map[int]string
}

func removeDuplicatesEdges(elements []Edge) []Edge {
	var output []Edge

OUTER:
	for _, e := range elements {
		for _, o := range output {
			if reflect.DeepEqual(o, e) {
				continue OUTER
			}
		}
		output = append(output, e)
	}

	return output
}

func (e Edge) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i, n := range e.nodes {
		var s string
		if e.m == nil {
			s = fmt.Sprintf("%v", n)
		} else {
			s = e.m[n]
		}
		buffer.WriteString(s)
		if i != len(e.nodes)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

func (e Edge) subedges() []Edge {
	var output []Edge

	for i := 0; i < len(e.nodes); i++ {
		if i == 0 {
			output = append(output, Edge{nodes: []int{e.nodes[0]}})
		} else {
			for j := 0; j < i; j++ {

				output = append(output, Edge{nodes: append(output[j].nodes, e.nodes[i])})
			}
		}
	}

	return output
}

func (e Edge) intersect(l []Edge) Edge {
	var output Edge

OUTER:
	for _, n := range e.nodes {
		for _, o := range l { // skip all nodes n that are not conained in ALL edges of l
			if !o.contains(n) {
				continue OUTER
			}
		}
		output.nodes = append(output.nodes, n)
	}

	return output
}

func Contains(l []Edge, v int) bool {

	for _, e := range l {
		for _, a := range e.nodes {
			if a == v {
				return true
			}
		}
	}
	return false
}

func (e Edge) contains(v int) bool {
	for _, a := range e.nodes {
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
	for _, a := range o.nodes {
		if e.contains(a) {
			return true
		}
	}
	return false
}

func (e Edge) areSNeighbours(o Edge, sep []int) bool {
	if reflect.DeepEqual(e, o) {
		return true
	}

OUTER:
	for _, a := range o.nodes {
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
		output = append(output, otherE.nodes...)
	}
	return removeDuplicates(output)
}
