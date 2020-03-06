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

var Version string
var Date string
var Build string

func outputStanza(algorithm string, decomp Decomp, msec float64, parsedGraph Graph, gml string, K int, heuristic float64) {
	decomp.RestoreSubedges()

	fmt.Println("Used algorithm: " + algorithm + " @" + Version)
	fmt.Println("Result ( ran with K =", K, ")\n", decomp)
	if heuristic > 0.0 {
		fmt.Print("Time: ", msec+heuristic, " ms ( decomp:", msec, ", heuristic:", heuristic, ")")
		if msec+heuristic > 60000.0 {
			fmt.Print("(", (msec+heuristic)/60000.0, "min )")
		}
	} else {
		fmt.Print("Time: ", msec, " ms")
		if msec > 60000.0 {
			fmt.Print("(", msec/60000.0, "min )")
		}
	}

	fmt.Println("\nWidth: ", decomp.CheckWidth())
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

func main() {

	//Command-Line Argument Parsing
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	logging := flag.Bool("log", false, "turn on extensive logs")
	computeSubedges := flag.Bool("sub", false, "turn off subedge computation for global option")
	width := flag.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	graphPath := flag.String("graph", "", "the file path to a hypergraph \n\t(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)")
	// choose := flag.Int("choice", 0, "only run one version\n\t1 ... Full Parallelism\n\t2 ... Search Parallelism\n\t3 ... Comp. Parallelism\n\t4 ... Sequential execution\n\t5 ... Local Full Parallelism\n\t6 ... Local Search Parallelism\n\t7 ... Local Comp. Parallelism\n\t8 ... Local Sequential execution.")
	localBal := flag.Bool("local", false, "Use local BalSep algorithm")
	globalBal := flag.Bool("global", false, "Use global BalSep algorithm")
	balanceFactorFlag := flag.Int("balfactor", 2, "Changes the factor that balanced separator check uses, default 2")
	useHeuristic := flag.Int("heuristic", 0, "turn on to activate edge ordering\n\t1 ... Degree Ordering\n\t2 ... Max. Separator Ordering\n\t3 ... MCSO\n\t4 ... Edge Degree Ordering")
	gyö := flag.Bool("g", false, "perform a GYÖ reduct")
	typeC := flag.Bool("t", false, "perform a Type Collapse")
	//hingeFlag := flag.Bool("hinge", false, "use isHinge Optimization")
	numCPUs := flag.Int("cpu", -1, "Set number of CPUs to use")
	bench := flag.Bool("bench", false, "Benchmark mode, reduces unneeded output (incompatible with -log flag)")
	akatovTest := flag.Bool("akatov", false, "Use Balanced Decomposition algorithm")
	detKTest := flag.Bool("det", false, "Use DetKDecomp algorithm")
	localBIP := flag.Bool("localbip", false, "To be used in combination with \"det\": turns on local subedge handling")
	divideTest := flag.Bool("divide", false, "Use divideKDecomp algoritm")
	divideParTest := flag.Bool("dividePar", false, "Use parallel divideKDecomp algorithm")
	balDetTest := flag.Int("balDet", 0, "Use the Hybrid BalSep - DetK algorithm. Number indicates depth, must be ≥ 1")
	gml := flag.String("gml", "", "Output the produced decomposition into the specified gml file ")
	pace := flag.Bool("pace", false, "Use PACE 2019 format for graphs\n\t(see https://pacechallenge.org/2019/htd/htd_format/ for correct format)")
	update := flag.Bool("update", false, "Use adapted PACE format, and call algorithm with initial special Edges")
	exact := flag.Bool("exact", false, "Compute exact width (width flag ignored)")

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
		out := fmt.Sprint("Usage of BalancedGo (", Version, ", https://github.com/cem-okulmus/BalancedGo/commit/", Build, " ,", Date, ")")
		fmt.Fprintln(os.Stderr, out)
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
	var currentGraph Graph
	var specialEdges []Special

	if *update {
		parsedGraph, currentGraph, specialEdges = GetGraphUpdate(string(dat))

	} else if !*pace {
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
			break
		case 2:
			parsedGraph.Edges = GetMaxSepOrder(parsedGraph.Edges)
			heuristicMessage = "Using max seperator ordering as a heuristic"
			break
		case 3:
			parsedGraph.Edges = GetMSCOrder(parsedGraph.Edges)
			heuristicMessage = "Using MSC ordering as a heuristic"
			break
		case 4:
			parsedGraph.Edges = GetEdgeDegreeOrder(parsedGraph.Edges)
			heuristicMessage = "Using edge degree ordering as a heuristic"
			break
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
	var removalMap map[int][]int
	// Performing Type Collapse
	if *typeC {
		count := 0
		reducedGraph, removalMap, count = parsedGraph.TypeCollapse()
		parsedGraph = reducedGraph
		if !*bench { // be silent when benchmarking
			fmt.Println("\n\n", *graphPath)
			fmt.Println("Graph after Type Collapse:")
			for _, e := range reducedGraph.Edges.Slice() {
				fmt.Printf("%v %v\n", e, Edge{Vertices: e.Vertices})
			}
			fmt.Println("Removed ", count, " vertex/vertices\n")
		}
	}

	var ops []GYÖReduct
	// Performing GYÖ reduction
	if *gyö {

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
			fmt.Println(ops, "\n")
		}

	}

	// Add all subdedges to graph
	if *globalBal && !*computeSubedges {
		parsedGraph = parsedGraph.ComputeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	if *update {
		var solver UpdateAlgorithm

		if *balDetTest > 0 {
			balDet := BalDetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, Depth: *balDetTest - 1}
			solver = balDet
		} else if *detKTest {
			det := DetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, SubEdge: *localBIP}
			solver = det
		} else {
			log.Panicln("No update Algorithm chosen!")
		}

		if solver != nil {
			var decomp Decomp
			start := time.Now()

			decomp = solver.FindDecompUpdate(*width, currentGraph, specialEdges)

			d := time.Now().Sub(start)
			msec := d.Seconds() * float64(time.Second/time.Millisecond)

			if decomp.Correct(parsedGraph) {
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
			}

			outputStanza(solver.Name(), decomp, msec, parsedGraph, *gml, *width, heuristic)
			return
		}

		fmt.Println("No algorithm or procedure selected.")
		return

	}

	var solver Algorithm

	// Check for multiple flags
	chosen := 0

	if *akatovTest {
		bal := BalKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = bal
		chosen++
	}

	if *balDetTest > 0 {
		balDet := BalDetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, Depth: *balDetTest - 1}
		solver = balDet
		chosen++
	}

	if *detKTest {
		det := DetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, SubEdge: *localBIP}
		solver = det
		chosen++
	}

	if *divideTest {
		div := DivideKDecomp{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
		solver = div
		chosen++
	}

	if *divideParTest {
		div := DivideKDecompPar{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
		solver = div
		chosen++
	}

	if *globalBal {
		global := BalSepGlobal{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = global
		chosen++
	}

	if *localBal {
		local := BalSepLocal{Graph: parsedGraph, BalFactor: BalancedFactor}
		solver = local
		chosen++
	}

	if chosen > 1 {
		fmt.Println("Only one algorithm may be chosen at a time. Make up your mind.")
		return
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

		if decomp.Correct(parsedGraph) {
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
		}

		outputStanza(solver.Name(), decomp, msec, parsedGraph, *gml, *width, heuristic)
		return
	}

	fmt.Println("No algorithm or procedure selected.")

}
