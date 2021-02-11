package tests

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

func TestCover(t *testing.T) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	graph := getRandomGraph()

	k := r.Intn(10) + 1

	c := NewCover(k, []int{}, graph.Edges, graph.Vertices())

	for c.HasNext {
		c.NextSubset()
	}

}

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

func TestCover2(t *testing.T) {

	graph := getRandomGraph()
	oldSep := getRandomSep(graph)

	conn := Inter(oldSep.Vertices(), graph.Vertices())
	bound := FilterVertices(graph.Edges, oldSep.Vertices())

	gen := NewCover(oldSep.Len(), conn, bound, graph.Vertices())

	for gen.HasNext {
		gen.NextSubset()
	}
}
