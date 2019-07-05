// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to DetKDecomp
package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

type BalDetKDecomp struct {
	Graph     Graph
	BalFactor int
}

// func (b balDetKDecomp) findBD(K int) Decomp {
// 	return b.findDecomp(K, b.Graph)
// }
