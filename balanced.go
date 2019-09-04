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
var Build string

func main() {

	// m = make(map[int]string)

	//Command-Line Argument Parsing
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	logging := flag.Bool("log", false, "turn on extensive logs")
	computeSubedges := flag.Bool("sub", false, "Compute the subedges of the graph and print it out")
	width := flag.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	graphPath := flag.String("graph", "", "the file path to a hypergraph \n\t(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)")
	choose := flag.Int("choice", 0, "only run one version\n\t1 ... Full Parallelism\n\t2 ... Search Parallelism\n\t3 ... Comp. Parallelism\n\t4 ... Sequential execution\n\t5 ... Local Full Parallelism\n\t6 ... Local Search Parallelism\n\t7 ... Local Comp. Parallelism\n\t8 ... Local Sequential execution.")
	balanceFactorFlag := flag.Int("balfactor", 2, "Determines the factor that balanced separator check uses")
	useHeuristic := flag.Int("heuristic", 0, "turn on to activate edge ordering\n\t1 ... Degree Ordering\n\t2 ... Max. Separator Ordering\n\t3 ... MCSO")
	gyö := flag.Bool("g", false, "perform a GYÖ reduct and show the resulting graph")
	typeC := flag.Bool("t", false, "perform a Type Collapse and show the resulting graph")
	//hingeFlag := flag.Bool("hinge", false, "use isHinge Optimization")
	numCPUs := flag.Int("cpu", -1, "Set number of CPUs to use")
	bench := flag.Bool("bench", false, "Benchmark mode, reduces unneeded output (incompatible with -log flag)")
	akatovTest := flag.Bool("akatov", false, "compute balanced decomposition")
	detKTest := flag.Bool("det", false, "Test out DetKDecomp")
	divideTest := flag.Bool("divide", false, "Test for divideKDecomp")
	balDetTest := flag.Int("balDet", 0, "Test hybrid balSep and DetK algorithm")

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
	if *graphPath == "" || *width <= 0 {
		fmt.Fprintf(os.Stderr, "Usage of BalancedGo (v%s, %s): \n", Version, Build)
		flag.VisitAll(func(f *flag.Flag) {
			if f.Name != "width" && f.Name != "graph" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			fmt.Printf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			fmt.Println("\t" + f.Usage)
		})

		fmt.Println("\nOptional Arguments: ")
		flag.VisitAll(func(f *flag.Flag) {
			if f.Name == "width" || f.Name == "graph" {
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

	parsedGraph, _ := GetGraph(string(dat))
	var reducedGraph Graph

	start := time.Now()
	switch *useHeuristic {
	case 1:
		parsedGraph.Edges = GetDegreeOrder(parsedGraph.Edges)
	case 2:
		parsedGraph.Edges = GetMaxSepOrder(parsedGraph.Edges)
	case 3:
		parsedGraph.Edges = GetMSCOrder(parsedGraph.Edges)
	}
	d := time.Now().Sub(start)

	if *useHeuristic > 0 && !*bench {
		switch *useHeuristic {
		case 1:
			fmt.Print("Using degree ordering")
		case 2:
			fmt.Print("Using max seperator ordering")
		case 3:
			fmt.Print("Using MSC ordering")
		}
		fmt.Println(" as a heuristic")
		msec := d.Seconds() * float64(time.Second/time.Millisecond)
		fmt.Printf("Time for heuristic: %.5f ms\n", msec)
		log.Printf("Ordering: %v", parsedGraph.String())

	}

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
	//fmt.Println("Graph ", parsedGraph)
	//fmt.Println("Min Distance", getMinDistances(parsedGraph))
	//return

	// count := 0
	// for _, e := range parsedGraph.Edges {
	// 	isOW, _ := e.OWcheck(parsedGraph.Edges)
	// 	if isOW {
	// 		count++
	// 	}
	// }
	// fmt.Println("No of OW Edges: ", count)

	// collapsedGraph, _ := parsedGraph.typeCollapse()

	// fmt.Println("No of vertices collapsable: ", len(parsedGraph.Vertices())-len(collapsedGraph.Vertices()))

	if *computeSubedges {
		parsedGraph = parsedGraph.ComputeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	if *akatovTest {
		var decomp Decomp
		start := time.Now()
		bal := BalKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor}

		switch *choose {
		case 1:
			decomp = bal.FindBDFullParallel(*width)
		// case 2:
		// 	decomp = global.FindGHDParallelSearch(*width)
		// case 3:
		// 	decomp = global.FindGHDParallelComp(*width)
		case 4:
			decomp = bal.FindBD(*width)
		default:
			panic("Not a valid choice")
		}

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", decomp.CheckWidth())
		fmt.Println("GHD-Width: ", decomp.Blowup().CheckWidth())
		fmt.Println("Correct: ", decomp.Correct(parsedGraph))
		return
	}

	if *balDetTest > 0 {
		var decomp Decomp
		start := time.Now()

		balDet := BalDetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, Depth: *balDetTest}
		decomp = balDet.FindGHD(*width)

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", decomp.CheckWidth())
		fmt.Println("Correct: ", decomp.Correct(parsedGraph))
		return
	}

	if *detKTest {
		var decomp Decomp
		start := time.Now()

		var Sp []Special
		// m[encode] = "test"
		// m[encode+1] = "test2"
		// Sp = []Special{Special{Vertices: []int{16, 18}, Edges: []Edge{Edge{Name: encode, Vertices: []int{16, 18}}}}, Special{Vertices: []int{15, 17, 19}, Edges: []Edge{Edge{Name: encode + 1, Vertices: []int{15, 17, 19}}}}}
		// encode = encode + 2

		det := DetKDecomp{Graph: parsedGraph, BalFactor: BalancedFactor, SubEdge: false}
		decomp = det.FindHD(*width, Sp)

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", decomp.CheckWidth())
		fmt.Println("Correct: ", decomp.Correct(parsedGraph))
		return
	}

	if *divideTest {
		var decomp Decomp
		start := time.Now()

		div := DivideKDecomp{Graph: parsedGraph, K: *width, BalFactor: BalancedFactor}
		decomp = div.FindDecomp()

		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", decomp.CheckWidth())
		fmt.Println("Correct: ", decomp.Correct(parsedGraph))
		return
	}

	global := BalSepGlobal{Graph: parsedGraph, BalFactor: BalancedFactor}
	local := BalSepLocal{Graph: parsedGraph, BalFactor: BalancedFactor}
	if *choose != 0 {
		var decomp Decomp
		start := time.Now()
		switch *choose {
		case 1:
			decomp = global.FindGHDParallelFull(*width)
		case 2:
			decomp = global.FindGHDParallelSearch(*width)
		case 3:
			decomp = global.FindGHDParallelComp(*width)
		case 4:
			decomp = global.FindGHD(*width)
		case 5:
			decomp = local.FindGHDParallelFull(*width)
			//decomp.RestoreSubedges()
		case 6:
			decomp = local.FindGHDParallelSearch(*width)
			//decomp.RestoreSubedges()
		case 7:
			decomp = local.FindGHDParallelComp(*width)
			//decomp.RestoreSubedges()
		case 8:
			decomp = local.FindGHD(*width)
		//	decomp.RestoreSubedges()
		default:
			panic("Not a valid choice")
		}
		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Graph \n", parsedGraph)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", *width)
		fmt.Println("Correct: ", decomp.Correct(parsedGraph))
		return
	}

	var output string

	// f, err := os.OpenFile("result.csv", os.O_APPEND|os.O_WRONLY, 0666)
	// if os.IsNotExist(err) {
	// 	f, err = os.Create("result.csv")
	// 	check(err)
	// 	f.WriteString("graph;edges;vertices;width;time parallel (ms) F;decomposed;time parallel S (ms);decomposed;time parallel C (ms);decomposed; time sequential (ms);decomposed\n")
	// }
	// defer f.Close()

	//fmt.Println("Width: ", *width)
	//fmt.Println("graphPath: ", *graphPath)

	output = output + *graphPath + ";"

	//f.WriteString(*graphPath + ";")
	//f.WriteString(fmt.Sprintf("%v;", *width))

	//fmt.Printf("parsedGraph %+v\n", parsedGraph)

	output = output + fmt.Sprintf("%v;", parsedGraph.Edges.Len())
	output = output + fmt.Sprintf("%v;", len(parsedGraph.Edges.Vertices()))
	output = output + fmt.Sprintf("%v;", *width)

	// Parallel Execution FULL
	start = time.Now()
	decomp := local.FindGHDParallelFull(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.Root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d = time.Now().Sub(start)
	msec := d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)

	// Parallel Execution Search
	start = time.Now()
	decomp = local.FindGHDParallelSearch(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.Root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)

	// Parallel Execution Comp
	start = time.Now()
	decomp = local.FindGHDParallelComp(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.Root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	// Sequential Execution
	start = time.Now()
	decomp = local.FindGHD(*width)

	//fmt.Printf("Decomp of parsedGraph: %v\n", decomp.Root)
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v\n", decomp.Correct(parsedGraph))
	//fmt.Println("Elapsed time for sequential:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())

	fmt.Print(output)

}
