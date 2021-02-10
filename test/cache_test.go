package tests

// this is intended to provide some basic unit tests for lib.Cache

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

//getRandomEdge will produce a random Edge
func getRandomEdge() lib.Edge {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	arity := r.Intn(100) + 1
	var vertices []int
	name := r.Intn(1000)

	for i := 0; i < arity; i++ {
		vertices = append(vertices, r.Intn(1000)+i)
	}

	return lib.Edge{Name: name, Vertices: vertices}
}

//getRandomGraph will produce a random Graph
func getRandomGraph() lib.Graph {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	size := r.Intn(100) + 1
	special := r.Intn(10)

	var edges []lib.Edge
	var SpEdges []lib.Edges

	for i := 0; i < size; i++ {
		edges = append(edges, getRandomEdge())
	}

	for i := 0; i < special; i++ {
		SpEdges = append(SpEdges, getRandomEdges())
	}

	return lib.Graph{Edges: lib.NewEdges(edges), Special: SpEdges}
}

//getRandomEdges will produce a random Edges struct
func getRandomEdges() lib.Edges {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	k := r.Intn(10) + 1
	var edges []lib.Edge
	for j := 0; j < k; j++ {
		edges = append(edges, getRandomEdge())
	}

	return lib.NewEdges(edges)
}

func getRandomSep(g lib.Graph) lib.Edges {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	k := r.Intn(10) + 1

	var selection []int
	for j := 0; j < k; j++ {
		selection = append(selection, r.Intn(g.Edges.Len()))
	}

	return lib.GetSubset(g.Edges, selection)

}

func TestCache(t *testing.T) {
	randomGraph := getRandomGraph()
	randomSep := getRandomSep(randomGraph)
	randomSep2 := getRandomSep(randomGraph)
	randomSep3 := getRandomSep(randomGraph)

	var cache lib.Cache
	var cacheCopy lib.Cache

	cache.CopyRef(&cacheCopy)

	comps, _, _ := randomGraph.GetComponents(randomSep)

	cache.AddNegative(randomSep, comps[0])
	cache.AddPositive(randomSep2, comps[0])

	if !cacheCopy.CheckNegative(randomSep, comps) {
		t.Errorf("negative sep not properly cached")
	}
	if cacheCopy.CheckNegative(randomSep2, comps) && randomSep2.Hash() != randomSep.Hash() {
		t.Errorf("negative sep not properly cached")
	}

	if cacheCopy.CheckNegative(randomSep3, comps) && randomSep3.Hash() != randomSep.Hash() {
		t.Errorf("negative sep not properly cached")
	}

	if !cacheCopy.CheckPositive(randomSep2, comps) {
		t.Errorf("positive sep not properly cached")
	}

	if cacheCopy.CheckPositive(randomSep, comps) && randomSep2.Hash() != randomSep.Hash() {
		t.Errorf("positive sep not properly cached")
	}
	if cacheCopy.CheckPositive(randomSep3, comps) && randomSep2.Hash() != randomSep3.Hash() {
		t.Errorf("positive sep not properly cached")
	}

	if cacheCopy.Len() != 2 {
		t.Errorf("cache len doesn't work")
	}

}
