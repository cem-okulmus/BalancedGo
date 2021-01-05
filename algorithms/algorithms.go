package algorithms

import . "github.com/cem-okulmus/BalancedGo/lib"

type Algorithm interface {
	Name() string
	FindDecomp() Decomp
	FindDecompGraph(G Graph) Decomp
	SetWidth(K int)
}

type UpdateAlgorithm interface {
	Name() string
	FindDecompUpdate(currentGraph Graph, savedScenes map[uint64]SceneValue, savedCache Cache) Decomp
	SetWidth(K int)
	GetCache() Cache
}
