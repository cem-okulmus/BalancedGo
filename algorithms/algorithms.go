// Package algorithms implements various algorithms to compute Generalized Hypertree Decompositions as well as
// the more restricted set of Hypertree Decompositions.
package algorithms

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/cem-okulmus/BalancedGo/lib"
)

// Algorithm serves as the common interface of all hypergraph decomposition algorithms
type Algorithm interface {
	// A Name is useful to identify the individual algorithms in the result
	SetGenerator(S lib.SearchGenerator)
	Name() string
	FindDecomp() lib.Decomp
	FindDecompGraph(G lib.Graph) lib.Decomp
	SetWidth(K int)
}

// Counters allow to track how often an algorithm had to backtrack, and at which level, and the toplevel completion as
// a percentage value between [0,1)
type Counters struct {
	backtrack          map[int]int
	cacheMux           *sync.RWMutex
	topLevelCompletion []lib.Generator
}

// CopyRef allows for safe copying of a cache by reference, not value
func (c *Counters) CopyRef(other *Counters) {
	if c.backtrack == nil { // to be sure only an initialised cache is copied
		c.Init()
	}
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	if other == nil {
		var temp Counters
		other = &temp
	}

	other.backtrack = c.backtrack
	other.cacheMux = c.cacheMux

}

// Init is used to set up the Counters struct
func (c *Counters) Init() {
	if c.backtrack == nil {
		var mux sync.RWMutex
		c.cacheMux = &mux
		c.backtrack = make(map[int]int)
	}
}

// AddBacktrack enables a thread-safe way to add new backtracks to the counter
func (c *Counters) AddBacktrack(level int) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	c.backtrack[level] = c.backtrack[level] + 1
}

func (c *Counters) String() string {
	var buffer bytes.Buffer

	for k, v := range c.backtrack {
		buffer.WriteString(fmt.Sprintln("Found ", v, " backtracks at level ", k))
	}

	// nominTotal, denomTotal := lib.GetPercentagesSlice(c.topLevelCompletion)
	// buffer.WriteString(fmt.Sprintln("Toplevel completion of was ", nominTotal, denomTotal, float32(nominTotal)/float32(denomTotal), "%"))

	return buffer.String()
}

// An AlgorithmDebug exports internal counters to see how far the computation has progressed. To be extracted in case
// of a timeout.
type AlgorithmDebug interface {
	GetCounters() Counters // GetCounters returns the counters collected during a run
}
