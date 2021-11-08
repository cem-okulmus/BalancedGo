package lib

import (
	"log"
)

// An EdgesCostMap associates costs to combinations of edges
type EdgesCostMap struct {
	e2c map[uint64]float64
}

func hash(comb []int) uint64 {
	var h uint64 = 31
	for _, e := range comb {
		h = 31*h + uint64(e)
	}
	return h
}

// Init an EdgeCostMap with the edges on which it will work
func (m *EdgesCostMap) Init() {
	m.e2c = make(map[uint64]float64)
}

// Put the cost of an edge combination into the map
func (m *EdgesCostMap) Put(edgeComb []int, c float64) {
	code := hash(edgeComb)
	if _, ok := m.e2c[code]; ok {
		log.Panicln("Edge combination", edgeComb, "already present")
	}
	m.e2c[code] = c
}

// Cost of an edge combination
func (m *EdgesCostMap) Cost(edgeComb []int) float64 {
	code := hash(edgeComb)
	if c, ok := m.e2c[code]; ok {
		return c
	}
	log.Panicln("Illegal edge combination:", edgeComb)
	return -1.0
}

// Records in this EdgesCostMap
func (m EdgesCostMap) Records() ([]uint64, []float64) {
	wComb, wCost := make([]uint64, len(m.e2c)), make([]float64, len(m.e2c))
	i := 0
	for k, c := range m.e2c {
		wComb[i] = k
		wCost[i] = c
		i++
	}
	return wComb, wCost
}
