package lib

// search.go implements a parallel search over a set of edges with a given predicate to look for

import (
	"runtime"
	"sync"
)

// A Search implements a parallel search for separators fulfilling some given predicate
type Search struct {
	H               *Graph
	Edges           *Edges
	BalFactor       int
	Result          []int
	Generators      []*CombinationIterator
	ExhaustedSearch bool
}

// A Predicate checks if for some subgraph and a separator, some condition holds
type Predicate interface {
	Check(H *Graph, sep *Edges, balancedFactor int) bool
}

// FindNext starts the search and stops if some separator which satisfies the predicate
// is found, or if the entire search space has been exhausted
func (s *Search) FindNext(pred Predicate) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	// log.Println("starting search")

	s.Result = []int{} // reset result

	var numProc = runtime.GOMAXPROCS(-1)

	var wg sync.WaitGroup
	wg.Add(numProc)
	finished := false
	// SEARCH:
	found := make(chan []int)
	wait := make(chan bool)
	//start workers
	for i := 0; i < numProc; i++ {
		go s.worker(i, found, &wg, &finished, pred)
	}

	go func() {
		wg.Wait()
		wait <- true
	}()

	select {
	case s.Result = <-found:
		close(found) //to terminate other workers waiting on found
		wg.Wait()
	case <-wait:
		s.ExhaustedSearch = true
	}

}

// a worker that actually runs the search within a single goroutine
func (s Search) worker(workernum int, found chan []int, wg *sync.WaitGroup, finished *bool, pred Predicate) {
	defer func() {
		if r := recover(); r != nil {
			// log.Printf("Worker %d 'forced' to quit, reason: %v", workernum, r)
			return
		}
	}()
	defer wg.Done()

	gen := s.Generators[workernum]

	for gen.hasNext() {
		if *finished {
			// log.Printf("Worker %d told to quit", workernum)
			return
		}
		// j := make([]int, len(gen.Combination))
		// copy(gen.Combination, j)
		j := gen.combination

		sep := GetSubset(*s.Edges, j)
		if pred.Check(s.H, &sep, s.BalFactor) {
			gen.balSep = true // cache result
			found <- j
			// log.Printf("Worker %d \" won \"", workernum)
			gen.confirm()

			// if !pred.IsParent() {
			// 	fmt.Println("Worker ", workernum, ": found balsep of combin", j, "( ", GetSubset(*s.Edges, j), ") from the set: ", *s.Edges)

			// }

			*finished = true
			return
		}
		gen.confirm()
	}
}

// BalancedCheck looks for Balanced Separators
type BalancedCheck struct{}

// Check performs the needed computation to ensure whether sep is a Balanced Separator
func (b BalancedCheck) Check(H *Graph, sep *Edges, balFactor int) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, _, _ := H.GetComponents(*sep)
	// log.Printf("Components of sep %+v\n", comps)

	balancednessLimit := (((H.Len()) * (balFactor - 1)) / balFactor)

	for i := range comps {
		if comps[i].Len() > balancednessLimit {
			//log.Printf("Using %+v component %+v has weight %d instead of %d\n", sep,
			//        comps[i], comps[i].Edges.Len()+len(compSps[i]), ((g.Edges.Len() + len(sp)) / 2))
			return false
		}
	}

	// Make sure that "special seps can never be used as separators"
	for i := range H.Special {
		if IntHash(H.Special[i].Vertices()) == IntHash(sep.Vertices()) {
			//log.Println("Special edge %+v\n used again", s)
			return false
		}
	}

	// fmt.Println("\nBALANECDCHECK:   I found the balsep ", sep, "for the hypergraph ", H, "\n\n")

	return true
}

// ParentCheck looks a separator that could function as the direct ancestor (or "parent")
// of some child node in the GHD, where the connecting vertices "Conn" are explicitly provided
type ParentCheck struct {
	Conn  []int
	Child []int
}

// Check performs the needed computation to ensure whether sep is a good parent
func (p ParentCheck) Check(H *Graph, sep *Edges, balFactor int) bool {
	// log.Printf("Current considered sep %+v\n", sep)
	// log.Printf("Current present SP %+v\n", sp)

	//balancedness condition
	comps, _, _ := H.GetComponents(*sep)

	foundCompLow := false
	var compLow Graph

	// log.Printf("Components of sep %+v\n", comps)

	balancednessLimit := (((H.Len()) * (balFactor - 1)) / balFactor)

	for i := range comps {
		if comps[i].Len() > balancednessLimit {
			foundCompLow = true
			compLow = comps[i]
		}
	}

	if !foundCompLow {
		return false // a bad parent :(
	}

	vertCompLow := compLow.Vertices()
	childχ := Inter(p.Child, vertCompLow)

	if !Subset(Inter(vertCompLow, p.Conn), sep.Vertices()) {
		// log.Println("Conn not covered by parent")
		// log.Println("Conn: ", PrintVertices(Conn))
		// log.Println("V(parentλ) \\cap Conn", PrintVertices(Inter(parentλ.Vertices(), Conn)))
		// log.Println("V(Comp_lo w) \\cap Conn ", PrintVertices(Inter(vertCompLow, Conn)))

		return false // also a bad parent :(
	}

	// Connectivity check
	if !Subset(Inter(vertCompLow, sep.Vertices()), childχ) {
		// log.Println("Child not connected to parent!")
		// log.Println("Parent lambda: ", PrintVertices(parentλ.Vertices()))
		// log.Println("Child lambda: ", PrintVertices(childλ.Vertices()))
		// log.Println("Child", childλ)

		return false // again a bad parent :( Calling child services ...
	}

	// fmt.Println("\nPARENTCHECK: I found the parent ", sep, " relative to child ", PrintVertices(p.Child), "for the hypergraph ", H, "\n\n")

	return true // found a good parent :)
}
