package main

import (
	"flag"
	"io/ioutil"
	"os"

	balgo "github.com/cem-okulmus/BalancedGo/lib"
)

func main() {
	graphPath := flag.String("graph", "", "the file path to a hypergraph in HyperBench Format")
	outPath := flag.String("out", "", "the path for outputting the graph in PACE 2019 format")

	flag.Parse()

	if *graphPath == "" {
		flag.Usage()
		return
	}

	dat, err := ioutil.ReadFile(*graphPath)
	if err != nil {
		panic(err)
	}

	parsedGraph, _ := balgo.GetGraph(string(dat))

	f, err := os.Create(*outPath)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.WriteString(parsedGraph.ToPACE())
	f.Sync()
}
