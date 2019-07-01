// Combination of BalSep and DetKDecomp, executing Balsep first (for constant number of rounds) then switching to DetKDecomp
package main

type balDetKDecomp struct {
	graph Graph
}

// func (b balDetKDecomp) findBD(K int) Decomp {
// 	return b.findDecomp(K, b.graph)
// }
