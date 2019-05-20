package main

type CombinIterator struct {
	n           int
	k           int
	current     CombinationGenerator
	combination []int
	stepSize    int
	extended    bool
	confirmed   bool
}

func ExtendedBinom(n int, k int) int {
	var output int

	for i := k; i >= 1; i-- {
		output = output + Binomial(n, i)
	}

	return output
}

func getCombin(n int, k int) CombinIterator {
	if k > n {
		k = n
	}
	return CombinIterator{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), stepSize: 1, extended: true, confirmed: true}
}

func getCombinUnextend(n int, k int) CombinIterator {
	if k > n {
		k = n
	}
	return CombinIterator{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), stepSize: 1, extended: false, confirmed: true}
}

func (c *CombinIterator) hasNext() bool {
	if !c.confirmed {
		return true
	}

	hasNext, steps := c.current.Next(c.stepSize)
	if !hasNext {
		if c.k == 1 || !c.extended {
			return false
		} else {
			c.k--
			c.current = *NewCombinationGenerator(c.n, c.k)
			c.current.Next(0)                  // initialize the iterator
			c.current.Next(c.stepSize - steps) // actually advance the iterator
		}
	}
	if len(c.combination) != c.k {
		if len(c.combination) < c.k {
			c.combination = make([]int, c.k) // generate new slice, if old is too small
		} else {
			c.combination = c.combination[0:c.k] // "shrink" combinations for smaller k
		}
	}

	c.current.Combination(c.combination)
	c.confirmed = false
	return true
}

func (c *CombinIterator) confirm() {
	c.confirmed = true
}

func splitCombin(n int, k int, split int, unextended bool) []*CombinIterator {
	if k > n {
		k = n
	}
	var output []*CombinIterator

	initial := CombinIterator{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), stepSize: split, extended: !unextended, confirmed: true}

	output = append(output, &initial)

	for i := 1; i < split; i++ {
		tempIter := CombinIterator{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), stepSize: split, extended: !unextended, confirmed: true}
		tempIter.hasNext()
		nextCombinationStep(tempIter.current.previous, n, k, i)

		output = append(output, &tempIter)
	}

	return output
}

// func main() {

// 	n := 10
// 	k := 2

// 	combination := make([]int, k)
// 	gen := NewCombinationGenerator(n, k)

// 	for {
// 		res, _ := gen.Next(1)
// 		if !res {
// 			break
// 		}
// 		gen.Combination(combination)
// 		fmt.Println("Combin: ", combination)
// 	}
// }
