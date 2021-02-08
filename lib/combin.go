package lib

// Based on "github.com/gonum/stat/combin" package:
// Copyright ©2016 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package combin implements routines involving combinatorics (permutations,
// combinations, etc.).

// binomial returns the binomial coefficient of (n,k), also commonly referred to
// as "n choose k".
//
// The binomial coefficient, C(n,k), is the number of unordered combinations of
// k elements in a set that is n elements big, and is defined as
//
//  C(n,k) = n!/((n-k)!k!)
//
// n and k must be non-negative with n >= k, otherwise binomial will panic.
// No check is made for overflow.
func binomial(n, k int) int {

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

// extendedBinom extends Binomial by summing over all calls Binomial(n,i), where k >= i >= 1.
func extendedBinom(n int, k int) int {
	var output int

	for i := k; i >= 1; i-- {
		output = output + binomial(n, i)
	}

	return output
}

// A CombinationIterator generates combinations iteratively.
type CombinationIterator struct {
	n           int
	k           int
	oldK        int // needed to compute current progress
	combination []int
	empty       bool
	stepSize    int
	extended    bool
	confirmed   bool
	balSep      bool // cache the result of balSep check
}

// getCombin produces a CombinationIterator
func getCombin(n int, k int) CombinationIterator {
	if k > n {
		k = n
	}
	return CombinationIterator{n: n, oldK: k, k: k, stepSize: 1, extended: true, confirmed: true}
}

// getCombinUnextend produces a CombinationIterator, with the flag extended set to false
func getCombinUnextend(n int, k int) CombinationIterator {
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
// Returns the number of steps performed
func (c *CombinationIterator) advance(step int) (bool, int) {
	if c.empty {
		return false, 0
	}
	if c.combination == nil {
		c.combination = make([]int, c.k)
		for i := range c.combination {
			c.combination[i] = i
		}
	} else {
		res, steps := nextCombinationStep(c.combination, c.n, c.k, step)
		c.empty = !res
		return res, steps
	}
	return true, step
}

// hasNext checks if the iterator still has new elements and advances the iterator if so
func (c *CombinationIterator) hasNext() bool {
	if !c.confirmed {
		return true
	}

	hasNext, stepsDone := c.advance(c.stepSize)
	if !hasNext {
		if c.k <= 1 || !c.extended {
			return false
		}

		c.k--
		c.combination, c.empty = nil, false   // discard old slice, reset flag
		c.advance(0)                          // initialize the iterator
		c.advance(c.stepSize - stepsDone - 1) // actually advance the iterator (-1 to count starting a new iterator)
	}

	c.confirmed = false
	return true
}

// confirm is used to double check the last element has been used. Useful only for concurrent searching
func (c *CombinationIterator) confirm() {
	c.balSep = false
	c.confirmed = true
}

//SplitCombin generates multiple iterators, splitting the search space into multiple "splits"
func SplitCombin(n int, k int, split int, unextended bool) []*CombinationIterator {
	if k > n {
		k = n
	}
	var output []*CombinationIterator

	initial := CombinationIterator{n: n, k: k, stepSize: split, extended: !unextended, confirmed: true}

	output = append(output, &initial)

	for i := 1; i < split; i++ {
		tempIter := CombinationIterator{n: n, k: k, stepSize: split, extended: !unextended, confirmed: true}

		tempIter.hasNext()
		nextCombinationStep(tempIter.combination, n, k, i)

		output = append(output, &tempIter)
	}

	return output
}

//combinatorialOrder computes the current combinatorial ordering of an iterator
func combinatorialOrder(combination []int) int {
	var output int

	for i := range combination {
		output = output + binomial(combination[i], i+1)
	}

	return output
}

// GetPercentage returns the current progress as a percentage,
// with 100% representing that all combinations have been visited
func (c CombinationIterator) GetPercentage() float32 {

	if !c.hasNext() {
		return 1.0
	}

	progressPresent := combinatorialOrder(c.combination) * 1.0

	if !c.extended {
		return float32(progressPresent) / float32(binomial(c.n, c.k))
	}

	allCombinations := 0
	progressPast := 0
	k := c.oldK

	for k >= 1 {
		allCombinations = allCombinations + binomial(c.n, k)
		k = k - 1
		if k > c.k {
			progressPast = progressPast + binomial(c.n, k)
		}
	}

	return (float32(progressPresent) + float32(progressPast)) / float32(allCombinations)
}
