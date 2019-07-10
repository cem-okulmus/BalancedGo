package lib

import (
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
)

// implements hashes for basic types (used for hash table implementations)

func (e Edge) Hash() uint32 {

	arrBytes := []byte{}
	for _, item := range e.Vertices {
		bs := make([]byte, 4)
		binary.PutVarint(bs, int64(item))
	}
	h := fnv.New32a()
	h.Write(arrBytes)
	return h.Sum32()

}

func (e Edges) Hash() uint32 {

	arrBytes := []byte{}
	for _, item := range e {
		jsonBytes, _ := json.Marshal(item)
		arrBytes = append(arrBytes, jsonBytes...)
	}
	h := fnv.New32a()
	h.Write(arrBytes)
	return h.Sum32()
}
