package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
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

// Heuristics to order the edges by

func getMSCOrder(edges []Edge) []Edge {
	var selected []Edge
	var chosen []bool

	//randomly select last edge in the ordering
	i := rand.Intn(len(edges))
	chosen[i] = true
	selected = append(selected, edges[i])

	for len(selected) < len(edges) {
		var candidates []int
		maxcard := 0

		for current := range edges {
			current_card := edges[current].numNeighbours(edges, chosen)
			if !chosen[current] && current_card >= maxcard {
				if current_card > maxcard {
					candidates = []int{}
					maxcard = current_card
				}

				candidates = append(candidates, current)
			}
		}

		//randomly select one of the edges with equal connectivity
		next_in_order := candidates[rand.Intn(len(candidates))]

		selected = append(selected, edges[next_in_order])
		chosen[next_in_order] = true
	}

	//reverse order of selected
	for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
		selected[i], selected[j] = selected[j], selected[i]
	}

	return selected
}

//Order the edges by how much  they increase shortest paths within the hypergraph

//basic Floyd-Warschall (using the primal graph)

func order(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func isInf(a int) bool {
	return a == math.MaxInt64
}

func addEdgeDistances(order map[int]int, output [][]int, e Edge) [][]int {

	for _, n := range e.vertices {
		for _, m := range e.vertices {
			n_index, _ := order[n]
			m_index, _ := order[m]
			if n_index != m_index {
				output[n_index][m_index] = 1
			}
		}
	}

	return output
}

func getMinDistances(edges []Edge) ([][]int, map[int]int) {
	var output [][]int
	order := make(map[int]int)
	vertices := Vertices(edges)

	for i, n := range vertices {
		order[n] = i
	}

	row := make([]int, len(vertices))
	for j := 0; j < len(vertices); j++ {
		row[j] = math.MaxInt64
	}

	for j := 0; j < len(vertices); j++ {
		new_row := make([]int, len(vertices))
		copy(new_row, row)
		output = append(output, new_row)
	}

	for _, e := range edges {
		output = addEdgeDistances(order, output, e)
	}

	for j := 0; j < len(edges); j++ {
		changed := false
		for k := range vertices {
			for l := range vertices {
				for m := range vertices {
					if isInf(output[k][l]) || isInf(output[l][m]) {
						continue
					}
					newdist := output[k][l] + output[l][m]
					if output[k][m] > newdist {
						output[k][m] = newdist
						changed = true
					}
				}
			}
		}
		if !changed {
			break
		}

	}

	return output, order
}

//  weight of each edge = (sum of path disconnected)*SepWeight  +  (sum of each path made longer * diff)
func diffDistances(old, new [][]int) int {
	var output int

	SepWeight := len(old) * len(old)

	for j := 0; j < len(old); j++ {
		for i := 0; i < len(old[j]); i++ {
			if isInf(old[j][i]) && !isInf(new[j][i]) { // disconnected a path
				output = output + SepWeight
			} else if !isInf(old[j][i]) && !isInf(new[j][i]) { // check if parth shortened
				diff := old[j][i] - new[j][i]
				output = output + diff
			}
		}
	}

	return output
}

func getMaxSepOrder(edges []Edge) []Edge {
	weights := make([]int, len(edges))

	initialDiff, order := getMinDistances(edges)

	for i, e := range edges {
		edges_wihout_e := diffEdges(edges, e)
		newDiff, _ := getMinDistances(edges_wihout_e)
		newDiffPrep := addEdgeDistances(order, newDiff, e)
		weights[i] = diffDistances(initialDiff, newDiffPrep)
	}

	sort.Slice(edges, func(i, j int) bool { return weights[i] > weights[j] })

	return edges
}

func edgeDegree(edges []Edge, edge Edge) int {
	var output int

	for _, v := range edge.vertices {
		output = output + getDegree(edges, v)
	}

	return output - len(edge.vertices)
}

func getDegreeOrder(edges []Edge) []Edge {
	sort.Slice(edges, func(i, j int) bool { return edgeDegree(edges, edges[i]) > edgeDegree(edges, edges[j]) })
	return edges
}

var BALANCED_FACTOR int

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
