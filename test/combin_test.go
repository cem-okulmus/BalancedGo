package tests

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// Test if SplitCombin produces the same search space indepenent of the split size

func TestCombin(t *testing.T) {

	for x := 0; x < 10; x++ {

		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)

		randGraph, _ := getRandomGraph(20)
		k := r.Intn(5) + 1
		prevSep := getRandomSep(randGraph, 5)

		k = max(k, prevSep.Len())

		allowedParent := lib.FilterVertices(randGraph.Edges, prevSep.Vertices())

		combinParallel := lib.SplitCombin(allowedParent.Len(), k, runtime.GOMAXPROCS(-1), false)
		combinSeq := lib.SplitCombin(allowedParent.Len(), k, 1, false)

		var seqCombins [][]int
		var parCombins [][]int

		for _, combin := range combinSeq {

			for combin.HasNext() {
				temp2 := combin.GetNext()
				temp := make([]int, len(temp2), len(temp2))

				copy(temp, temp2)

				seqCombins = append(seqCombins, temp)

				combin.Confirm()
			}
		}

		for _, combin := range combinParallel {
			for combin.HasNext() {
				temp2 := combin.GetNext()
				temp := make([]int, len(temp2), len(temp2))

				copy(temp, temp2)

				parCombins = append(parCombins, temp)

				combin.Confirm()
			}
		}

	Outer:
		for i := range seqCombins {
			for j := range parCombins {
				if reflect.DeepEqual(seqCombins[i], parCombins[j]) {
					continue Outer
				}
			}

			fmt.Println("Len: ", allowedParent.Len())
			fmt.Println("K,", k)
			fmt.Println("Split: ", runtime.GOMAXPROCS(-1))

			fmt.Println("CombinIterator: ", combinParallel[0])

			fmt.Print("\n All stuff in combinPar: ")
			for _, c := range parCombins {

				fmt.Print(c, ",")

			}

			fmt.Print("\n\n All stuff in combinSeq: ")
			for _, c := range seqCombins {

				fmt.Print(c, ",")

			}

			t.Error("SplitCombins don't match!")

			break Outer

		}

	}

}
