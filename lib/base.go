// This package provides various functions, data structures and methods to aid in the design of algorithms to
// compute structural decomposition methods.
package lib

import (
	"bytes"
	"sort"
)

// Empty used for maps of type struct{}
var Empty struct{}

//RemoveDuplicates is using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
func RemoveDuplicates(elements []int) []int {
	if len(elements) == 0 {
		return elements
	}
	sort.Ints(elements)

	j := 0
	for i := 1; i < len(elements); i++ {
		if elements[j] == elements[i] {
			continue
		}
		j++

		// only set what is required
		elements[j] = elements[i]
	}

	return elements[:j+1]
}

// mem checks if an integer b occurs inside a slice as
func mem(as []int, b int) bool {
	for _, a := range as {
		if a == b {
			return true
		}
	}
	return false
}

// Diff computes the set difference between two slices a b
func Diff(a, b []int) []int {
	output := make([]int, 0, len(a))

OUTER:
	for _, n := range a {
		for _, k := range b {
			if k == n {
				continue OUTER
			}
		}
		output = append(output, n)
	}

	return output
}

// mem64 is the same as mem, but for uint64
func mem64(as []uint64, b uint64) bool {
	for _, a := range as {
		if a == b {
			return true
		}
	}
	return false
}

// diffEdges computes the set difference between a and e
func diffEdges(a Edges, e ...Edge) Edges {
	var output []Edge
	var hashes []uint64

	for i := range e {
		hashes = append(hashes, e[i].Hash())
	}
	for i := range a.Slice() {
		if !mem64(hashes, a.Slice()[i].Hash()) {
			output = append(output, a.Slice()[i])
		}
	}

	return NewEdges(output)
}

// Inter is the set intersection between slices as and bs
func Inter(as, bs []int) []int {
	var output []int
OUTER:
	for _, a := range as {
		for _, b := range bs {
			if a == b {
				output = append(output, a)
				continue OUTER
			}
		}
	}

	return output
}

// Subset returns true if as subset of bs, false otherwise
func Subset(as []int, bs []int) bool {
	if len(as) == 0 {
		return true
	}
	encounteredB := make(map[int]struct{})
	var Empty struct{}
	for _, b := range bs {
		encounteredB[b] = Empty
	}

	for _, a := range as {
		if _, ok := encounteredB[a]; !ok {
			return false
		}
	}

	return true
}

type twoSlicesEdge struct {
	mainSlice  []Edge
	otherSlice []int
}

type twoSlicesBool struct {
	mainSlice  []bool
	otherSlice []int
}

type sortByOtherEdge twoSlicesEdge

func (sbo sortByOtherEdge) Len() int {
	return len(sbo.mainSlice)
}

func (sbo sortByOtherEdge) Swap(i, j int) {
	sbo.mainSlice[i], sbo.mainSlice[j] = sbo.mainSlice[j], sbo.mainSlice[i]
	sbo.otherSlice[i], sbo.otherSlice[j] = sbo.otherSlice[j], sbo.otherSlice[i]
}

func (sbo sortByOtherEdge) Less(i, j int) bool {
	return sbo.otherSlice[i] > sbo.otherSlice[j]
}

type sortByOtherBool twoSlicesBool

func (sbo sortByOtherBool) Len() int {
	return len(sbo.mainSlice)
}

func (sbo sortByOtherBool) Swap(i, j int) {
	sbo.mainSlice[i], sbo.mainSlice[j] = sbo.mainSlice[j], sbo.mainSlice[i]
	sbo.otherSlice[i], sbo.otherSlice[j] = sbo.otherSlice[j], sbo.otherSlice[i]
}

func (sbo sortByOtherBool) Less(i, j int) bool {
	return sbo.otherSlice[i] > sbo.otherSlice[j]
}

func sortBySliceEdge(a []Edge, b []int) {
	tmp := make([]int, len(b))
	copy(tmp, b)
	two := twoSlicesEdge{mainSlice: a, otherSlice: tmp}
	sort.Sort(sortByOtherEdge(two))
}

func sortBySliceBool(a []bool, b []int) {
	tmp := make([]int, len(b))
	copy(tmp, b)
	two := twoSlicesBool{mainSlice: a, otherSlice: tmp}
	sort.Sort(sortByOtherBool(two))
}

// PrintVertices will pretty print an int slice using the encodings in the m map
func PrintVertices(vertices []int) string {
	mutex.RLock()
	defer mutex.RUnlock()

	var buffer bytes.Buffer

	buffer.WriteString("(")
	for i, v := range vertices {
		buffer.WriteString(m[v])
		if i != len(vertices)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")

	return buffer.String()
}

// max returns the larger of two integers a and b
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
