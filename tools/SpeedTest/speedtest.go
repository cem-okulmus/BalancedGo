package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
)

func main() {
	graphPath := flag.String("graph", "", "the file path to a hypergraph in HyperBench Format")
	width := flag.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")

	flag.Parse()

	if *graphPath == "" || *width == 0 {
		flag.Usage()
		return
	}

	dat, err := ioutil.ReadFile(*graphPath)
	if err != nil {
		panic(err)
	}

	parsedGraph, _ := lib.GetGraph(string(dat))
	var startTime time.Time
	var timePassed time.Duration
	var counter int
	var solution []int

	go func() {
		select {
		case <-time.After(time.Duration(5) * time.Minute):
			timePassed := time.Now().Sub(startTime)

			fmt.Println(timePassed, " time has passed")
			fmt.Println(" Counters: ", counter)

			os.Exit(0)

		}

	}()

	check := lib.BalancedCheck{}

	gen := lib.SplitCombin(parsedGraph.Edges.Len(), *width, 1, true)[0]

	startTime = time.Now()

	for gen.HasNext() && len(solution) == 0 {
		counter++

		// j := make([]int, len(gen.Combination))
		// copy(gen.Combination, j)
		j := gen.GetNext()

		sep := lib.GetSubset(parsedGraph.Edges, j) // check new possible sep

		if check.Check(&parsedGraph, &sep, 2) {
			gen.Found() // cache result

			timePassed = time.Now().Sub(startTime)

			solution = make([]int, len(j))
			copy(solution, j)

		}
		gen.Confirm()
	}

	fmt.Println("Found seperator, with indices ", solution)
	fmt.Println("Time needed", timePassed.Seconds(), " sec")

}
