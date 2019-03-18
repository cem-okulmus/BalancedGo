package main

import "log"

type Combin struct {
	n           int
	k           int
	current     CombinationGenerator
	combination []int
	left        int
	confirmed   bool
}

func ExtendedBinom(n int, k int) int {
	var output int

	for i := k; i >= 1; i-- {
		output = output + Binomial(n, i)
	}

	return output
}

func getCombin(n int, k int) Combin {
	if k > n {
		k = n
	}
	return Combin{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), left: ExtendedBinom(n, k), confirmed: true}
}

func getCombinUnextend(n int, k int) Combin {
	if k > n {
		k = n
	}
	return Combin{n: n, k: k, current: *NewCombinationGenerator(n, k), combination: make([]int, k), left: Binomial(n, k), confirmed: true}
}

func (c *Combin) hasNext() bool {
	if !c.confirmed {
		return true
	}

	if c.left <= 0 {
		return false
	}
	if !c.current.Next() {
		if c.k == 1 {
			log.Panicf("can't find more elements in generator (wrong initialisation? ) ", c)
		} else {
			c.k--
			c.current = *NewCombinationGenerator(c.n, c.k)
			c.current.Next()
		}
	}
	if len(c.combination) != c.k {
		if len(c.combination) < c.k {
			c.combination = make([]int, c.k) // generate new slice, if old is too small
		} else {
			c.combination = c.combination[0:c.k] // "shrink" combinations for smaller k
		}
	}
	c.left--
	c.current.Combination(c.combination)
	c.confirmed = false
	return true
}

func (c *Combin) confirm() {
	c.confirmed = true
}

func splitCombin(n int, k int, split int, unextended bool) []*Combin {
	if k > n {
		k = n
	}
	var output []*Combin

	var number int

	if unextended {
		number = Binomial(n, k)
	} else {
		number = ExtendedBinom(n, k)
	}

	remainder := number % split
	quotient := ((number - remainder) / split)

	var splitPoints []int

	end := quotient

	for i := 0; i < split; i++ {
		if i < remainder { // increase the first part of the generators by 1 to cover remainder
			splitPoints = append(splitPoints, end+1)
			end = end + quotient + 1
		} else {
			splitPoints = append(splitPoints, end)
			end = end + quotient
		}
	}

	//produce the Combin in one pass
	temp := *NewCombinationGenerator(n, k)
	start := 0
	for i := 0; i < len(splitPoints); i++ { //skip forward
		generator := temp
		generator.previous = make([]int, k)
		copy(generator.previous, temp.previous)

		output = append(output, &Combin{n: n, k: k, current: generator, combination: make([]int, k), left: splitPoints[i] - start, confirmed: true})

		if i == len(splitPoints)-1 {
			continue // skip the compuation in the final iteration
		}
		for j := 0; j < (splitPoints[i] - start); j++ {
			if !temp.Next() {
				if k > 1 {
					k--
					temp = *NewCombinationGenerator(n, k)
					temp.Next()
				} else {
					log.Panicf("Reached end of combin during initialisation! k:%v", k)
				}
			}
		}

		start = splitPoints[i]
	}

	return output
}

// func main() {

// 	n := 8
// 	k := 2

// 	procNum := 24

// 	combins := splitCombin(n, k, procNum)
// 	for _, c := range combins {
// 		i := 0
// 		for c.hasNext() {
// 			// fmt.Println("left: ", c.left)
// 			c.confirm()
// 			fmt.Println(c.combination)
// 			i++
// 		}
// 		fmt.Println("Work items ", i)
// 	}

// }
