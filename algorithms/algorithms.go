package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

type Algorithm interface {
	Name() string
	FindDecomp(K int) Decomp
}

type UpdateAlgorithm interface {
	Name() string
	FindDecompUpdate(K int, Sp []Special) Decomp
}
