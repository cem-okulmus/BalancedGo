package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
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

var BALANCED_FACTOR int

// func checkDupTypes (g Graph) int {
// 	encountered := make(map[[]int]struct{})
// }

func main() {
	logActive(false)

	//Command-Line Argument Parsing
	compute_subedes := flag.Bool("sub", false, "(optional) Compute the subedges of the graph and print it out")
	width := flag.Int("width", 0, "a positive, non-zero integer indicating the width of the GHD to search for")
	graphPath := flag.String("graph", "", "the file path to a hypergraph \n(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)")
	choose := flag.Int("choice", 0, "(optional) only run one version\n\t1 ... Full Parallelism\n\t2 ... Search Parallelism\n\t3 ... Comp. Parallelism\n\t4 ... Sequential execution\n\t5 ... Local Full Parallelism\n\t6 ... Local Search Parallelism\n\t7 ... Local Comp. Parallelism\n\t8 ... Local Sequential execution.")
	balance_factor := flag.Int("balfactor", 2, "(optional) Determines the factor that balanced separator check uses")
	use_heuristic := flag.Int("heuristic", 0, "(optional) turn on to activate edge ordering\n\t1 ... Degree Ordering\n\t2 ... Max. Separator Ordering")
	// OW_optim := flag.Bool("OWremoval", false, "(optional) remove edges with single indicent edges and add them to Decomp afterwards")
	// flag.Parse()

	BALANCED_FACTOR = *balance_factor

	if *graphPath == "" || *width <= 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s: \n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	dat, err := ioutil.ReadFile(*graphPath)
	check(err)

	parsedGraph := getGraph(string(dat))

	//fmt.Println("Graph ", parsedGraph)
	//fmt.Println("Min Distance", getMinDistances(parsedGraph))
	//return

	count := 0
	for _, e := range parsedGraph.edges {
		isOW, _ := e.OWcheck(parsedGraph.edges)
		if isOW {
			count++
		}
	}
	fmt.Println("No of OW edges: ", count)

	if *use_heuristic > 0 {
		switch *use_heuristic {
		case 1:
			parsedGraph.edges = getDegreeOrder(parsedGraph.edges)
		case 2:
			parsedGraph.edges = getMaxSepOrder(parsedGraph.edges)
		}
	}

	if *compute_subedes {
		parsedGraph = parsedGraph.computeSubEdges(*width)

		fmt.Println("Graph with subedges \n", parsedGraph)
	}

	global := balsepGlobal{graph: parsedGraph}
	local := balsepLocal{graph: parsedGraph}
	if *choose != 0 {
		var decomp Decomp
		start := time.Now()
		switch *choose {
		case 1:
			decomp = global.findGHDParallelFull(*width)
		case 2:
			decomp = global.findGHDParallelSearch(*width)
		case 3:
			decomp = global.findGHDParallelComp(*width)
		case 4:
			decomp = global.findGHD(*width)
		case 5:
			decomp = local.findGHDParallelFull(*width)
		case 6:
			decomp = local.findGHDParallelSearch(*width)
		case 7:
			decomp = local.findGHDParallelComp(*width)
		case 8:
			decomp = local.findGHD(*width)
		default:
			panic("Not a valid choice")
		}
		d := time.Now().Sub(start)
		msec := d.Seconds() * float64(time.Second/time.Millisecond)

		fmt.Println("Result \n", decomp)
		fmt.Println("Time", msec, " ms")
		fmt.Println("Width: ", *width)
		//fmt.Println("Correct: ", decomp.correct(parsedGraph))
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

	output = output + fmt.Sprintf("%v;", len(parsedGraph.edges))
	output = output + fmt.Sprintf("%v;", len(Vertices(parsedGraph.edges)))
	output = output + fmt.Sprintf("%v;", *width)

	// Parallel Execution FULL
	start := time.Now()
	decomp := global.findGHDParallelFull(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d := time.Now().Sub(start)
	msec := d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v;", decomp.correct(parsedGraph))

	// Parallel Execution Search
	start = time.Now()
	decomp = global.findGHDParallelSearch(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v;", decomp.correct(parsedGraph))

	// Parallel Execution Comp
	start = time.Now()
	decomp = global.findGHDParallelComp(*width)

	//fmt.Printf("Decomp of parsedGraph:\n%v\n", decomp.root)

	//fmt.Println("Elapsed time for parallel:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v;", decomp.correct(parsedGraph))

	// Sequential Execution
	start = time.Now()
	decomp = global.findGHD(*width)

	//fmt.Printf("Decomp of parsedGraph: %v\n", decomp.root)
	d = time.Now().Sub(start)
	msec = d.Seconds() * float64(time.Second/time.Millisecond)
	output = output + fmt.Sprintf("%.5f;", msec)
	output = output + fmt.Sprintf("%v\n", decomp.correct(parsedGraph))
	//fmt.Println("Elapsed time for sequential:", time.Now().Sub(start))
	//fmt.Println("Correct decomposition:", decomp.correct())

	fmt.Print(output)

}
