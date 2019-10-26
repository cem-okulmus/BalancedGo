package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	. "github.com/cem-okulmus/BalancedGo/algorithms"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

func logActive(b bool) {
	log.SetFlags(0)
	if b {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}

}

func outputStanza(algorithm string, decomp Decomp, msec float64, parsedGraph Graph, gml string, K int, heuristic float64) {
	decomp.RestoreSubedges()

	fmt.Println("Used algorithm: " + algorithm)
	fmt.Println("Result ( ran with K =", K, ")\n", decomp)
	if heuristic > 0.0 {
		fmt.Print("Time: ", msec+heuristic, " ms ( decomp:", msec, ", heuristic:", heuristic, ")")
		if msec+heuristic > 60000.0 {
			fmt.Println("(", msec+heuristic/60000.0, "min )")
		} else {
			fmt.Println("")
		}
	} else {
		fmt.Print("Time: ", msec, " ms")
		if msec > 60000.0 {
			fmt.Println("(", msec/60000.0, "min )")
		} else {
			fmt.Println("")
		}
	}

	fmt.Println("Width: ", decomp.CheckWidth())
	correct := decomp.Correct(parsedGraph)
	fmt.Println("Correct: ", correct)
	if correct && len(gml) > 0 {
		f, err := os.Create(gml)
		check(err)

		defer f.Close()
		f.WriteString(decomp.ToGML())
		f.Sync()

	}
}

var Version string
var Build string

func main() {

	// m = make(map[int]string)

	//Command-Line Argument Parsing
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	logging := flag.Bool("log", false, "turn on extensive logs")
	computeSubedges := flag.Bool("sub", false, "Compute the subedges of the graph and print it out")
	width := flag.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	graphPath := flag.String("graph", "", "the file path to a hypergraph \n\t(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)")
	// choose := flag.Int("choice", 0, "only run one version\n\t1 ... Full Parallelism\n\t2 ... Search Parallelism\n\t3 ... Comp. Parallelism\n\t4 ... Sequential execution\n\t5 ... Local Full Parallelism\n\t6 ... Local Search Parallelism\n\t7 ... Local Comp. Parallelism\n\t8 ... Local Sequential execution.")
	localBal := flag.Bool("local", false, "Test out local BalSep")
	globalBal := flag.Bool("global", false, "Test out global BalSep")
	balanceFactorFlag := flag.Int("balfactor", 2, "Determines the factor that balanced separator check uses")
	useHeuristic := flag.Int("heuristic", 0, "turn on to activate edge ordering\n\t1 ... Degree Ordering\n\t2 ... Max. Separator Ordering\n\t3 ... MCSO")
	gyö := flag.Bool("g", false, "perform a GYÖ reduct and show the resulting graph")
	typeC := flag.Bool("t", false, "perform a Type Collapse and show the resulting graph")
	//hingeFlag := flag.Bool("hinge", false, "use isHinge Optimization")
	numCPUs := flag.Int("cpu", -1, "Set number of CPUs to use")
	bench := flag.Bool("bench", false, "Benchmark mode, reduces unneeded output (incompatible with -log flag)")
	akatovTest := flag.Bool("akatov", false, "compute balanced decomposition")
	detKTest := flag.Bool("det", false, "Test out DetKDecomp")
	localBIP := flag.Bool("localbip", false, "To be used in combination with \"det\": turns on subedge handling")
	divideTest := flag.Bool("divide", false, "Test for divideKDecomp")
	balDetTest := flag.Int("balDet", 0, "Test hybrid balSep and DetK algorithm")
	gml := flag.String("gml", "", "Output the produced decomposition into the specified gml file ")
	pace := flag.Bool("pace", false, "Use PACE 2019 format for graphs\n\t(see https://pacechallenge.org/2019/htd/htd_format/ for correct format)")
	exact := flag.Bool("exact", false, "Compute exact width (width flag not needed)")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)

		defer pprof.StopCPUProfile()
	}

	if *bench { // no logging output when running benchmarks
		*logging = false
	}
	logActive(*logging)

	BalancedFactor := *balanceFactorFlag

	runtime.GOMAXPROCS(*numCPUs)

	// Outpt usage message if graph and width not specified
	if *graphPath == "" || (*width <= 0 && !*exact) {
		fmt.Fprintf(os.Stderr, "Usage of BalancedGo (v%s, https://github.com/cem-okulmus/BalancedGo/commit/%s): \n", Version, Build)
		flag.VisitAll(func(f *flag.Flag) {
			if f.Name != "width" && f.Name != "graph" && f.Name != "exact" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				fmt.Printf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				fmt.Printf("  -%-10s \n", f.Name)
			}
			fmt.Println("\t" + f.Usage)
		})

		fmt.Println("\nOptional Arguments: ")
		flag.VisitAll(func(f *flag.Flag) {
			if f.Name == "width" || f.Name == "graph" || f.Name == "exact" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				fmt.Printf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				fmt.Printf("  -%-10s \n", f.Name)
			}
			fmt.Println("\t" + f.Usage)
		})

		return
	}

	dat, err := ioutil.ReadFile(*graphPath)
	check(err)

	var parsedGraph Graph
	if !*pace {
		parsedGraph, _ = GetGraph(string(dat))
	} else {
		parsedGraph = GetGraphPACE(string(dat))
	}
	log.Println("BIP: ", parsedGraph.GetBIP())
	var reducedGraph Graph
	var heuristic float64

	// Sorting Edges to find separators faster
	if *useHeuristic > 0 {
		var heuristicMessage string

		start := time.Now()
		switch *useHeuristic {
		case 1:
			parsedGraph.Edges = GetDegreeOrder(parsedGraph.Edges)
			heuristicMessage = "Using degree ordering as a heuristic"
		case 2:
			parsedGraph.Edges = GetMaxSepOrder(parsedGraph.Edges)
			heuristicMessage = "Using max seperator ordering as a heuristic"
		case 3:
			parsedGraph.Edges = GetMSCOrder(parsedGraph.Edges)
			heuristicMessage = "Using MSC ordering as a heuristic"
		}
		d := time.Now().Sub(start)

		if !*bench {
			fmt.Println(heuristicMessage)
			msec := d.Seconds() * float64(time.Second/time.Millisecond)
			heuristic = msec
			fmt.Printf("Time for heuristic: %.5f ms\n", msec)
			fmt.Printf("Ordering: %v\n", parsedGraph.String())
		}
	}

	// Performing Type Collapse
	if *typeC {
		count := 0
		reducedGraph, _, count = parsedGraph.TypeCollapse()
		parsedGraph = reducedGraph
		if !*bench { // be silent when benchmarking
			fmt.Println("\n\n", *graphPath)
			fmt.Println("Graph after Type Collapse:")
			for _, e := range reducedGraph.Edges.Slice() {
				fmt.Printf("%v %v\n", e, Edge{Vertices: e.Vertices})
			}
			fmt.Println("Removed ", count, " vertex/vertices")
		}
	}

	// Performing GYÖ reduction
	if *gyö {

		var ops []GYÖReduct
		if *typeC {
			reducedGraph, ops = reducedGraph.GYÖReduct()
		} else {
			reducedGraph, ops = parsedGraph.GYÖReduct()
		}

		parsedGraph = reducedGraph
		if !*bench { // be silent when benchmarking
			fmt.Println("Graph after GYÖ:")
			fmt.Println(reducedGraph)
			fmt.Println("Reductions:")
			fmt.Println(ops)
		}

	}

	// Add all subdedges to graph
	if *computeSubedges {
		parsedGraph = parsedGraph.ComputeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	var solver Algorithm

	if *akatovTest {
		bal := BalKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = bal
	}

	if *balDetTest > 0 {
		balDet := BalDetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, Depth: *balDetTest - 1}
		solver = balDet
	}

	if *detKTest {
		det := DetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, SubEdge: *localBIP}
		solver = det
	}

	if *divideTest {
		div := DivideKDecomp{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
		solver = div
	}

	if *globalBal {
		global := BalSepGlobal{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = global
	}

	if *localBal {
		local := BalSepLocal{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = local
	}

	if solver != nil {
		var decomp Decomp
		start := time.Now()

		if *exact {
			k := 1
			solved := false
			for !solved {
				decomp = solver.FindDecomp(k)
				solved = decomp.Correct(parsedGraph)
				k++
			}
			*width = k - 1 // for correct output
		} else {
			decomp = solver.FindDecomp(*width)
		}

		if *akatovTest {
			decomp.Blowup()
		}

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		outputStanza(solver.Name(), decomp, msec, parsedGraph, *gml, *width, heuristic)
		return
	}

	//   SIMPLE BENCHMARKING SUITE FOR LOCAL BALSEP
	//  ============================================

	var output string
	local := BalSepLocal{Graph: parsedGraph, BalFactor: BalancedFactor}

	output = output + *graphPath + ";"

	output = output + fmt.Sprintf("%v;", parsedGraph.Edges.Len())
	output = output + fmt.Sprintf("%v;", len(parsedGraph.Edges.Vertices()))
	output = output + fmt.Sprintf("%v;", *width)

	// Parallel Execution FULL
	start := time.Now()
	decomp := local.FindGHDParallelFull(*width)

	d := time.Now().Sub(start)
	msec := d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)

	// Parallel Execution Search
	start = time.Now()
	decomp = local.FindGHDParallelSearch(*width)

	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)

	// Parallel Execution Comp
	start = time.Now()
	decomp = local.FindGHDParallelComp(*width)

	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	// Sequential Execution
	start = time.Now()
	decomp = local.FindGHD(*width)

	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v\n", decomp.Correct(parsedGraph))

	fmt.Print(output)
}
