// This package implements various algorithms to compute Generalized Hypertree Decompositions as well as
// the more restricted set of Hypertree Deocmpositions.

package algorithms

import "github.com/cem-okulmus/BalancedGo/lib"

// Algorithm serves as the common interface of all hypergraph decomposition algorithms
type Algorithm interface {
	// A Name is useful to identify the individual algorithms in the result
	Name() string
	FindDecomp() lib.Decomp
	FindDecompGraph(G lib.Graph) lib.Decomp
	SetWidth(K int)
}
