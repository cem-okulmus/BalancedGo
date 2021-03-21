package tests

// this is intended to provide some basic unit tests for lib.Cache

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

var EDGE int

//getRandomEdge will produce a random Edge
func getRandomEdge(size int) lib.Edge {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	arity := r.Intn(size) + 1
	var vertices []int
	name := r.Intn(size*10) + EDGE + 1
	EDGE = name

	for i := 0; i < arity; i++ {

		vertices = append(vertices, r.Intn(size*10)+i+1)
	}

	return lib.Edge{Name: name, Vertices: vertices}
}

//getRandomGraph will produce a random Graph
func getRandomGraph(size int) (lib.Graph, map[string]int) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	card := r.Intn(size) + 1

	var edges []lib.Edge
	var SpEdges []lib.Edges

	for i := 0; i < card; i++ {
		edges = append(edges, getRandomEdge(size))
	}

	outString := lib.Graph{Edges: lib.NewEdges(edges), Special: SpEdges}.ToHyberBenchFormat()
	parsedGraph, pGraph := lib.GetGraph(outString)

	return parsedGraph, pGraph.Encoding
}

//getRandomEdges will produce a random Edges struct
func getRandomEdges(size int) lib.Edges {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	k := r.Intn(size) + 1
	var edges []lib.Edge
	for j := 0; j < k; j++ {
		edges = append(edges, getRandomEdge(size*10))
	}

	return lib.NewEdges(edges)
}

func getRandomSep(g lib.Graph, size int) lib.Edges {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	k := r.Intn(size) + 1

	var selection []int
	for j := 0; j < k; j++ {
		selection = append(selection, r.Intn(g.Edges.Len()))
	}

	return lib.GetSubset(g.Edges, selection)
}

func TestCache(t *testing.T) {
	randomGraph, _ := getRandomGraph(100)
	randomSep := getRandomSep(randomGraph, 10)
	randomSep2 := getRandomSep(randomGraph, 10)
	randomSep3 := getRandomSep(randomGraph, 10)

	var cache lib.Cache
	var cacheCopy lib.Cache

	cache.CopyRef(&cacheCopy)

	comps, _, _ := randomGraph.GetComponents(randomSep)

	if len(comps) == 0 { // randomSep covers the entire hypergraph. Nothing you can do
		return
	}

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
