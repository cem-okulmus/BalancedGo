package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

type Algorithm interface {
	Name() string
	FindDecomp(K int) Decomp
}
