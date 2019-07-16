package lib

import (
	"encoding/binary"
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

func (e Edges) Hash() uint32 {

	arrBytes := []byte{}
	sort.Sort(Edges(e)) // cache this via flag
	for _, item := range e.Slice {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, item.Hash())
		arrBytes = append(arrBytes, bs...)
	}
	h := fnv.New32a()
	h.Write(arrBytes)
	return h.Sum32()
}
