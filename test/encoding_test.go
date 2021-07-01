package tests

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/cem-okulmus/BalancedGo/lib"
)

func TestEncoding(t *testing.T) {

	graphInitial, _ := getRandomGraph(10)

	// attempt to encode edges of graph

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(graphInitial)
	if err != nil {
		log.Fatal("encode error", err)
	}
	fmt.Println("Output from encoding:")
	fmt.Println(buffer.Bytes())

	dec := gob.NewDecoder(&buffer)

	var decodedGraph lib.Graph

	dec.Decode(&decodedGraph)

	if !reflect.DeepEqual(decodedGraph, graphInitial) {
		t.Errorf("Graphs not equal: %v, %v ", graphInitial, decodedGraph)
	}

}
