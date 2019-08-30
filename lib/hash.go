package lib

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
)

// implements hashes for basic types (used for hash table implementations)

func (e Edge) Hash() uint32 {

	arrBytes := []byte{}
	sort.Ints(e.Vertices)
	for _, item := range e.Vertices {
		bs := make([]byte, 4)
		binary.PutVarint(bs, int64(item))
		arrBytes = append(arrBytes, bs...)
	}
	h := fnv.New32a()
	h.Write(arrBytes)
	return h.Sum32()

}

func (e *Edges) Hash() uint32 {
	if e.hash == nil {
		arrBytes := []byte{}
		sort.Sort(*e)
		//	fmt.Println(e)
		for _, item := range e.Slice() {
			bs := make([]byte, 4)
			binary.LittleEndian.PutUint32(bs, item.Hash())
			arrBytes = append(arrBytes, bs...)
		}
		h := fnv.New32a()
		h.Write(arrBytes)
		result := h.Sum32()
		e.hash = &result
	}

	return *e.hash
}

func testHash() {

	e1 := Edge{Vertices: []int{58, 96, 97}}
	e2 := Edge{Vertices: []int{65, 66, 67}}

	edges := Edges{slice: []Edge{e1, e2}}

	sort.Sort(edges)
	fmt.Println("Hash 1", edges)
	sort.Sort(edges)
	fmt.Println("Hash 1", edges)
	sort.Sort(edges)
	fmt.Println("Hash 1", edges)
	sort.Sort(edges)
	fmt.Println("Hash 1", edges)
	sort.Sort(edges)
	fmt.Println("Hash 1", edges)
	sort.Sort(edges)
	fmt.Println("Hash 1", edges)

}
