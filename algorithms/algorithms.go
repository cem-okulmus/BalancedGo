package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

// Algorithm serves as the common interfacea of all hypergraph decomposition algorithms
type Algorithm interface {
	// A Name is useful to identify the individual algorithms in the result
	Name() string
	FindDecomp() Decomp
	FindDecompGraph(G Graph) Decomp
	SetWidth(K int)
}
