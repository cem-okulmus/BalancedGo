package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

func pruneTree(n *Node) {

	var nuChildren []Node

	for i, c := range n.Children {

		if i > len(n.Children) {
			fmt.Println("Current node", n)
			fmt.Println("Children: ", n.Children)

			fmt.Println("Index: ", c)

			log.Panicln("the fuck?")
		}
		pruneTree(&c)
		if Subset(c.Bag, n.Bag) {
			nuChildren = append(nuChildren, c.Children...)

		} else {
			nuChildren = append(nuChildren, c)
		}
	}

	n.Children = nuChildren

}

func main() {
	graphPath := flag.String("graph", "", "the file path to a hypergraph in HyperBench Format")
	decompPath := flag.String("decomp", "", "the file path to a decomposition in GML format")
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

	parsedGraph, parseGraph := GetGraph(string(dat))

	var reducedGraph Graph

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

		decomp := GetDecompGML(string(dis), parsedGraph, parseGraph.Encoding)

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

		var ops []GYÖReduct
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
