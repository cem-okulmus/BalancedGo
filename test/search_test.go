package tests

import (
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// a unit test for the parallel search using the struct Search

//TestSearchBal ensures that the parallel search for balanced separators always returns the same results,
// no matter how many splits are generated and run in parallel
func TestSearchBal(t *testing.T) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	randGraph, _ := getRandomGraph(20)
	k := r.Intn(5) + 1

	combinParallel := lib.SplitCombin(randGraph.Edges.Len(), k, runtime.GOMAXPROCS(-1), false)
	combinSeq := lib.SplitCombin(randGraph.Edges.Len(), k, 1, false)

	parallelSearch := lib.ParallelSearch{
		H:          &randGraph,
		Edges:      &randGraph.Edges,
		BalFactor:  2,
		Generators: combinParallel,
	}
	seqSearch := lib.ParallelSearch{
		H:          &randGraph,
		Edges:      &randGraph.Edges,
		BalFactor:  2,
		Generators: combinSeq,
	}
	pred := lib.BalancedCheck{}

	var allSepsSeq []lib.Edges
	var allSepsPar []lib.Edges

	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(pred) {
		sep := lib.GetSubset(randGraph.Edges, parallelSearch.Result)
		allSepsPar = append(allSepsPar, sep)
	}

	for ; !seqSearch.ExhaustedSearch; seqSearch.FindNext(pred) {
		sep := lib.GetSubset(randGraph.Edges, seqSearch.Result)
		allSepsSeq = append(allSepsSeq, sep)
	}

OUTER:
	for i := range allSepsSeq {
		sep := allSepsPar[i]

		for j := range allSepsPar {
			other := allSepsPar[j]
			if other.Hash() == sep.Hash() {
				continue OUTER // found matching sep
			}
		}

		t.Errorf("Mismatch in returned seps between sequential and parallel Search")
	}
}

//TestSearchPar ensures that the parallel search for good parents always returns the same results,
// no matter how many splits are generated and run in parallel
func TestSearchPar(t *testing.T) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	randGraph, _ := getRandomGraph(30)
	k := r.Intn(5) + 1
	prevSep := getRandomSep(randGraph, k)

	allowedParent := lib.FilterVertices(randGraph.Edges, prevSep.Vertices())

	combinParallel := lib.SplitCombin(allowedParent.Len(), k, runtime.GOMAXPROCS(-1), false)
	combinSeq := lib.SplitCombin(allowedParent.Len(), k, 1, false)

	parallelSearch := lib.ParallelSearch{
		H:          &randGraph,
		Edges:      &allowedParent,
		BalFactor:  2,
		Generators: combinParallel,
	}
	seqSearch := lib.ParallelSearch{
		H:          &randGraph,
		Edges:      &allowedParent,
		BalFactor:  2,
		Generators: combinSeq,
	}
	predPar := lib.ParentCheck{Conn: lib.Inter(prevSep.Vertices(), randGraph.Vertices()), Child: prevSep.Vertices()}

	var allSepsSeq []lib.Edges
	var allSepsPar []lib.Edges

	for ; !parallelSearch.ExhaustedSearch; parallelSearch.FindNext(predPar) {
		sep := lib.GetSubset(randGraph.Edges, parallelSearch.Result)
		allSepsPar = append(allSepsPar, sep)
	}

	for ; !seqSearch.ExhaustedSearch; seqSearch.FindNext(predPar) {
		sep := lib.GetSubset(randGraph.Edges, seqSearch.Result)
		allSepsSeq = append(allSepsSeq, sep)
	}

OUTER:
	for i := range allSepsSeq {
		sep := allSepsSeq[i]

		for j := range allSepsPar {
			other := allSepsPar[j]
			if other.Hash() == sep.Hash() {
				continue OUTER // found matching sep
			}
		}

		t.Errorf("Mismatch in returned seps between sequential and parallel Search")
	}
}
