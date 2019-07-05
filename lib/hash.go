package lib

import (
	"encoding/json"
	"hash/fnv"
)

// implements hashes for basic types (used for hash table implementations)

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
