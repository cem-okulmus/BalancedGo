package tests

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// max returns the larger of two integers a and b
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// a unit test for the parallel search using the struct Search

//TestSearchBal ensures that the parallel search for balanced separators always returns the same results,
// no matter how many splits are generated and run in parallel
func TestSearchBal(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	randGraph, _ := getRandomGraph(20)
	k := r.Intn(5) + 1

	split := runtime.GOMAXPROCS(-1)

	combinParallel := lib.SplitCombin(randGraph.Edges.Len(), k, split, false)
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
		sep := allSepsSeq[i]

		for j := range allSepsPar {
			other := allSepsPar[j]
			if other.Hash() == sep.Hash() {
				continue OUTER // found matching sep
			}
		}

		if len(allSepsSeq) != len(allSepsPar) {

			fmt.Println("Graph", randGraph)
			fmt.Println("k: ", k)

			combinParallel2 := lib.SplitCombin(randGraph.Edges.Len(), k, split, false)
			combinSeq2 := lib.SplitCombin(randGraph.Edges.Len(), k, 1, false)

			fmt.Print("\n All stuff in combinPar: ")
			for _, combin := range combinParallel2 {

				fmt.Print("\n")

				for combin.HasNext() {
					j := combin.GetNext()
					fmt.Print(j)
					combin.Confirm()
				}

				fmt.Print("\n")

			}

			fmt.Print("\n\n All stuff in combinSeq2: ")
			for _, combin := range combinSeq2 {

				for combin.HasNext() {
					j := combin.GetNext()
					fmt.Print(j)
					combin.Confirm()
				}

			}

			fmt.Println("\n\n Number of splits in parallel: ", len(combinParallel))

			fmt.Println("split var: ", split)

			fmt.Println("Seps found by seq, ", allSepsSeq)
			fmt.Println("Seps found by par, ", allSepsPar)

		}

		t.Errorf("Mismatch in returned seps between sequential and parallel Search")
	}
}
