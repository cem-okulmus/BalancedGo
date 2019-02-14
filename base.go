package main

import "reflect"

func removeDuplicates(elements []int) []int {
	// Use map to record duplicates as we find them.
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

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

func diffEdges(a, b []Edge) []Edge {
	var output []Edge
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

func inter(a, b []int) []int {
	var output []int

	for _, n := range a {
		found := false
		for _, k := range b {
			if k == n {
				found = true
				break
			}
		}
		if found {
			output = append(output, n)
		}
	}

	return output

}

func subset(a []int, b []int) bool {

OUTER:
	for _, n := range a {
		for _, k := range b {
			if k == n {
				continue OUTER
			}
		}
		return false
	}
	return true
}
