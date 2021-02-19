package tests

// some basic unit tests for the subedge package

import (
	"testing"

	"github.com/cem-okulmus/BalancedGo/lib"
)

//TestSubedge creates a random scenario for the SepSub struct and checks if some subedges are being produced
func TestSubedge(t *testing.T) {
	graph, _ := getRandomGraph(20)
	sep := getRandomSep(graph, 10)

	test := lib.GetSepSub(graph.Edges, sep, sep.Len())

	count := 0
	for test.HasNext() {
		current := test.GetCurrent()
		// fmt.Println("gotten subedge", current)

		if !lib.Subset(current.Vertices(), sep.Vertices()) {
			t.Errorf("Invalid Subedge produced: not a subset of original")
		}

		count++
	}

	if count == 0 {
		t.Errorf("No subedges produced")
	}
}
