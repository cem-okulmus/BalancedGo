// Based on "github.com/gonum/stat/combin" package:
// Copyright ©2016 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package combin implements routines involving combinatorics (permutations,
// combinations, etc.).
package main

const (
	badNegInput = "combin: negative input"
	badSetSize  = "combin: n < k"
	badInput    = "combin: wrong input slice length"
)

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
		panic(badNegInput)
	}
	if n < k {
		panic(badSetSize)
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

// CombinationGenerator generates combinations iteratively. Combinations may be
// called to generate all combinations collectively.
type CombinationGenerator struct {
	n        int
	k        int
	previous []int
	empty    bool
}

// NewCombinationGenerator returns a CombinationGenerator for generating the
// combinations of k elements from a set of size n.
//
// n and k must be non-negative with n >= k, otherwise NewCombinationGenerator
// will panic.
func NewCombinationGenerator(n, k int) *CombinationGenerator {
	return &CombinationGenerator{
		n: n,
		k: k,
	}
}

// Next advances the iterator if there are combinations remaining to be generated,
// and returns false if all combinations have been generated. Next must be called
// to initialize the first value before calling Combination or Combination will
// panic. The value returned by Combination is only changed during calls to Next.
//
// Step simply advances the iterator multiple steps at a time
// Returns the number of steps perfomed
func (c *CombinationGenerator) Next(step int) (bool, int) {
	if c.empty {
		return false, 0
	}
	if c.previous == nil {
		c.previous = make([]int, c.k)
		for i := range c.previous {
			c.previous[i] = i
		}
	} else {
		res, steps := nextCombinationStep(c.previous, c.n, c.k, step)
		c.empty = !res
		return res, steps
	}
	return true, step
}

// Combination generates the next combination. If next is non-nil, it must have
// length k and the result will be stored in-place into combination. If combination
// is nil a new slice will be allocated and returned. If all of the combinations
// have already been constructed (Next() returns false), Combination will panic.
//
// Next must be called to initialize the first value before calling Combination
// or Combination will panic. The value returned by Combination is only changed
// during calls to Next.
func (c *CombinationGenerator) Combination(combination []int) []int {
	if c.empty {
		panic("no combinations left")
	}
	if c.previous == nil {
		panic("combin: Combination called before Next")
	}
	if combination == nil {
		combination = make([]int, c.k)
	}
	if len(combination) != c.k {
		panic(badInput)
	}
	copy(combination, c.previous)
	return combination
}

// nextCombination generates the combination after s, overwriting the input value.
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

// returns whether the iterator could be advanced step many times, and the number of steps that were possible (useful for extended combin)
func nextCombinationStep(s []int, n, k, step int) (bool, int) {
	for i := 0; i < step; i++ {
		if !nextCombination(s, n, k) {
			return false, i
		}
	}

	return true, step
}
