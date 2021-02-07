package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

type Algorithm interface {
	Name() string
	FindDecomp() Decomp
	FindDecompGraph(G Graph) Decomp
	SetWidth(K int)
}

