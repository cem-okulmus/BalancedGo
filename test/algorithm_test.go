package tests

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	algo "github.com/cem-okulmus/BalancedGo/algorithms"
	"github.com/cem-okulmus/BalancedGo/lib"
)

func solve(solver algo.Algorithm, graph lib.Graph, hinget *lib.Hingetree) lib.Decomp {
	var output lib.Decomp

	if hinget != nil {
		output = hinget.DecompHinge(solver, graph)
	} else {
		output = solver.FindDecomp()
	}

	return output
}

type solverDecomp struct {
	solver algo.Algorithm
	decomp lib.Decomp
}

func logActive(b bool) {
	if b {
		log.SetOutput(os.Stderr)

		log.SetFlags(0)
	} else {

		log.SetOutput(ioutil.Discard)
	}
}

func TestAlgo(t *testing.T) {

	// logActive(false)

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	graphInitial, encoding := getRandomGraph(10)
	// graph := lib.GetGraphPACE(graphFirst.ToPACE())
	var graph lib.Graph
	graph = graphInitial
	width := r.Intn(6) + 1
	BalFactor := 2

	heuristicRand := r.Intn(4)

	// Preprocessing on graph
	order1 := lib.GetDegreeOrder(graph.Edges)
	order2 := lib.GetMaxSepOrder(graph.Edges)
	order3 := lib.GetMSCOrder(graph.Edges)
	order4 := lib.GetEdgeDegreeOrder(graph.Edges)

	switch heuristicRand {
	case 1:
		graph.Edges = order1
		break
	case 2:
		graph.Edges = order2
		break
	case 3:
		graph.Edges = order3
		break
	case 4:
		graph.Edges = order4
		break
	}

	var removalMap map[int][]int
	var ops []lib.GYÖReduct
	graph, removalMap, _ = graph.TypeCollapse()
	graph, ops = graph.GYÖReduct()

	hinget := lib.GetHingeTree(graph)

	var algoTestsHD []algo.Algorithm
	var algoTestsGHD []algo.Algorithm

	balDet := &algo.BalSepHybrid{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
		Depth:     1,
	}
	algoTestsGHD = append(algoTestsGHD, balDet)

	seqBalDet := &algo.BalSepHybridSeq{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
		Depth:     1,
	}

	algoTestsGHD = append(algoTestsGHD, seqBalDet)

	det := &algo.DetKDecomp{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
		SubEdge:   false,
	}

	algoTestsHD = append(algoTestsHD, det)

	localBIP := &algo.DetKDecomp{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
		SubEdge:   true,
	}

	algoTestsGHD = append(algoTestsGHD, localBIP)

	logK := &algo.LogKDecomp{
		Graph:     graph,
		K:         width,
		BalFactor: BalFactor,
	}

	algoTestsHD = append(algoTestsHD, logK)

	meta := r.Intn(300)
	logkHybridChoice := r.Intn(4)

	logKHyb := &algo.LogKHybrid{
		Graph:     graph,
		K:         width,
		BalFactor: BalFactor,
	}
	logKHyb.Size = meta

	var pred algo.HybridPredicate

	switch logkHybridChoice {
	case 0:
		pred = logKHyb.NumberEdgesPred
	case 1:
		pred = logKHyb.SumEdgesPred
	case 2:
		pred = logKHyb.ETimesKDivAvgEdgePred
	case 3:
		pred = logKHyb.OneRoundPred

	}
	logKHyb.Predicate = pred // set the predicate to use

	algoTestsHD = append(algoTestsHD, logKHyb)

	global := &algo.BalSepGlobal{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
	}

	algoTestsGHD = append(algoTestsGHD, global)

	local := &algo.BalSepLocal{
		K:         width,
		Graph:     graph,
		BalFactor: BalFactor,
	}

	algoTestsGHD = append(algoTestsGHD, local)

	// test out all algorithms

	first := true
	prevAnswer := false
	prevAlgo := ""
	var out lib.Decomp
	prevDecomp := lib.Decomp{}

	for _, algorithm := range algoTestsHD {

		algorithm.SetGenerator(lib.ParallelSearchGen{})

		out = solve(algorithm, graph, &hinget)

		if !reflect.DeepEqual(out, lib.Decomp{}) || (len(ops) > 0 && graph.Edges.Len() == 0) {
			var result bool
			out.Root, result = out.Root.RestoreGYÖ(ops)
			if !result {
				fmt.Println("Partial decomp:", out.Root)
				log.Panicln("GYÖ reduction failed")
			}
			out.Root, result = out.Root.RestoreTypes(removalMap)
			if !result {
				fmt.Println("Partial decomp:", out.Root)
				log.Panicln("Type Collapse reduction failed")
			}
		}
		if !reflect.DeepEqual(out, lib.Decomp{}) {
			out.Graph = graphInitial
		}

		answer := out.Correct(graphInitial)
		if answer {
			if out.CheckWidth() > width {
				t.Errorf("Out decomp of higher width than required: %v, width  %v", out, width)
			}
		}

		if !first && answer != prevAnswer {

			fmt.Println("GraphInitial ", graphInitial.ToHyberBenchFormat())
			fmt.Println("Graph ", graph)
			fmt.Println("Width: ", width)

			fmt.Println("Current algo ", algorithm.Name(), "answer: ", answer, " and decomp: ", out)
			fmt.Println("Current algo ", prevAlgo, "answer: ", prevAnswer, " and decomp: ", prevDecomp)

			t.Errorf("Found disagreement among the algorithms: %v %v", algorithm.Name(), prevAlgo)
		}

		prevAnswer = answer
		prevAlgo = algorithm.Name()
		prevDecomp = out
		first = false
	}

	first = true
	prevAnswer = false
	prevAlgo = ""
	prevDecomp = lib.Decomp{}

	for _, algorithm := range algoTestsGHD {

		algorithm.SetGenerator(lib.ParallelSearchGen{})

		out = solve(algorithm, graph, &hinget)

		if !reflect.DeepEqual(out, lib.Decomp{}) || (len(ops) > 0 && graph.Edges.Len() == 0) {
			var result bool
			out.Root, result = out.Root.RestoreGYÖ(ops)
			if !result {
				fmt.Println("Partial decomp:", out.Root)
				log.Panicln("GYÖ reduction failed")
			}
			out.Root, result = out.Root.RestoreTypes(removalMap)
			if !result {
				fmt.Println("Partial decomp:", out.Root)
				log.Panicln("Type Collapse reduction failed")
			}
		}
		if !reflect.DeepEqual(out, lib.Decomp{}) {
			out.Graph = graphInitial
		}

		answer := out.Correct(graphInitial)

		if answer {
			if out.CheckWidth() > width {
				t.Errorf("Out decomp of higher width than required: %v, width  %v", out, width)
			}
		}

		if !first && answer != prevAnswer {

			fmt.Println("GraphInitial ", graphInitial.ToHyberBenchFormat())
			fmt.Println("Graph ", graph)
			fmt.Println("Width: ", width)

			fmt.Println("Current algo ", algorithm.Name(), "answer: ", answer, " and decomp: ", out)
			fmt.Println("Current algo ", prevAlgo, "answer: ", prevAnswer, " and decomp: ", prevDecomp)

			t.Errorf("Found disagreement among the algorithms: %v %v", algorithm.Name(), prevAlgo)
		}

		prevAnswer = answer
		prevAlgo = algorithm.Name()
		prevDecomp = out
		first = false
	}

	//produce a GML
	gml := out.ToGML()
	if !reflect.DeepEqual(out, lib.Decomp{}) {
		lib.GetDecompGML(gml, graphInitial, encoding)
	}

}
