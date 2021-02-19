package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

//TestCover simply checks if cover can still run to the end, for a random input graph
func TestCover(t *testing.T) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	graph, _ := getRandomGraph(100)

	k := r.Intn(10) + 1

	c := lib.NewCover(k, []int{}, graph.Edges, graph.Vertices())

	for c.HasNext {
		c.NextSubset()
	}

}

func shuffle(input lib.Edges) lib.Edges {
	rand.Seed(time.Now().UTC().UnixNano())
	a := make([]lib.Edge, len(input.Slice()))
	copy(a, input.Slice())

	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}

	return lib.NewEdges(a)
}

//TestCover2 also checks if Cover works for non-empty Conn
func TestCover2(t *testing.T) {
	graph, _ := getRandomGraph(100)
	oldSep := getRandomSep(graph, 10)

	conn := lib.Inter(oldSep.Vertices(), graph.Vertices())
	bound := lib.FilterVertices(graph.Edges, oldSep.Vertices())

	gen := lib.NewCover(oldSep.Len(), conn, bound, graph.Vertices())

	for gen.HasNext {
		gen.NextSubset()
	}
}
