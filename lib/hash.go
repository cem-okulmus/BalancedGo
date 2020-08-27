package lib

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
)

// implements hashes for basic types (used for hash table implementations)
var hashMux sync.Mutex

func IntHash(vertices []int) uint32 {
	var output uint32
	// arrBytes := []byte{}
	//sort.Ints(vertices)
	for _, item := range vertices {

		h := fnv.New32a()
		bs := make([]byte, 4)
		binary.PutVarint(bs, int64(item))
		// arrBytes = append(arrBytes, bs...)
		h.Write(bs)
		output = output + h.Sum32()

	}
	return output

}

func (e Edge) Hash() uint32 {
	var output uint32
	// arrBytes := []byte{}
	//	sort.Ints(e.Vertices)
	for _, item := range e.Vertices {
		h := fnv.New32a()
		bs := make([]byte, 4)
		binary.PutVarint(bs, int64(item))
		// arrBytes = append(arrBytes, bs...)
		h.Write(bs)
		output = output + h.Sum32()
	}

	// h.Write(arrBytes)
	// return h.Sum32()
	return output

}

func (e *Edges) Hash() uint64 {
	if e.hash != nil {
		return *e.hash
	}
	var output uint64

	e.hashMux.Lock() // ensure that hash is computed only on one gorutine at a time
	if e.hash == nil {
		// cpy := make([]Edge, len(e.slice))
		// copy(cpy, e.slice)
		// copyE := NewEdges(cpy)
		// arrBytes := []byte{}
		// sort.Sort(copyE)

		//sort.Sort(e)
		//	fmt.Println(e)
		for i := range e.Slice() {
			h := fnv.New64a()
			bs := make([]byte, 8)
			binary.LittleEndian.PutUint64(bs, uint64(e.Slice()[i].Hash()))
			// arrBytes = append(arrBytes, bs...)
			h.Write(bs)
			output = output + h.Sum64()
		}
		// h := fnv.New32a()
		// h.Write(arrBytes)
		// result := h.Sum32()
		e.hash = &output
	}
	e.hashMux.Unlock()

	return *e.hash
}

// func testHash() {

// 	e1 := Edge{Vertices: []int{58, 96, 97}}
// 	e2 := Edge{Vertices: []int{65, 66, 67}}
// 	//	e3 := Edge{Vertices: []int{61, 18, 7}}

// 	edges := Edges{slice: []Edge{e2, e1}}
// 	fmt.Println("Edges ", edges)

// 	fmt.Println("Hash 1", edges.Hash())
// 	sort.Sort(edges)
// 	fmt.Println("Hash 1", edges.Hash())
// 	sort.Sort(edges)
// 	fmt.Println("Hash 1", edges.Hash())
// 	// sort.Sort(edges)
// 	// fmt.Println("Hash 1", edges)
// 	// sort.Sort(edges)
// 	// fmt.Println("Hash 1", edges)
// 	// sort.Sort(edges)
// 	// fmt.Println("Hash 1", edges)

// 	// var cache map[uint32][]uint32

// 	fmt.Println("Edges ", edges)

// }
