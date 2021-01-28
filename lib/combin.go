// Based on "github.com/gonum/stat/combin" package:
// Copyright ©2016 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package combin implements routines involving combinatorics (permutations,
// combinations, etc.).
package lib

// Binomial returns the binomial coefficient of (n,k), also commonly referred to
// as "n choose k".
//
// The binomial coefficient, C(n,k), is the number of unordered combinations of
// k elements in a set that is n elements big, and is defined as
//
//  C(n,k) = n!/((n-k)!k!)
//
// n and k must be non-negative with n >= k, otherwise Binomial will panic.
// No check is made for overflow.
func Binomial(n, k int) int {

	if n < 0 || k < 0 {
		panic("combin: negative input")
	}
	if n < k {
		panic("combin: n < k")
	}
	// (n,k) = (n, n-k)
	if k > n/2 {
		k = n - k
	}
	b := 1
	for i := 1; i <= k; i++ {
		b = (n - k + i) * b / i
	}
	return b
}

// ExtendedBinom extends Binomial by summing over all calls Binomial(n,i), where k >= i >= 1.
func ExtendedBinom(n int, k int) int {
	var output int

	for i := k; i >= 1; i-- {
		output = output + Binomial(n, i)
	}

	return output
}

// A CombinationIterator generates combinations iteratively.
type CombinationIterator struct {
	n           int
	k           int
	oldK        int // needed to compute current progress
	Combination []int
	empty       bool
	stepSize    int
	extended    bool
	confirmed   bool
	BalSep      bool // cache the result of balSep check
}

func GetCombin(n int, k int) CombinationIterator {
	if k > n {
		k = n
	}
	return CombinationIterator{n: n, oldK: k, k: k, stepSize: 1, extended: true, confirmed: true}
}

func GetCombinUnextend(n int, k int) CombinationIterator {
	if k > n {
		k = n
	}
	return CombinationIterator{n: n, oldK: k, k: k, stepSize: 1, extended: false, confirmed: true}
}

// nextCombination generates the combination after s, overwriting s
func nextCombination(s []int, n, k int) bool {
	for j := k - 1; j >= 0; j-- {
		if s[j] == n+j-k {
			continue
		}
		// for i := 0; i < 1000; i++ {
		// 	s[j]++
		// 	s[j]--
		// }

		s[j]++

		for l := j + 1; l < k; l++ {
			s[l] = s[j] + l - j
		}
		return true
	}
	return false
}

// nextCombinationStep returns whether the iterator could be advanced step many times,
// and the number of steps that were possible (useful for extended combin)
func nextCombinationStep(s []int, n, k, step int) (bool, int) {
	for i := 0; i < step; i++ {
		if !nextCombination(s, n, k) {
			return false, i
		}
	}

	return true, step
}

// advances the iterator if there are combinations remaining to be generated,
// and returns false if all combinations have been generated. Next must be called
// to initialize the first value before calling Combination or Combination will
// panic. The value returned by Combination is only changed during calls to Next.
//
// Step simply advances the iterator multiple steps at a time
// Returns the number of steps perfomed
func (c *CombinationIterator) advance(step int) (bool, int) {
	if c.empty {
		return false, 0
	}
	if c.Combination == nil {
		c.Combination = make([]int, c.k)
		for i := range c.Combination {
			c.Combination[i] = i
		}
	} else {
		res, steps := nextCombinationStep(c.Combination, c.n, c.k, step)
		c.empty = !res
		return res, steps
	}
	return true, step
}

func (c *CombinationIterator) HasNext() bool {
	if !c.confirmed {
		return true
	}

	hasNext, stepsDone := c.advance(c.stepSize)
	if !hasNext {
		if c.k <= 1 || !c.extended {
			return false
		}

		c.k--
		c.Combination, c.empty = nil, false   // discard old slice, reset flag
		c.advance(0)                          // initialize the iterator
		c.advance(c.stepSize - stepsDone - 1) // actually advance the iterator (-1 to count starting a new iterator)
	}

	c.confirmed = false
	return true
}

func (c *CombinationIterator) Confirm() {
	c.BalSep = false
	c.confirmed = true
}

func SplitCombin(n int, k int, split int, unextended bool) []*CombinationIterator {
	if k > n {
		k = n
	}
	var output []*CombinationIterator

	initial := CombinationIterator{n: n, k: k, stepSize: split, extended: !unextended, confirmed: true}

	output = append(output, &initial)

	for i := 1; i < split; i++ {
		tempIter := CombinationIterator{n: n, k: k, stepSize: split, extended: !unextended, confirmed: true}

		tempIter.HasNext()
		nextCombinationStep(tempIter.Combination, n, k, i)

		output = append(output, &tempIter)
	}

	return output
}

func CombinatorialOrder(combination []int) int {
	var output int

	for i := range combination {
		output = output + Binomial(combination[i], i+1)
	}

	return output
}

// get the current progress as a percentage, with 100% representing that all combinations have been visited
func (c CombinationIterator) GetPercentage() float32 {

	if !c.HasNext() {
		return 1.0
	}

	progressPresent := CombinatorialOrder(c.Combination) * 1.0

	if !c.extended {

		return float32(progressPresent) / float32(Binomial(c.n, c.k))
	} else {

		allCombinatistions := 0
		progressPast := 0
		k := c.oldK

		for k >= 1 {
			allCombinatistions = allCombinatistions + Binomial(c.n, k)
			k = k - 1

			if k > c.k {
				progressPast = progressPast + Binomial(c.n, k)
			}
		}

		return (float32(progressPresent) + float32(progressPast)) / float32(allCombinatistions)
	}
}

// func allCombinations(index int, c *CombinationIterator) int {
// 	count := 0

// 	for c.HasNext() {
// 		//fmt.Println(index, " Checking combin ", c.Combination)
// 		c.Confirm()
// 		count++
// 	}

// 	return count
// }

// func main() {

// 	// dat, err := ioutil.ReadFile("hypergraphs/adlerexample.hg")
// 	// check(err)
// 	// parsedGraph := getGraph(string(dat))
// 	// //runtime.GOMAXPROCS(1)

// 	// e1 := parsedGraph.Edges[0]
// 	// e2 := parsedGraph.Edges[1]
// 	// e3 := parsedGraph.Edges[2]
// 	// e4 := parsedGraph.Edges[3]
// 	// e5 := parsedGraph.Edges[4]
// 	// e6 := parsedGraph.Edges[5]
// 	// e7 := parsedGraph.Edges[6]
// 	// e8 := parsedGraph.Edges[7]

// 	// fmt.Printf("%v %v\n", e1, Edge{Vertices: e1.Vertices})
// 	// fmt.Printf("%v %v\n", e2, Edge{Vertices: e2.Vertices})
// 	// fmt.Printf("%v %v\n", e3, Edge{Vertices: e3.Vertices})
// 	// fmt.Printf("%v %v\n", e4, Edge{Vertices: e4.Vertices})
// 	// fmt.Printf("%v %v\n", e5, Edge{Vertices: e5.Vertices})
// 	// fmt.Printf("%v %v\n", e6, Edge{Vertices: e6.Vertices})
// 	// fmt.Printf("%v %v\n", e7, Edge{Vertices: e7.Vertices})
// 	// fmt.Printf("%v %v\n", e8, Edge{Vertices: e8.Vertices})

// 	// H := Graph{Edges: []Edge{e2, e3, e4}}
// 	// Sp := []Special{Special{Edges: []Edge{Edge{Vertices: []int{e5.Vertices[0], e5.Vertices[2]}}, e1}, Vertices:
// Vertices([]Edge{Edge{Vertices: []int{e5.Vertices[0], e5.Vertices[2]}}, e1})}}

// 	// fmt.Println("H: ", H)
// 	// fmt.Println("Sp: ", Sp)

// 	// edges := CutEdges(parsedGraph.Edges, append(H.Vertices(), VerticesSpecial(Sp)...))

// 	// fmt.Println("Bionomial", Binomial(len(edges), 2))
// 	// generators := SplitCombin(len(edges), 2, runtime.GOMAXPROCS(-1), true)
// 	// for _, e := range edges {

// 	// 	fmt.Println(e)

// 	// }

// 	// BalancedFactor = 2

// 	// logActive(false)
// 	// for {

// 	// 	var found []int
// 	// 	parallelSearch(H, Sp, edges, &found, generators)

// 	// 	if len(found) == 0 { // meaning that the search above never found anything
// 	// 		fmt.Println("nothing more found")
// 	// 		break
// 	// 	} else {
// 	// 		fmt.Println("Balsep", GetSubset(edges, found))
// 	// 	}
// 	// }

// 	n := 10
// 	k := 3
// 	sum := 0
// 	ch := make(chan int)
// 	runtime.GOMAXPROCS(1)
// 	generators := SplitCombin(n, k, runtime.GOMAXPROCS(-1), true)

// 	for i, g := range generators {
// 		go func(i int, g *CombinationIterator, ch chan int) {

// 			res := allCombinations(i, g)
// 			//fmt.Println("Combin: ", res)
// 			ch <- res
// 		}(i, g, ch)
// 	}

// 	for i := range generators {

// 		temp := <-ch
// 		sum = sum + temp + i - i
// 	}

// 	fmt.Println("Sum: ", sum)
// 	fmt.Println("Bionomial", Binomial(n, k))
// 	fmt.Println("Extended:", ExtendedBinom(n, k))
// }
