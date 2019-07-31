package lib

import (
	"fmt"
	"log"
	"sort"
)

type Cover struct {
	K          int          //maximal size of cover
	coverSize  int          //current size of cover
	covered    map[int]int8 //map if each vertex is covered, and by how many edges
	Uncovered  int          //number of vertices that needs to be covered
	inComp     []bool       //indicates for each Edge if its in comp. or not
	inCompSel  int          //number of edges inComp already selected
	covWeights []int        //summed weights from toCover
	Bound      Edges        //the edges at boundary
	Subset     []int        //The current selection
	pos        int
	HasNext    bool
	first      bool
}

// func DivideCompEdges(GEdges Edges, HEdges Edges, Connector []int) Edges {
// 	covered := make(map[int]bool)
// 	var inner Edges
// 	var bound Edges

// 	for _, v := range HEdges.Vertices() {
// 		covered[v] = true
// 	}

// 	for _, e := range GEdges.Slice() {
// 		innerEdge := false

// 		for _, v := range e.Vertices {
// 			if covered[v] {
// 				innerEdge := true
// 				break
// 			}
// 		}

// 		if innerEdge {
// 			inner.append(e)
// 		} else {
// 			outer.append(e)
// 		}
// 	}

// }

func NewCover(K int, vertices []int, bound Edges, comp Edges) Cover {

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
				if Mem(comp.Vertices(), v) {
					if !inComp[i] {
						inComp[i] = true
					}
				}
			}
		}
		toCover[i] = sum
	}

	//fmt.Println("to Cover: ", toCover)

	sort.Slice(bound.Slice(), func(i, j int) bool { return toCover[i] > toCover[j] })
	sort.Slice(inComp, func(i, j int) bool { return toCover[i] > toCover[j] })
	sort.Slice(toCover, func(i, j int) bool { return toCover[i] > toCover[j] })

	//fmt.Println("to Cover (post): ", toCover)

	covWeights := make([]int, len(bound.Slice()))
	sum := 0
	for i := len(toCover) - 1; i >= 0; i-- {
		sum = sum + toCover[i]
		covWeights[i] = sum
	}

	return Cover{K: K, coverSize: len(vertices), covered: covered, Uncovered: len(vertices),
		inComp: inComp, covWeights: covWeights, Bound: bound, pos: 0, HasNext: true, first: true}

}

// Returns number of selected edges, or -1 if no alternative possible
func (c *Cover) NextSubset() int {
	if !c.first {
		if !c.backtrack() {
			return -1 // no more backtracking possible
		}
	}
	c.first = false

	var covered bool
	if c.Uncovered > 0 {
		covered = false
	} else {
		covered = true
	}

	// Big Loop here that continues until conditions met (or returns empty set)
	for ; !covered; c.pos++ {

		//check if remaining edges can cover all that's needed
		i := c.pos + (c.K - len(c.Subset))
		log.Println("i ", i, " pos ", c.pos, " size ", len(c.covWeights), " len ", len(c.Subset))
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
		log.Println("weight ", weight, " uncovered ", c.Uncovered)
		if (weight < c.Uncovered) || (weight == 0) {
			if !c.backtrack() {
				return -1 // no more backtracking possible
			}
			continue
		}

		//	log.Println("Edge actual ", c.Bound.Slice()[c.pos])
		//	log.Println("Current selection: ", GetSubset(c.Bound, c.Subset))
		//check if current edge lies in component (precomputation via inComp ?)
		selected := false

		// fmt.Println("c.inComp[c.pos]", c.inComp[c.pos])
		// fmt.Println("c.inCompSel > 0", c.inCompSel > 0)
		// fmt.Println("len(c.Subset) < (c.K-1) ", len(c.Subset) < (c.K-1))

		if c.inComp[c.pos] || c.inCompSel > 0 || len(c.Subset) < (c.K-1) {
			for _, v := range c.Bound.Slice()[c.pos].Vertices {
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
			if c.inComp[c.pos] {
				c.inCompSel++
			}

			for _, v := range c.Bound.Slice()[c.pos].Vertices {
				val, ok := c.covered[v]
				if ok {
					if val == 0 {
						c.Uncovered--
					}
					c.covered[v]++
				}
			}

			if c.Uncovered < 0 {
				log.Panicln("negative Uncovered, you messed something up!")
			}

			if c.Uncovered == 0 {
				covered = true
			}

		}
	}

	return len(c.Subset)
}

// returns false if no more backtracking possible
func (c *Cover) backtrack() bool {

	if len(c.Subset) == 0 {
		c.HasNext = false
		return false
	}

	log.Println("Subset before ", c.Subset)

	c.pos = c.Subset[len(c.Subset)-1]
	c.Subset = c.Subset[:len(c.Subset)-1]
	if c.inComp[c.pos] {
		c.inCompSel--
	}
	log.Println("Subset after ", c.Subset)

	if c.inCompSel < 0 {
		log.Panicln("negative inCompSel, you messed something up!")
	}

	for _, v := range c.Bound.Slice()[c.pos].Vertices {
		val, ok := c.covered[v]
		if ok {
			if val == 1 {
				c.Uncovered++
			}
			c.covered[v]--

			if c.Uncovered < 0 {
				log.Panicln("negative Uncovered, you messed something up!")
			}
		}
	}
	c.pos++
	return true

}

func testCover() {

	e1 := Edge{Vertices: []int{1, 2, 3, 4}}
	e2 := Edge{Vertices: []int{5, 6, 7, 8}}
	e3 := Edge{Vertices: []int{4, 9, 10}}
	e4 := Edge{Vertices: []int{5, 11, 12}}

	edges := Edges{slice: []Edge{e1, e2, e3, e4}}

	//comp := Edges{Slice: []Edge{e3, e4}}
	//	sep := Edges{Slice: []Edge{e1, e2}}

	c := NewCover(3, []int{3, 4, 5, 6}, edges, edges)

	for c.HasNext {
		out := c.NextSubset()
		if out > 0 {
			fmt.Println("Subset: ", GetSubset(edges, c.Subset))
		}
	}

}
