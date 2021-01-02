package lib

import (
	"log"
	"math/rand"
	"sort"
	"time"
)

type Cover struct {
	K          int          //maximal size of cover
	covered    map[int]int8 //map if each vertex is covered, and by how many edges
	Uncovered  int          //number of vertices that needs to be covered
	InComp     []bool       //indicates for each Edge if its in comp. or not
	inCompSel  int          //number of edges inComp already selected
	covWeights []int        //summed weights from toCover
	Bound      Edges        //the edges at boundary
	Subset     []int        //The current selection
	pos        int
	HasNext    bool
	first      bool
}

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
				if Mem(compVertices, v) {
					if !inComp[i] {
						inComp[i] = true
					}
				}
			}
		}
		toCover[i] = sum
	}
	// fmt.Println("cover pre", toCover)
	// fmt.Println("inComp pre", inComp)
	// fmt.Println("Bound pre", bound)
	sortBySliceEdge(bound.Slice(), toCover)
	sortBySliceBool(inComp, toCover)
	sort.Slice(toCover, func(i, j int) bool { return toCover[i] > toCover[j] })
	// fmt.Println("Bound post", bound)
	// fmt.Println("inComp post", inComp)

	covWeights := make([]int, len(bound.Slice()))
	sum := 0
	for i := len(toCover) - 1; i >= 0; i-- {
		sum = sum + toCover[i]
		covWeights[i] = sum
	}

	return Cover{K: K, covered: covered, Uncovered: len(vertices),
		InComp: inComp, covWeights: covWeights, Bound: bound, pos: 0, HasNext: true, first: true}

}

// Returns number of selected edges, or -1 if no alternative possible
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
	if c.Uncovered > 0 {
		covered = false
	} else {
		covered = true
	}

	// Big Loop here that continues until conditions met (or returns empty set)
	for ; !covered; c.pos++ {
		// fmt.Println("Current: ", append(c.Subset, c.pos), "Uncovered ", c.Uncovered)
		//check if remaining edges can cover all that's needed
		i := c.pos + (c.K - len(c.Subset))
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

		// fmt.Println("weight ", weight)
		if (weight < c.Uncovered) || (weight == 0) {
			if !c.backtrack() {
				// log.Println("No more covers available.")
				return -1 // no more backtracking possible
			}
			continue
		}

		//  log.Println("Edge actual ", c.Bound.Slice()[c.pos])
		//  log.Println("Current selection: ", GetSubset(c.Bound, c.Subset))
		//check if current edge lies in component (precomputation via inComp ?)
		selected := false

		// fmt.Println("c.inComp[c.pos]", c.inComp[c.pos])
		// fmt.Println("c.inCompSel > 0", c.inCompSel > 0)
		// fmt.Println("len(c.Subset) < (c.K-1) ", len(c.Subset) < (c.K-1))

		if c.InComp[c.pos] || c.inCompSel > 0 || len(c.Subset) < (c.K-1) {
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
			if c.InComp[c.pos] {
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

	c.pos = c.Subset[len(c.Subset)-1]

	c.Subset = c.Subset[:len(c.Subset)-1]
	if c.InComp[c.pos] {
		c.inCompSel--
	}

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
	return true

}

// func testCover() {

// 	e1 := Edge{Vertices: []int{1, 2, 3, 4}}
// 	e2 := Edge{Vertices: []int{5, 6, 7, 8}}
// 	e3 := Edge{Vertices: []int{4, 9, 10}}
// 	e4 := Edge{Vertices: []int{5, 11, 12}}

// 	edges := Edges{slice: []Edge{e1, e2, e3, e4}}

// 	//comp := Edges{Slice: []Edge{e3, e4}}
// 	//  sep := Edges{Slice: []Edge{e1, e2}}

// 	c := NewCover(3, []int{3, 4, 5, 6}, edges, edges.Vertices())

// 	for c.HasNext {
// 		out := c.NextSubset()
// 		if out > 0 {
// 			fmt.Println("Subset: ", GetSubset(edges, c.Subset))
// 		}
// 	}

// }

func shuffle(input Edges) Edges {
	rand.Seed(time.Now().UTC().UnixNano())
	a := make([]Edge, len(input.Slice()))
	copy(a, input.Slice())

	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}

	return NewEdges(a)
}

// func test2Cover() {
// 	dat, _ := ioutil.ReadFile("/home/cem/Dubois-026.xml_goodorder.hg")
// 	//dat, _ := ioutil.ReadFile("/home/cem/Dubois-026.xml_badorder.hg")
// 	parsedGraph, parsedParseGraph := GetGraph(string(dat))
// 	e25 := parsedParseGraph.GetEdge("E25 ( xL75J, xL23J, xL24J )")
// 	e26 := parsedParseGraph.GetEdge("E26 ( xL24J, xL76J, xL77J )")
// 	e27 := parsedParseGraph.GetEdge("E27 ( xL25J, xL76J, xL77J )")
// 	e28 := parsedParseGraph.GetEdge("E28 ( xL25J, xL26J, xL75J )")

// 	fmt.Println("Graph ")
// 	for _, e := range parsedGraph.Edges.Slice() {
// 		fmt.Println(e.FullString())
// 	}

// 	fmt.Println("\n\n", e25.FullString())
// 	fmt.Println(e28.FullString())

// 	oldSep := NewEdges([]Edge{e25, e28})
// 	largerGraph := NewEdges(append(parsedGraph.Edges.Slice(), []Edge{e26, e27, e25, e28}...))

// 	conn := Inter(oldSep.Vertices(), parsedGraph.Vertices())
// 	bound := FilterVertices(largerGraph, oldSep.Vertices())

// 	fmt.Println("larger graph ", largerGraph)
// 	fmt.Println("bound, ", bound)
// 	fmt.Println("Conn, ", Edge{Vertices: conn})

// 	gen := NewCover(2, conn, bound, parsedGraph.Edges.Vertices())

// 	for gen.HasNext {
// 		gen.NextSubset()
// 		if len(gen.Subset) == 0 {
// 			break
// 		}
// 		cover := GetSubset(bound, gen.Subset)
// 		fmt.Println("\033[33m Selection: ", cover, "\033[0m")

// 	}

// }
