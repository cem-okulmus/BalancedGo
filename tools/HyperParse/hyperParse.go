package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// mem checks if an integer b occurs inside a slice as
func mem(as []int, b int) bool {
	for _, a := range as {
		if a == b {
			return true
		}
	}
	return false
}

func pruneTree(n *lib.Node) {

	var nuChildren []lib.Node

	for i, c := range n.Children {

		if i > len(n.Children) {
			fmt.Println("Current node", n)
			fmt.Println("Children: ", n.Children)

			fmt.Println("Index: ", c)

			log.Panicln("the fuck?")
		}
		pruneTree(&c)
		if lib.Subset(c.Bag, n.Bag) {
			nuChildren = append(nuChildren, c.Children...)

		} else {
			nuChildren = append(nuChildren, c)
		}
	}

	n.Children = nuChildren

}

func getDegree(edges lib.Edges, e lib.Edge) int {
	output := 0

	for _, node := range e.Vertices {
		degree := 0
		for _, e2 := range edges.Slice() {
			if mem(e2.Vertices, node) {
				degree++
			}
		}
		if degree > output {
			output = degree
		}
	}

	return output
}

func main() {
	graphPath := flag.String("graph", "", "the file path to a hypergraph in HyperBench Format")
	decompPath := flag.String("decomp", "", "the file path to a decomposition in GML format")
	outPath := flag.String("out", "", "the path for outputting the graph in PACE 2019 format")
	statFlag := flag.Bool("stats", false, "output stats of the hypergraph")
	starRedlag := flag.Bool("statsReduced", false, "output stats of the reduced hypergraph")

	flag.Parse()

	if *graphPath == "" {
		flag.Usage()
		return
	}

	dat, err := ioutil.ReadFile(*graphPath)
	if err != nil {
		panic(err)
	}

	parsedGraph, parseGraph := lib.GetGraph(string(dat))

	if *statFlag {
		sumEdgeSizes := 0
		maxEdge := 0
		maxDegree := 0

		for _, e := range parsedGraph.Edges.Slice() {
			sumEdgeSizes = sumEdgeSizes + len(e.Vertices)

			if len(e.Vertices) > maxEdge {
				maxEdge = len((e.Vertices))
			}
			degree := getDegree(parsedGraph.Edges, e)
			if degree > maxDegree {
				maxDegree = degree
			}
		}

		average := float32(sumEdgeSizes) / float32(len(parsedGraph.Edges.Slice()))

		fmt.Print(average, ",", len(parsedGraph.Vertices()), ",", len(parsedGraph.Edges.Slice()), ",", maxDegree, ",", maxEdge, ",", sumEdgeSizes)

		os.Exit(0)
	}

	var reducedGraph lib.Graph

	if *starRedlag {

		// Performing Type Collapse
		reducedGraph, _, _ = parsedGraph.TypeCollapse()

		// Performing GYÖ reduction
		reducedGraph, _ = reducedGraph.GYÖReduct()

		hinget := lib.GetHingeTree(reducedGraph)

		reducedGraph = hinget.GetLargestGraph()

		sumEdgeSizes := 0
		maxEdge := 0
		maxDegree := 0

		for _, e := range reducedGraph.Edges.Slice() {
			sumEdgeSizes = sumEdgeSizes + len(e.Vertices)

			if len(e.Vertices) > maxEdge {
				maxEdge = len((e.Vertices))
			}
			degree := getDegree(reducedGraph.Edges, e)
			if degree > maxDegree {
				maxDegree = degree
			}
		}

		average := float32(sumEdgeSizes) / float32(len(reducedGraph.Edges.Slice()))

		fmt.Print(average, ",", len(reducedGraph.Vertices()), ",", len(reducedGraph.Edges.Slice()),
			",", maxDegree, ",", maxEdge, ",", sumEdgeSizes)

		os.Exit(0)
	}

	if *decompPath != "" {

		dis, err2 := ioutil.ReadFile(*decompPath)
		if err2 != nil {
			panic(err)
		}

		f, err := os.Create(*outPath)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		decomp := lib.GetDecompGML(string(dis), parsedGraph, parseGraph.Encoding)

		if decomp.Correct(parsedGraph) {

			pruneTree(&decomp.Root)

			if !decomp.Correct(parsedGraph) {
				log.Panicln("pruning broke the decomp")
			}

			f.WriteString(decomp.ToGML())
			f.Sync()

			os.Exit(0)
		}

		// fmt.Println("parsed Decomp ", decomp)

		var removalMap map[int][]int
		// Performing Type Collapse
		reducedGraph, removalMap, _ = parsedGraph.TypeCollapse()
		parsedGraph = reducedGraph

		var ops []lib.GYÖReduct
		// Performing GYÖ reduction

		reducedGraph, ops = reducedGraph.GYÖReduct()

		decomp.Graph = reducedGraph

		// parsedGraph = reducedGraph
		// fmt.Println("Graph after GYÖ:")
		// fmt.Println(reducedGraph)
		// fmt.Println("Reductions:")
		// fmt.Print(ops, "\n\n")

		if !decomp.Correct(reducedGraph) {
			log.Panicln("decomp isn't correct decomp of reduced graph")
		}

		var result bool
		decomp.Root, result = decomp.Root.RestoreGYÖ(ops)
		if !result {
			fmt.Println("Partial decomp:", decomp.Root)

			log.Panicln("GYÖ reduction failed")
		}

		decomp.Root, result = decomp.Root.RestoreTypes(removalMap)
		if !result {
			fmt.Println("Partial decomp:", decomp.Root)

			log.Panicln("Type Collapse reduction failed")
		}

		decomp.Graph = parsedGraph

		var correct bool
		correct = decomp.Correct(parsedGraph)

		if !correct {
			log.Panicln("Decomp after GYO restoration not correct")
		}

		f.WriteString(decomp.ToGML())
		f.Sync()

		os.Exit(0)

	}

	f, err := os.Create(*outPath)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.WriteString(parsedGraph.ToPACE())
	f.Sync()
}
