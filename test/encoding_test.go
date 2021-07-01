package tests

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"testing"
)

func TestEncoding(t *testing.T) {

	graphInitial, _ := getRandomGraph(10)

	// attempt to encode edges of graph

	var Encodebuffer bytes.Buffer
	enc := gob.NewEncoder(&Encodebuffer)

	err := enc.Encode(graphInitial)
	if err != nil {
		log.Fatal("encode error", err)
	}
	fmt.Println("Output from encoding:")
	fmt.Println(Encodebuffer.Bytes())
}
