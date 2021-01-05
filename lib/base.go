package lib

import (
	"bytes"
	"sort"
)

var Empty struct{}

// func RemoveDuplicates(elements []int) []int {
//  // Use map to record duplicates as we find them.
//  encountered := make(map[int]struct{})
//  result := []int{}

//  for v := range elements {
//      if _, ok := encountered[elements[v]]; ok {
//          // Do not add duplicate.
//      } else {
//          // Record this element as an encountered element.
//          var Empty struct{}
//          encountered[elements[v]] = Empty
//          // Append to result slice.
//          result = append(result, elements[v])
//      }
//  }
//  // Return the new slice.
//  return result
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
//  encountered_b := map[int]bool{}

//  for _, b := range bs {
//      encountered_b[b] = true
//  }

//  var output []int

//  for _, a := range as {
//      if _, ok := encountered_b[a]; !ok {
//          output = append(output, a)
//      }
//  }
//  // if !reflect.DeepEqual(output, diff2(as, bs)) {
//  //  log.Panicf("What the hell?", output, diff2(as, bs))
//  // }
//  return output

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
	//var output []int
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

// func Diff(a, b []int) []int {
//  var output []int

//  // a = RemoveDuplicates(a)
//  // b = RemoveDuplicates(b)

//  sort.Ints(a)
//  sort.Ints(b)

//  i, j := 0, 0

//  for i < len(a) && j < len(b) {
//      if b[j] > a[i] {
//          output = append(output, a[i])
//          i++
//      } else if b[j] < a[i] {
//          j++
//      } else {
//          i++
//      }

//  }

//  for i < len(a) {
//      output = append(output, a[i])
//      i++

//  }

//  return output

// }

func Mem64(as []uint64, b uint64) bool {
	for _, a := range as {
		if a == b {
			return true
		}
	}
	return false
}

func DiffEdges(a Edges, e ...Edge) Edges {
	var output []Edge
	/// log.Println("Edges ", a, "Other ", e)

	var hashes []uint64
	for i := range e {
		hashes = append(hashes, e[i].Hash())
	}

	for i := range a.Slice() {
		if !Mem64(hashes, a.Slice()[i].Hash()) {
			output = append(output, a.Slice()[i])
		}
	}

	//  log.Println("Result ", output)

	return NewEdges(output)

}

// func DiffSpecial(a, b []Special) []Special {
// 	var output []Special
// OUTER:
// 	for _, n := range a {
// 		for _, k := range b {
// 			if reflect.DeepEqual(n, k) {
// 				continue OUTER
// 			}
// 		}
// 		output = append(output, n)
// 	}

// 	return output

// }

// func Inter(as, bs []int) []int {
//  encountered_b := make(map[int]struct{})
//  for _, b := range bs {
//      encountered_b[b] = Empty
//  }

//  var output []int

//  for _, a := range as {
//      if _, ok := encountered_b[a]; ok {
//          output = append(output, a)
//      }
//  }

//  return output

// }

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

func Subset(as []int, bs []int) bool {
	if len(as) == 0 {
		return true
	}
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

func Equiv(as []int, bs []int) bool {
	if len(as) == 0 {
		return true
	}
	encountered_b := make(map[int]struct{})
	encountered_a := make(map[int]struct{})
	var Empty struct{}
	for _, b := range bs {
		encountered_b[b] = Empty
	}

	for _, a := range as {
		if _, ok := encountered_b[a]; !ok {
			return false
		}
		encountered_a[a] = Empty
	}

	return true && len(encountered_a) == len(encountered_b)
}

// func Subset(a []int, b []int) bool {

// OUTER:
//  for _, n := range a {
//      for _, k := range b {
//          if k == n {
//              continue OUTER
//          }
//      }
//      return false
//  }
//  return true
// }

type TwoSlicesEdge struct {
	main_slice  []Edge
	other_slice []int
}

type TwoSlicesBool struct {
	main_slice  []bool
	other_slice []int
}

// type TwoSlicesInt struct {
// 	main_slice  []int
// 	other_slice []int
// }

type SortByOtherEdge TwoSlicesEdge

func (sbo SortByOtherEdge) Len() int {
	return len(sbo.main_slice)
}

func (sbo SortByOtherEdge) Swap(i, j int) {
	sbo.main_slice[i], sbo.main_slice[j] = sbo.main_slice[j], sbo.main_slice[i]
	sbo.other_slice[i], sbo.other_slice[j] = sbo.other_slice[j], sbo.other_slice[i]
}

func (sbo SortByOtherEdge) Less(i, j int) bool {
	return sbo.other_slice[i] > sbo.other_slice[j]
}

type SortByOtherBool TwoSlicesBool

func (sbo SortByOtherBool) Len() int {
	return len(sbo.main_slice)
}

func (sbo SortByOtherBool) Swap(i, j int) {
	sbo.main_slice[i], sbo.main_slice[j] = sbo.main_slice[j], sbo.main_slice[i]
	sbo.other_slice[i], sbo.other_slice[j] = sbo.other_slice[j], sbo.other_slice[i]
}

func (sbo SortByOtherBool) Less(i, j int) bool {
	return sbo.other_slice[i] > sbo.other_slice[j]
}

// type SortByOtherInt TwoSlicesInt

// func (sbo SortByOtherInt) Len() int {
// 	return len(sbo.main_slice)
// }

// func (sbo SortByOtherInt) Swap(i, j int) {
// 	sbo.main_slice[i], sbo.main_slice[j] = sbo.main_slice[j], sbo.main_slice[i]
// 	sbo.other_slice[i], sbo.other_slice[j] = sbo.other_slice[j], sbo.other_slice[i]
// }

// func (sbo SortByOtherInt) Less(i, j int) bool {
// 	return sbo.other_slice[i] > sbo.other_slice[j]
// }

func sortBySliceEdge(a []Edge, b []int) {
	tmp := make([]int, len(b))
	copy(tmp, b)
	two := TwoSlicesEdge{main_slice: a, other_slice: tmp}
	sort.Sort(SortByOtherEdge(two))
}

func sortBySliceBool(a []bool, b []int) {
	tmp := make([]int, len(b))
	copy(tmp, b)
	two := TwoSlicesBool{main_slice: a, other_slice: tmp}
	sort.Sort(SortByOtherBool(two))
}

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

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
