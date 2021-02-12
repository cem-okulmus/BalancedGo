package lib

// cover.go implements an iterator for hypergraph covers, based on Samer and Gottlob 2009, as used in det-k-decomp

import (
	"log"
	"sort"
)

// Cover is used to quickly iterate over all valid hypergraph covers for a subset of vertices
type Cover struct {
	k          int          //maximal size of cover
	covered    map[int]int8 //map if each vertex is covered, and by how many edges
	uncovered  int          //number of vertices that needs to be covered
	InComp     []bool       //indicates for each Edge if its in comp. or not
	inCompSel  int          //number of edges inComp already selected
	covWeights []int        //summed weights from toCover
	bound      Edges        //the edges at boundary
	Subset     []int        //The current selection
	pos        int
	HasNext    bool
	first      bool
}

//NewCover acts as a constructor for Cover
func NewCover(K int, vertices []int, bound Edges, compVertices []int) Cover {
	covered := make(map[int]int8)

	for _, v := range vertices {
		covered[v] = 0
	}

	toCover := make([]int, len(bound.Slice()))
	inComp := make([]bool, len(bound.Slice()))

	for i, e := range bound.Slice() {
		sum := 0
		for _, v := range e.Vertices {
			_, ok := covered[v]
			if ok {
				sum++
			} else {
				if mem(compVertices, v) {
					if !inComp[i] {
						inComp[i] = true
					}
				}
			}
		}
		toCover[i] = sum
	}

	sortBySliceEdge(bound.Slice(), toCover)
	sortBySliceBool(inComp, toCover)
	sort.Slice(toCover, func(i, j int) bool { return toCover[i] > toCover[j] })

	covWeights := make([]int, len(bound.Slice()))
	sum := 0
	for i := len(toCover) - 1; i >= 0; i-- {
		sum = sum + toCover[i]
		covWeights[i] = sum
	}

	return Cover{k: K, covered: covered, uncovered: len(vertices),
		InComp: inComp, covWeights: covWeights, bound: bound, pos: 0, HasNext: true, first: true}
}

// NextSubset returns number of selected edges, or -1 if no alternative possible
func (c *Cover) NextSubset() int {
	if !c.first {
		if !c.backtrack() {
			// log.Println("No more covers possible.")
			return -1 // no more backtracking possible
		}
		c.pos++
	}
	c.first = false

	var covered bool
	if c.uncovered > 0 {
		covered = false
	} else {
		covered = true
	}

	// Big Loop here that continues until conditions met (or returns empty set)
	for ; !covered; c.pos++ {

		//check if remaining edges can cover all that's needed
		i := c.pos + (c.k - len(c.Subset))
		weight := 0
		if i < len(c.covWeights) {
			weight = c.covWeights[c.pos] - c.covWeights[i]
		} else {
			if c.pos < len(c.covWeights) {
				weight = c.covWeights[c.pos]
			} else {
				weight = 0
			}
		}

		if (weight < c.uncovered) || (weight == 0) {
			if !c.backtrack() {
				return -1 // no more backtracking possible
			}
			continue
		}

		//check if current edge lies in component (precomputation via inComp ?)
		selected := false

		if c.InComp[c.pos] || c.inCompSel > 0 || len(c.Subset) < (c.k-1) {
			for _, v := range c.bound.Slice()[c.pos].Vertices {
				_, ok := c.covered[v]
				if ok {
					selected = true
					break
				}
			}
		}

		//Actually add it to the edge
		if selected {
			c.Subset = append(c.Subset, c.pos)
			if c.InComp[c.pos] {
				c.inCompSel++
			}

			for _, v := range c.bound.Slice()[c.pos].Vertices {
				val, ok := c.covered[v]
				if ok {
					if val == 0 {
						c.uncovered--
					}
					c.covered[v]++
				}
			}

			if c.uncovered < 0 {
				log.Panicln("negative Uncovered, you messed something up!")
			}

			if c.uncovered == 0 {
				covered = true
			}
		}
	}

	return len(c.Subset)
}

// backtrack returns false if no more backtracking possible
func (c *Cover) backtrack() bool {
	if len(c.Subset) == 0 {
		c.HasNext = false
		return false
	}

	c.pos = c.Subset[len(c.Subset)-1]

	c.Subset = c.Subset[:len(c.Subset)-1]
	if c.InComp[c.pos] {
		c.inCompSel--
	}

	if c.inCompSel < 0 {
		log.Panicln("negative inCompSel, you messed something up!")
	}

	for _, v := range c.bound.Slice()[c.pos].Vertices {
		val, ok := c.covered[v]
		if ok {
			if val == 1 {
				c.uncovered++
			}
			c.covered[v]--

			if c.uncovered < 0 {
				log.Panicln("negative Uncovered, you messed something up!")
			}
		}
	}
	return true
}
