package lib

// search.go implements a parallel search over a set of edges with a given predicate to look for

import (
	"runtime"
	"sync"

	"github.com/cem-okulmus/disjoint"
)

// A Search implements a parallel search for separators fulfilling some given predicate
type Search interface {
	FindNext(pred Predicate)
	SearchEnded() bool // return true if every element has been returned once
	GetResult() []int  // get the last found result
}

// A SearchGenerator sets up a Search interface
type SearchGenerator interface {
	GetSearch(H *Graph, Edges *Edges, BalFactor int, Generators []Generator) Search
}

// A Generator is black-box view for any kind of generation of items to look at in linear order, and it provides some helpful methods for the search
type Generator interface {
	HasNext() bool    // check if generator still has new elements
	GetNext() []int   // the slice of int represents some choice of edges, with an underlying order
	Confirm()         // confirm that the current selection has been checked *and* sent to central goroutine
	Found()           // used to cache the check result
	CheckFound() bool // used by search to see if previous run already performed the check
}

type ParallelSearch struct {
	H               *Graph
	Edges           *Edges
	BalFactor       int
	Result          []int
	Generators      []Generator
	ExhaustedSearch bool
}

type ParallelSearchGen struct{}

func (p ParallelSearchGen) GetSearch(H *Graph, Edges *Edges, BalFactor int, Gens []Generator) Search {
	return &ParallelSearch{
		H:               H,
		Edges:           Edges,
		BalFactor:       BalFactor,
		Result:          []int{},
		Generators:      Gens,
		ExhaustedSearch: false,
	}
}

// SearchEnded returns true if search is completed
func (s *ParallelSearch) SearchEnded() bool {
	return s.ExhaustedSearch
}

// GetResult returns the last found result
func (s *ParallelSearch) GetResult() []int {
	return s.Result
}

// A Predicate checks if for some subgraph and a separator, some condition holds
type Predicate interface {
	Check(H *Graph, sep *Edges, balancedFactor int, Vertices map[int]*disjoint.Element) bool
}

// FindNext starts the search and stops if some separator which satisfies the predicate
// is found, or if the entire search space has been exhausted
func (s *ParallelSearch) FindNext(pred Predicate) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	s.Result = []int{} // reset result
	var numProc int
	if runtime.GOMAXPROCS(-1) > len(s.Generators) {
		numProc = len(s.Generators)
	} else {
		numProc = runtime.GOMAXPROCS(-1)
	}

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
func (s ParallelSearch) worker(workernum int, found chan []int, wg *sync.WaitGroup, finished *bool, pred Predicate) {
	defer func() {
		if r := recover(); r != nil {
			// log.Printf("Worker %d 'forced' to quit, reason: %v", workernum, r)
			return
		}
	}()
	defer wg.Done()
	var Vertices = make(map[int]*disjoint.Element)

	gen := s.Generators[workernum]

	for gen.HasNext() {
		if *finished {
			// log.Printf("Worker %d told to quit", workernum)
			return
		}
		// j := make([]int, len(gen.Combination))
		// copy(gen.Combination, j)
		j := gen.GetNext()

		sep := GetSubset(*s.Edges, j)
		if pred.Check(s.H, &sep, s.BalFactor, Vertices) {
			gen.Found() // cache result
			found <- j
			// log.Println("Worker", workernum, "won, found: ", j)
			gen.Confirm()
			*finished = true
			return
		}
		gen.Confirm()
	}
}

// BalancedCheck looks for Balanced Separators
type BalancedCheck struct{}

// Check performs the needed computation to ensure whether sep is a Balanced Separator
func (b BalancedCheck) Check(H *Graph, sep *Edges, balFactor int, Vertices map[int]*disjoint.Element) bool {

	//balancedness condition
	comps, _, _ := H.GetComponents(*sep, Vertices)

	balancednessLimit := (((H.Len()) * (balFactor - 1)) / balFactor)

	for i := range comps {
		if comps[i].Len() > balancednessLimit {
			return false
		}
	}

	// Make sure that "special seps can never be used as separators"
	for i := range H.Special {
		if IntHash(H.Special[i].Vertices()) == IntHash(sep.Vertices()) {
			return false
		}
	}

	return true
}

// CheckOut does the same as Check, except it also passes on the components found, if output is true
func (b BalancedCheck) CheckOut(H *Graph, sep *Edges, balFactor int, Vertices map[int]*disjoint.Element) (bool, []Graph, []Edge) {

	//balancedness condition
	comps, _, isolated := H.GetComponents(*sep, Vertices)

	balancednessLimit := (((H.Len()) * (balFactor - 1)) / balFactor)

	for i := range comps {
		if comps[i].Len() > balancednessLimit {
			return false, []Graph{}, []Edge{}
		}
	}

	// Make sure that "special seps can never be used as separators"
	for i := range H.Special {
		if IntHash(H.Special[i].Vertices()) == IntHash(sep.Vertices()) {
			return false, []Graph{}, []Edge{}
		}
	}

	return true, comps, isolated
}
