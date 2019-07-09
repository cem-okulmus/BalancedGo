package lib

import (
	"reflect"
	"sort"
)

// func RemoveDuplicates(elements []int) []int {
// 	// Use map to record duplicates as we find them.
// 	encountered := make(map[int]struct{})
// 	result := []int{}

// 	for v := range elements {
// 		if _, ok := encountered[elements[v]]; ok {
// 			// Do not add duplicate.
// 		} else {
// 			// Record this element as an encountered element.
// 			var Empty struct{}
// 			encountered[elements[v]] = Empty
// 			// Append to result slice.
// 			result = append(result, elements[v])
// 		}
// 	}
// 	// Return the new slice.
// 	return result
// }

//using an algorithm from "SliceTricks" https://github.com/golang/go/wiki/SliceTricks
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

// func Diff(as, bs []int) []int {
// 	encountered_b := map[int]bool{}

// 	for _, b := range bs {
// 		encountered_b[b] = true
// 	}

// 	var output []int

// 	for _, a := range as {
// 		if _, ok := encountered_b[a]; !ok {
// 			output = append(output, a)
// 		}
// 	}
// 	// if !reflect.DeepEqual(output, diff2(as, bs)) {
// 	// 	log.Panicf("What the hell?", output, diff2(as, bs))
// 	// }
// 	return output

// }

func Mem(as []int, b int) bool {
	for _, a := range as {
		if a == b {
			return true
		}
	}
	return false
}

func Diff(a, b []int) []int {
	var output []int

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

// func Diff(a, b []int) []int {
// 	var output []int

// 	// a = RemoveDuplicates(a)
// 	// b = RemoveDuplicates(b)

// 	sort.Ints(a)
// 	sort.Ints(b)

// 	i, j := 0, 0

// 	for i < len(a) && j < len(b) {
// 		if b[j] > a[i] {
// 			output = append(output, a[i])
// 			i++
// 		} else if b[j] < a[i] {
// 			j++
// 		} else {
// 			i++
// 		}

// 	}

// 	for i < len(a) {
// 		output = append(output, a[i])
// 		i++

// 	}

// 	return output

// }

func DiffEdges(a []Edge, e Edge) []Edge {
	var output []Edge

	for _, n := range a {
		length := len(Inter(n.Vertices, e.Vertices))
		if (length > 0) && (length < (len(e.Vertices))) {
			output = append(output, n)
		}
	}

	return output

}

func DiffSpecial(a, b []Special) []Special {
	var output []Special
OUTER:
	for _, n := range a {
		for _, k := range b {
			if reflect.DeepEqual(n, k) {
				continue OUTER
			}
		}
		output = append(output, n)
	}

	return output

}

func Inter(as, bs []int) []int {
	encountered_b := make(map[int]struct{})
	var Empty struct{}
	for _, b := range bs {
		encountered_b[b] = Empty
	}

	var output []int

	for _, a := range as {
		if _, ok := encountered_b[a]; ok {
			output = append(output, a)
		}
	}

	return output

}

// func Inter(as, bs []int) []int {
// 	var output []int
// OUTER:
// 	for _, a := range as {
// 		for _, b := range bs {
// 			if a == b {
// 				output = append(output, a)
// 				continue OUTER
// 			}
// 		}

// 	}

// 	return output
// }

func Subset(as []int, bs []int) bool {
	encountered_b := make(map[int]struct{})
	var Empty struct{}
	for _, b := range bs {
		encountered_b[b] = Empty
	}

	for _, a := range as {
		if _, ok := encountered_b[a]; !ok {
			return false
		}
	}

	return true
}

// func Subset(a []int, b []int) bool {

// OUTER:
// 	for _, n := range a {
// 		for _, k := range b {
// 			if k == n {
// 				continue OUTER
// 			}
// 		}
// 		return false
// 	}
// 	return true
// }