package lib

type Separator struct {
	Found    []int
	EdgeComb []int
	Cost     float64
}

type JoinHeap []*Separator

func (jh JoinHeap) Len() int { return len(jh) }

func (jh JoinHeap) Less(i, j int) bool {
	return jh[i].Cost < jh[j].Cost // make it reliable
}

func (jh JoinHeap) Swap(i, j int) {
	jh[i], jh[j] = jh[j], jh[i]
}

func (jh *JoinHeap) Push(x interface{}) {
	item := x.(*Separator)
	*jh = append(*jh, item)
}

func (jh *JoinHeap) Pop() interface{} {
	old := *jh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*jh = old[0 : n-1]
	return item
}
