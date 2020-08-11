package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"

	. "github.com/cem-okulmus/BalancedGo/algorithms"

	. "github.com/cem-okulmus/BalancedGo/lib"
)

func logActive(b bool) {
	if b {
		log.SetOutput(os.Stderr)

		log.SetFlags(0)
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

type labelTime struct {
	time  float64
	label string
}

func (l labelTime) String() string {
	return fmt.Sprintf("%s : %.5f ms", l.label, l.time)
}

func outputStanza(algorithm string, decomp Decomp, times []labelTime, parsedGraph Graph, gml string, K int, skipCheck bool) {
	decomp.RestoreSubedges()

	fmt.Println("Used algorithm: " + algorithm + " @" + Version)
	fmt.Println("Result ( ran with K =", K, ")\n", decomp)

	// Print the times
	var sumTotal float64

	for _, time := range times {
		sumTotal = sumTotal + time.time
	}
	fmt.Printf("Time: %.5f ms\n", sumTotal)

	fmt.Println("Time Composition: ")
	for _, time := range times {
		fmt.Println(time)
	}

	fmt.Println("\nWidth: ", decomp.CheckWidth())
	var correct bool
	if !skipCheck {
		correct = decomp.Correct(parsedGraph)
	} else {
		correct = true
	}

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

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)

	//Command-Line Argument Parsing
	cpuprofile := flagSet.String("cpuprofile", "", "write cpu profile to file")
	logging := flagSet.Bool("log", false, "turn on extensive logs")
	computeSubedges := flagSet.Bool("sub", false, "turn off subedge computation for global option")
	width := flagSet.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	graphPath := flagSet.String("graph", "", "the file path to a hypergraph \n\t(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)")
	// choose := flagSet.Int("choice", 0, "only run one version\n\t1 ... Full Parallelism\n\t2 ... Search Parallelism\n\t3 ... Comp. Parallelism\n\t4 ... Sequential execution\n\t5 ... Local Full Parallelism\n\t6 ... Local Search Parallelism\n\t7 ... Local Comp. Parallelism\n\t8 ... Local Sequential execution.")
	localBal := flagSet.Bool("local", false, "Use local BalSep algorithm")
	globalBal := flagSet.Bool("global", false, "Use global BalSep algorithm")
	balanceFactorFlag := flagSet.Int("balfactor", 2, "Changes the factor that balanced separator check uses, default 2")
	useHeuristic := flagSet.Int("heuristic", 0, "turn on to activate edge ordering\n\t1 ... Degree Ordering\n\t2 ... Max. Separator Ordering\n\t3 ... MCSO\n\t4 ... Edge Degree Ordering")
	gyö := flagSet.Bool("g", false, "perform a GYÖ reduct")
	typeC := flagSet.Bool("t", false, "perform a Type Collapse")
	hingeFlag := flagSet.Bool("h", false, "use hingeTree Optimization")
	numCPUs := flagSet.Int("cpu", -1, "Set number of CPUs to use")
	bench := flagSet.Bool("bench", false, "Benchmark mode, reduces unneeded output (incompatible with -log flag)")
	akatovTest := flagSet.Bool("akatov", false, "Use Balanced Decomposition algorithm")
	logDetKAlgo := flagSet.Bool("logdetk", false, "Use log-depth version of DetKDecomp algorithm")
	logK := flagSet.Bool("logk", false, "Use LogKDecomp algoritm")
	detKTest := flagSet.Bool("det", false, "Use DetKDecomp algorithm")
	localBIP := flagSet.Bool("localbip", false, "To be used in combination with \"det\": turns on local subedge handling")
	// divideTest := flagSet.Bool("divide", false, "Use divideKDecomp algoritm")
	// divideParTest := flagSet.Bool("dividePar", false, "Use parallel divideKDecomp algorithm")
	balDetTest := flagSet.Int("balDet", 0, "Use the Hybrid BalSep - DetK algorithm. Number indicates depth, must be ≥ 1")
	gml := flagSet.String("gml", "", "Output the produced decomposition into the specified gml file ")
	pace := flagSet.Bool("pace", false, "Use PACE 2019 format for graphs\n\t(see https://pacechallenge.org/2019/htd/htd_format/ for correct format)")
	// update := flagSet.Bool("update", false, "Use adapted PACE format, and call algorithm with initial special Edges")
	exact := flagSet.Bool("exact", false, "Compute exact width (width flag ignored)")
	approx := flagSet.Int("approx", 0, "Compute approximated width and set a timeout in seconds (width flag ignored)")
	decomp := flagSet.String("decomp", "", "A decomposition to be used as a starting point, needs to have certain nodes marked (those which need to be updated).")

	parseError := flagSet.Parse(os.Args[1:])
	if parseError != nil {
		fmt.Print("Parse Error:\n", parseError.Error(), "\n\n")
	}

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
	if parseError != nil || *graphPath == "" || (*width <= 0 && !*exact && *approx == 0) {
		out := fmt.Sprint("Usage of BalancedGo (", Version, ", https://github.com/cem-okulmus/BalancedGo/commit/", Build, ", ", Date, ")")
		fmt.Fprintln(os.Stderr, out)
		flagSet.VisitAll(func(f *flag.Flag) {
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
		flagSet.VisitAll(func(f *flag.Flag) {
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
	var parseGraph ParseGraph

	if !*pace {
		parsedGraph, parseGraph = GetGraph(string(dat))
	} else {
		parsedGraph = GetGraphPACE(string(dat))
	}
	log.Println("BIP: ", parsedGraph.GetBIP())
	var reducedGraph Graph

	var times []labelTime

	var solverUpdate UpdateAlgorithm
	var parsedDecomp Decomp

	// Check if shortcut present, before applying heuristics
	if *decomp != "" {

		// Determine solver

		if *detKTest {
			det := DetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, SubEdge: *localBIP}
			solverUpdate = det
		}

		if *balDetTest > 0 {
			balDet := BalDetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, Depth: *balDetTest - 1}
			solverUpdate = balDet
		}

		// read and parse decomposition

		dis, err2 := ioutil.ReadFile(*decomp)
		check(err2)

		start_pars := time.Now()

		parsedDecomp = GetDecomp(dis, parsedGraph, parseGraph.Encoding)
		// fmt.Println("parsed Decomp", parsedDecomp)

		msec_pars := time.Now().Sub(start_pars).Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec_pars, label: "Parsing"})

		start_check := time.Now()

		// Check if this decomp is already correct
		check := parsedDecomp.Correct(parsedGraph)

		msec_check := time.Now().Sub(start_check).Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec_check, label: "Correctness Check"})

		if check {
			fmt.Println(" Parsed Decomposition already correct, skipping update computation.")

			outputStanza(solverUpdate.Name(), parsedDecomp, times, parsedGraph, *gml, *width, true)

			return
		}

	}

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
		msec := d.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec, label: "Heuristic"})

		if !*bench {
			fmt.Println(heuristicMessage)
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
			fmt.Print("Removed ", count, " vertex/vertices\n\n")
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
			fmt.Print(ops, "\n\n")
		}

	}

	// Add all subdedges to graph
	if *globalBal && !*computeSubedges {
		parsedGraph = parsedGraph.ComputeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	var hinget Hingetree
	var msecHinge float64

	if *hingeFlag {
		startHinge := time.Now()

		hinget = GetHingeTree(parsedGraph)

		dHinge := time.Now().Sub(startHinge)
		msecHinge = dHinge.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msecHinge, label: "Hingetree"})

		if !*bench {
			fmt.Println("Produced Hingetree: ")
			fmt.Println(hinget)
		}
	}

	if *decomp != "" {

		start_sc := time.Now()

		scenes := parsedDecomp.SceneCreation(parsedGraph)

		d_sc := time.Now().Sub(start_sc)
		msec_sc := d_sc.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec_sc, label: "Scene Creation"})

		fmt.Println("Extracted scenes: ", len(scenes))

		if solverUpdate != nil {
			var decomp Decomp
			start := time.Now()

			decomp = solverUpdate.FindDecompUpdate(*width, parsedGraph, scenes)

			d := time.Now().Sub(start)
			msec := d.Seconds() * float64(time.Second/time.Millisecond)
			times = append(times, labelTime{time: msec, label: "Decomposition"})

			if !reflect.DeepEqual(decomp, Decomp{}) {
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

			outputStanza(solverUpdate.Name(), decomp, times, parsedGraph, *gml, *width, false)
			return
		}

		fmt.Println("No supported update Algorithm chosen.")
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

	if *logK {
		logK := LogKDecomp{Graph: parsedGraph, K: *width}
		solver = logK
		chosen++
	}

	if *logDetKAlgo {
		logDetk := LogDetKDecomp{Graph: parsedGraph, SubEdge: *localBIP}
		solver = logDetk
		chosen++
	}

	// if *divideTest {
	// 	div := DivideKDecomp{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
	// 	solver = div
	// 	chosen++
	// }

	// if *divideParTest {
	// 	div := DivideKDecompPar{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
	// 	solver = div
	// 	chosen++
	// }

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
				if *hingeFlag {
					decomp = hinget.DecompHinge(solver, k, parsedGraph)
				} else {
					decomp = solver.FindDecomp(k)
				}

				solved = decomp.Correct(parsedGraph)
				k++
			}
			*width = k - 1 // for correct output
		} else if *approx > 0 {
			ch := make(chan int, 1)
			go func() {
				m := parsedGraph.Edges.Len()
				k := int(math.Ceil(float64(m) / 2))
				decomp = solver.FindDecomp(k)
				k = decomp.CheckWidth()
				solved := false

				var newDecomp Decomp
				for !solved {
					newK := k - 1
					if *hingeFlag {
						newDecomp = hinget.DecompHinge(solver, newK, parsedGraph)
					} else {
						newDecomp = solver.FindDecomp(newK)
					}
					if newDecomp.Correct(parsedGraph) {
						k = newDecomp.CheckWidth()
						decomp = newDecomp
					} else {
						solved = true
					}
				}
				ch <- k
			}()

			select {
			case res := <-ch:
				*width = res
			case <-time.After(time.Duration(*approx) * time.Second):
				*width = decomp.CheckWidth()
			}
		} else {
			if *hingeFlag {
				decomp = hinget.DecompHinge(solver, *width, parsedGraph)
			} else {
				decomp = solver.FindDecomp(*width)
			}
		}

		if *akatovTest {
			decomp.Blowup()
		}

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec, label: "Decomposition"})

		if !reflect.DeepEqual(decomp, Decomp{}) {
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

		outputStanza(solver.Name(), decomp, times, parsedGraph, *gml, *width, false)
		return
	}

	fmt.Println("No algorithm or procedure selected.")

}
