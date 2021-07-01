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
		return 0
	}

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
	N           int
	K           int
	OldK        int // needed to compute current progress
	Combination []int
	Empty       bool
	StepSize    int
	Extended    bool
	Confirmed   bool
	BalSep      bool // cache the result of balSep check
}

// getCombin produces a CombinationIterator
func getCombin(n int, k int) CombinationIterator {
	if k > n {
		k = n
	}

	return CombinationIterator{N: n, OldK: k, K: k, StepSize: 1, Extended: true, Confirmed: true}
}

// getCombinUnextend produces a CombinationIterator, with the flag extended set to false
func getCombinUnextend(n int, k int) CombinationIterator {
	if k > n {
		k = n
	}
	return CombinationIterator{N: n, OldK: k, K: k, StepSize: 1, Extended: false, Confirmed: true}
}

// nextCombination generates the combination after s, overwriting s
func nextCombination(s []int, n, k int) bool {
	for j := k - 1; j >= 0; j-- {
		if s[j] == n+j-k {
			continue
		}

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
	if c.Empty {
		return false, 0
	}
	if c.Combination == nil {
		c.Combination = make([]int, c.K)
		for i := range c.Combination {
			c.Combination[i] = i
		}
	} else {
		res, steps := nextCombinationStep(c.Combination, c.N, c.K, step)
		c.Empty = !res
		return res, steps
	}
	return true, step
}

// CheckFound returns the current value of the cached result
func (c *CombinationIterator) CheckFound() bool {

	return c.BalSep
}

// Found is used by the search to cache the previous check result
func (c *CombinationIterator) Found() {
	c.BalSep = true
}

// GetNext returns the currently selected combination
func (c *CombinationIterator) GetNext() []int {
	return c.Combination
}

// HasNext checks if the iterator still has new elements and advances the iterator if so
func (c *CombinationIterator) HasNext() bool {
	if !c.Confirmed {
		return true
	}

	hasNext, stepsDone := c.advance(c.StepSize)
	if !hasNext {
		if c.K <= 1 || !c.Extended {
			return false
		}

		c.K--
		c.Combination, c.Empty = nil, false   // discard old slice, reset flag
		c.advance(0)                          // initialize the iterator
		c.advance(c.StepSize - stepsDone - 1) // actually advance the iterator (-1 to count starting a new iterator)
	}

	c.Confirmed = false
	return true
}

// Confirm is used to double check the last element has been used. Useful only for concurrent searching
func (c *CombinationIterator) Confirm() {
	c.BalSep = false
	c.Confirmed = true
}

//SplitCombin generates multiple iterators, splitting the search space into multiple "splits"
func SplitCombin(n int, k int, split int, unextended bool) []Generator {
	if k > n {
		k = n
	}
	var output []Generator

	initial := CombinationIterator{N: n, OldK: k, K: k, StepSize: split, Extended: !unextended, Confirmed: true}
	output = append(output, &initial)

	for i := 1; i < split; i++ {
		tempIter := CombinationIterator{N: n, OldK: k, K: k, StepSize: split, Extended: !unextended, Confirmed: true}
		tempIter.HasNext()
		nextCombinationStep(tempIter.Combination, n, k, i)
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
	nomin, denom := c.getPercentageFull()

	return float32(nomin) / float32(denom)
}

// GetPercentage returns the current progress as a percentage,
// with 100% representing that all combinations have been visited
func (c CombinationIterator) getPercentageFull() (int, int) {

	progressPresent := combinatorialOrder(c.Combination) * 1.0

	if !c.Extended {

		if !c.HasNext() {
			return binomial(c.N, c.K), binomial(c.N, c.K)
		}

		return progressPresent, binomial(c.N, c.K)
	}

	allCombinations := 0
	progressPast := 0
	k := c.OldK

	for k >= 1 {

		allCombinations = allCombinations + binomial(c.N, k)
		k = k - 1
		if k > c.K {
			progressPast = progressPast + binomial(c.N, k)
		}
	}

	if !c.HasNext() {
		return allCombinations, allCombinations
	}

	return progressPresent + progressPast, allCombinations
}

// GetPercentagesSlice calculates a total percentage from a slice of CombinationIterators
func GetPercentagesSlice(cs []*CombinationIterator) (int, int) {
	var nominTotal int
	var denomTotal int

	for i := range cs {
		nomin, denom := cs[i].getPercentageFull()
		nominTotal = nominTotal + nomin
		denomTotal = denomTotal + denom
	}

	return nominTotal, denomTotal

}
