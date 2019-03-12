package main

import (
	"reflect"
)

func removeDuplicates(elements []int) []int {
	// Use map to record duplicates as we find them.
	encountered := make(map[int]struct{})
	result := []int{}

	for v := range elements {
		if _, ok := encountered[elements[v]]; ok {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			var Empty struct{}
			encountered[elements[v]] = Empty
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

// func diff(as, bs []int) []int {
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

func diff(a, b []int) []int {
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

// func diff(a, b []int) []int {
// 	var output []int

// 	// a = removeDuplicates(a)
// 	// b = removeDuplicates(b)

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

func diffEdges(a []Edge, e Edge) []Edge {
	var output []Edge

	for _, n := range a {
		length := len(inter(n.nodes, e.nodes))
		if (length > 0) && (length < (len(e.nodes))) {
			output = append(output, n)
		}
	}

	return output

}

func diffSpecial(a, b []Special) []Special {
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

func inter(as, bs []int) []int {
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

// func inter(as, bs []int) []int {
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

func subset(as []int, bs []int) bool {
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

// func subset(a []int, b []int) bool {

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
