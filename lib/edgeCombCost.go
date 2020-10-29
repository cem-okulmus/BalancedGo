package lib

import (
	"fmt"
)

// An EdgesCostMap associates costs to combinations of edges
type EdgesCostMap struct {
	k        int
	original []int
	edges    map[int]int
	e2c      map[uint64]float64
}

func encodeComb(comb []int, all map[int]int) uint64 {
	code := uint64(0)
	for _, e := range comb {
		p := uint(all[e])
		code = code | uint64(1)<<p
	}
	return code
}

func decodeComb(code uint64, all []int, maxElem int) []int {
	comb := make([]int, maxElem)
	p := 0
	for i := 0; i < 64; i++ {
		//fmt.Println("code=", code, "->", strconv.FormatInt(code, 2))
		//fmt.Println("1 <<", i)
		if (code & (uint64(1) << uint(i))) != 0 {
			comb[p] = all[i]
			p++
		}
	}
	return comb
}

// Init an EdgeCostMap with the edges on which it will work
func (m *EdgesCostMap) Init(edges []int, k int) {
	if k <= 0 {
		fmt.Println(k, "must be positive.")
	}
	m.k = k
	m.original = make([]int, len(edges))
	m.edges = make(map[int]int)
	for i, e := range edges {
		m.original[i] = e
		m.edges[e] = i
	}
	m.e2c = make(map[uint64]float64)
	fmt.Println("m.e2c=", m.e2c)
}

// Put the cost of an edge comibnation into the map
func (m *EdgesCostMap) Put(edgeComb []int, c float64) {
	code := encodeComb(edgeComb, m.edges)
	fmt.Println(code, c)
	m.e2c[code] = c
}

// Cost of an edge combination
func (m *EdgesCostMap) Cost(edgeComb []int) float64 {
	code := encodeComb(edgeComb, m.edges)
	if c, ok := m.e2c[code]; ok {
		return c
	}
	fmt.Println("Illegal edge combination:", edgeComb)
	return -1.0
}

// Records in this map
func (m *EdgesCostMap) Records() ([][]int, []float64) {
	combs := make([][]int, len(m.e2c))
	costs := make([]float64, len(m.e2c))
	p := 0
	for code, v := range m.e2c {
		combs[p] = decodeComb(code, m.original, m.k)
		costs[p] = v
		p++
	}
	return combs, costs
}
