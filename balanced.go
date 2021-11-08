// BalancedGo - A research prototype to compute structural decompositions of Conjunctive Queries and CSPs
// via the use of Balanced Separators with a focus on parallelism using the programming language Go.
//
// For more detailed information,  see "Fast and Parallel Decomposition of Constraint Satisfaction Problems",
// Georg Gottlob, Cem Okulmus, Reinhard Pichler, released in Proc. IJCAI 2020.
// https://www.ijcai.org/Proceedings/2020/161
//
// The tool is split into three packages. main is responsible to actually run the various algorithms supported
// by the tool, lib is used to implement various functionality used by the algorithms and lastly algorithms which
// implements the actual algorithms to compute various decompositions.
//
// In addition to this, there is also a tool subdirectory in the repository which is intended to support functionality
// not directly related to the computation of decompositions, such as changing the formatting of hypergraphs, or fixing
// a faulty decomposition.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"syscall"
	"time"

	algo "github.com/cem-okulmus/BalancedGo/algorithms"
	"github.com/cem-okulmus/BalancedGo/lib"
)

// Decomp used to improve readability
type Decomp = lib.Decomp

// Edge used to improve readability
type Edge = lib.Edge

// Graph used to improve readability
type Graph = lib.Graph

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

//Version indicates the version exported from the Git repository
var Version string

//Date indicates the build date exported from the Git repository
var Date string

//Build indicates the exact build when current version was compiled
var Build string

type labelTime struct {
	time  float64
	label string
}

func (l labelTime) String() string {
	return fmt.Sprintf("%s : %.5f ms", l.label, l.time)
}

func outputStanza(algorithm string, decomp Decomp, times []labelTime, graph Graph, gml string, K int, skipCheck bool) {
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
		correct = decomp.Correct(graph)
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

	// ==============================================
	// Command-Line Argument Parsing

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)

	// input flags
	graphPath := flagSet.String("graph", "", "input (for format see hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf)")
	width := flagSet.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	exact := flagSet.Bool("exact", false, "Compute exact width (width flag ignored)")
	approx := flagSet.Int("approx", 0, "Compute approximated width and set a timeout in seconds (width flag ignored)")

	// algorithms  flags
	localBal := flagSet.Bool("local", false, "Use local BalSep algorithm")
	globalBal := flagSet.Bool("global", false, "Use global BalSep algorithm")
	logK := flagSet.Bool("logk", false, "Use LogKDecomp algorithm")
	logKHybrid := flagSet.Int("logkHybrid", 0, "Use DetK - LogK Hybrid algorithm. Choose which predicate to use")
	detKFlag := flagSet.Bool("det", false, "Use DetKDecomp algorithm")
	localBIP := flagSet.Bool("localbip", false, "Used in combination with \"det\": turns on local subedge handling")
	balDetFlag := flagSet.Int("balDet", 0, "Use the Hybrid BalSep-DetK algorithm. Number indicates depth, must be ≥ 1")
	seqBalDetFlag := flagSet.Int("seqBalDet", 0, "Use sequential Hybrid BalSep - DetK algorithm.")

	// heuristic flags
	heur := "1 ... Vertex Degree Ordering\n\t2 ... Max. Separator Ordering\n\t3 ... MCSO\n\t4 ... Edge Degree Ordering"
	useHeuristic := flagSet.Int("heuristic", 0, "turn on to activate edge ordering\n\t"+heur)
	gyö := flagSet.Bool("g", false, "perform a GYÖ reduct")
	typeC := flagSet.Bool("t", false, "perform a Type Collapse")
	hingeFlag := flagSet.Bool("h", false, "use hingeTree Optimization")

	//other optional  flags
	cpuprofile := flagSet.String("cpuprofile", "", "write cpu profile to file")
	logging := flagSet.Bool("log", false, "turn on extensive logs")
	computeSubedges := flagSet.Bool("sub", false, "turn off subedge computation for global option")
	balanceFactorFlag := flagSet.Int("balfactor", 2, "Changes the factor that balanced separator check uses, default 2")
	numCPUs := flagSet.Int("cpu", -1, "Set number of CPUs to use")
	bench := flagSet.Bool("bench", false, "Benchmark mode, reduces unneeded output (incompatible with -log flag)")
	gml := flagSet.String("gml", "", "Output the produced decomposition into the specified gml file ")
	pace := flagSet.Bool("pace", false, "Use PACE 2019 format for graphs (see pacechallenge.org/2019/htd/htd_format/)")
	meta := flagSet.Int("meta", 0, "meta parameter for LogKHybrid")
	complete := flagSet.Bool("complete", false, "Forces the computation of complete decompositions.")
	jCostPath := flagSet.String("joinCost", "", "The file path to a join cost function.")

	parseError := flagSet.Parse(os.Args[1:])
	if parseError != nil {
		fmt.Print("Parse Error:\n", parseError.Error(), "\n\n")
	}

	// Output usage message if graph and width not specified
	if parseError != nil || *graphPath == "" || (*width <= 0 && !*exact && *approx == 0) {
		out := fmt.Sprint("Usage of BalancedGo (", Version, ", https://github.com/cem-okulmus/BalancedGo/commit/",
			Build, ", ", Date, ")")
		fmt.Fprintln(os.Stderr, out)
		flagSet.VisitAll(func(f *flag.Flag) {
			if f.Name != "width" && f.Name != "graph" && f.Name != "exact" && f.Name != "approx" {
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
			if f.Name == "width" || f.Name == "graph" || f.Name == "exact" || f.Name == "approx" {
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

	// END Command-Line Argument Parsing
	// ==============================================

	if *exact && (*approx > 0) {
		fmt.Println("Cannot have exact and approx flags set at the same time. Make up your mind.")
		return
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

	BalFactor := *balanceFactorFlag

	runtime.GOMAXPROCS(*numCPUs)

	dat, err := ioutil.ReadFile(*graphPath)
	check(err)

	var parsedGraph Graph
	var parseGraph lib.ParseGraph

	if !*pace {
		parsedGraph, parseGraph = lib.GetGraph(string(dat))
	} else {
		parsedGraph = lib.GetGraphPACE(string(dat))
	}

	originalGraph := parsedGraph

	if !*bench { // skip any output if bench flag is set
		log.Println("BIP: ", parsedGraph.GetBIP())
	}

	var reducedGraph Graph

	var times []labelTime

	// Sorting Edges to find separators faster
	if *useHeuristic > 0 {
		var heuristicMessage string

		start := time.Now()
		switch *useHeuristic {
		case 1:
			parsedGraph.Edges = lib.GetDegreeOrder(parsedGraph.Edges)
			heuristicMessage = "Using degree ordering as a heuristic"
			break
		case 2:
			parsedGraph.Edges = lib.GetMaxSepOrder(parsedGraph.Edges)
			heuristicMessage = "Using max separator ordering as a heuristic"
			break
		case 3:
			parsedGraph.Edges = lib.GetMSCOrder(parsedGraph.Edges)
			heuristicMessage = "Using MSC ordering as a heuristic"
			break
		case 4:
			parsedGraph.Edges = lib.GetEdgeDegreeOrder(parsedGraph.Edges)
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

	var ops []lib.GYÖReduct
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

	// Complete Decomp preprocessing
	var addedVertices []int
	if *complete {
		addedVertices = parsedGraph.MakeEdgesDistinct()
	}

	// Add all subedges to graph
	if *globalBal && !*computeSubedges {
		parsedGraph = parsedGraph.ComputeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	var hinget lib.Hingetree
	var msecHinge float64

	if *hingeFlag {
		startHinge := time.Now()

		hinget = lib.GetHingeTree(parsedGraph)

		dHinge := time.Now().Sub(startHinge)
		msecHinge = dHinge.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msecHinge, label: "Hingetree"})

		if !*bench {
			fmt.Println("Produced Hingetree: ")
			fmt.Println(hinget)
		}
	}

	var solver algo.Algorithm

	// Check for multiple flags
	chosen := 0

	if *balDetFlag > 0 {
		balDet := &algo.BalSepHybrid{
			K:         *width,
			Graph:     parsedGraph,
			BalFactor: BalFactor,
			Depth:     *balDetFlag - 1,
		}
		solver = balDet
		chosen++
	}

	if *seqBalDetFlag > 0 {
		seqBalDet := &algo.BalSepHybridSeq{
			K:         *width,
			Graph:     parsedGraph,
			BalFactor: BalFactor,
			Depth:     *seqBalDetFlag - 1,
		}
		solver = seqBalDet
		chosen++
	}

	if *detKFlag {
		det := &algo.DetKDecomp{
			K:         *width,
			Graph:     parsedGraph,
			BalFactor: BalFactor,
			SubEdge:   *localBIP,
		}
		solver = det
		chosen++
	}

	if *logK {
		logK := &algo.LogKDecomp{
			Graph:     parsedGraph,
			K:         *width,
			BalFactor: BalFactor,
		}
		solver = logK
		chosen++
	}

	if *logKHybrid > 0 {
		logKHyb := &algo.LogKHybrid{
			Graph:     parsedGraph,
			K:         *width,
			BalFactor: BalFactor,
		}
		logKHyb.Size = *meta

		var pred algo.HybridPredicate

		switch *logKHybrid {
		case 1:
			pred = logKHyb.NumberEdgesPred
		case 2:
			pred = logKHyb.SumEdgesPred
		case 3:
			pred = logKHyb.ETimesKDivAvgEdgePred
		case 4:
			pred = logKHyb.OneRoundPred

		}

		logKHyb.Predicate = pred // set the predicate to use

		solver = logKHyb
		chosen++
	}

	if *globalBal {
		global := &algo.BalSepGlobal{
			K:         *width,
			Graph:     parsedGraph,
			BalFactor: BalFactor,
		}
		solver = global
		chosen++
	}

	if *localBal {
		local := &algo.BalSepLocal{
			K:         *width,
			Graph:     parsedGraph,
			BalFactor: BalFactor,
		}
		solver = local
		chosen++
	}

	if chosen > 1 {
		fmt.Println("Only one algorithm may be chosen at a time. Make up your mind.")
		return
	}

	if *jCostPath != "" {
		if !*localBal && *balDetFlag == 0 {
			fmt.Println("Join cost can be used only in combination with: local, balDet.")
			return
		}
		if *pace {
			fmt.Println("Join cost cannot be used with PACE input format.")
			return
		}

		// load cost function
		// 1. read the csv file
		csvfile, err := os.Open(*jCostPath)
		if err != nil {
			fmt.Println("Can't open jCost", *jCostPath, err)
			return
		}

		// 2. init map
		var w lib.EdgesCostMap
		w.Init()

		r := csv.NewReader(csvfile)
		r.FieldsPerRecord = -1
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				return
			}

			// 3. put the record into the map
			last := len(record) - 1
			cost, _ := strconv.ParseFloat(record[last], 64)
			rec := record[:last]
			comb := make([]int, len(rec))
			for p, s := range rec {
				comb[p] = parseGraph.Encoding[s]
			}
			sort.Ints(comb)
			w.Put(comb, cost)
		}

		//fmt.Println("Printing w:")
		//wComb, wCost := w.Records()
		//for i := 0; i < len(wComb); i++ {
		//	fmt.Println(i, wComb[i], wCost[i])
		//}
		fmt.Println()

		// initialize solver
		if *localBal {
			local := &algo.JCostBalSepLocal{
				K:         *width,
				Graph:     parsedGraph,
				BalFactor: BalFactor,
				JCosts:    w,
			}
			solver = local
			//} else if *globalBal {
			//jGlobal := JCostBalSepGlobal{Graph: parsedGraph, BalFactor: BalancedFactor, JCosts: w}
			//solver = jGlobal
		} else if *balDetFlag != 0 {
			//jBalDet := &algo.JCostBalDetKDecomp{Graph: parsedGraph, BalFactor: BalFactor, Depth: *balDetFlag - 1, JCosts: w}
			//solver = jBalDet
		} else {
			fmt.Println("Weird solver chosen.")
			return
		}
	}

	if solver != nil {

		solver.SetGenerator(lib.ParallelSearchGen{})

		if *logKHybrid > 0 {
			cancelChan := make(chan os.Signal, 1)
			signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

			go func(c chan os.Signal, a algo.AlgorithmDebug) {

				select {
				case <-c:
					algoDebug := a.GetCounters()
					fmt.Println("Counters: \n", algoDebug.String())
				}
			}(cancelChan, solver.(algo.AlgorithmDebug)) // this is risky (maybe add a check if the right flag is set?)
		}

		var decomp Decomp
		start := time.Now()

		if *exact {
			solved := false
			k := 1
			for ; !solved; k++ {
				solver.SetWidth(k)

				if *hingeFlag {
					decomp = hinget.DecompHinge(solver, parsedGraph)
				} else {
					decomp = solver.FindDecomp()
				}

				solved = decomp.Correct(parsedGraph)
			}
			*width = k - 1 // for correct output
		} else if *approx > 0 {
			ch := make(chan int, 1)
			go func() {
				m := parsedGraph.Edges.Len()
				k := int(math.Ceil(float64(m) / 2))
				firstApprox := algo.SplitDecomp{Graph: parsedGraph}
				firstApprox.SetWidth(k)
				decomp = firstApprox.FindDecomp()
				k = decomp.CheckWidth()
				solved := false

				var newDecomp Decomp
				for !solved {
					newK := k - 1
					solver.SetWidth(newK)

					if *hingeFlag {
						newDecomp = hinget.DecompHinge(solver, parsedGraph)
					} else {
						newDecomp = solver.FindDecomp()
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
				decomp = hinget.DecompHinge(solver, parsedGraph)
			} else {
				decomp = solver.FindDecomp()
			}
		}

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)
		times = append(times, labelTime{time: msec, label: "Decomposition"})

		// complete Decomposition post-processing
		if *complete {
			decomp.Root.RemoveVertices(addedVertices)
		}

		if !reflect.DeepEqual(decomp, Decomp{}) || (len(ops) > 0 && parsedGraph.Edges.Len() == 0) {
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

		if !reflect.DeepEqual(decomp, Decomp{}) {
			decomp.Graph = originalGraph
		}
		outputStanza(solver.Name(), decomp, times, originalGraph, *gml, *width, false)

		return
	}

	fmt.Println("No algorithm or procedure selected.")
}
