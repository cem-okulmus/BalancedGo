package lib

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestIntHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 1000; x++ {
		arity := rand.Intn(100) + 1

		var vertices []int

		for i := 0; i < arity; i++ {
			vertices = append(vertices, rand.Intn(1000)+i)
		}

		hash1 := IntHash(vertices)

		r.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

		hash2 := IntHash(vertices)

		newVal := rand.Intn(100) + len(vertices)
		different := vertices[len(vertices)/2] != newVal
		vertices[len(vertices)/2] = newVal

		hash3 := IntHash(vertices)

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if different && hash3 == hash2 {

			fmt.Println("vertex", vertices)
			t.Errorf("hash collision")
		}

	}

	// Collission Test
	// generate two different integers and see if their hashs collide

	for x := 0; x < 1000; x++ {

		// arity1 := rand.Intn(100) + 1
		arity := rand.Intn(20) + 1

		var temp1 []int
		var temp2 []int

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, rand.Intn(100))
		}

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, rand.Intn(100))
		}

		if reflect.DeepEqual(temp1, temp2) {
			continue
		}

		hash1 := IntHash(temp1)

		hash2 := IntHash(temp2)

		if hash1 == hash2 {
			fmt.Println("Collission", temp1, temp2)
			t.Errorf("hash collision")
		}
	}

}

func BenchmarkSeparator(b *testing.B) {

	dat, err := ioutil.ReadFile("/home/cem/Desktop/scripts/BalancedGo/hypergraphs/Nonogram-007-table.xml.hg")
	check(err)

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	parsedGraph, _ := GetGraph(string(dat))

	pred := BalancedCheck{}

	for i := 0; i < b.N; i++ {

		var edges []int

		k := 20

		for i := 0; i < k; i++ {
			edges = append(edges, r.Intn(parsedGraph.Edges.Len()))
		}

		sep := GetSubset(parsedGraph.Edges, edges)

		pred.Check(&parsedGraph, []Special{}, &sep, 1)
	}
}

func TestEdgeHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 100; x++ {
		arity := rand.Intn(100) + 1

		var vertices []int

		name := r.Intn(1000)

		for i := 0; i < arity; i++ {
			vertices = append(vertices, r.Intn(1000)+i)
		}

		edge := Edge{Name: name, Vertices: vertices}

		hash1 := edge.Hash()

		r.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

		edge2 := Edge{Name: name, Vertices: vertices}

		hash2 := edge2.Hash()

		newVal := r.Intn(100) + len(vertices)
		different := vertices[len(vertices)/2] != newVal
		vertices[len(vertices)/2] = newVal

		edge3 := Edge{Name: name, Vertices: vertices}

		hash3 := edge3.Hash()

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if different && hash3 == hash2 {

			fmt.Println("vertex", vertices)
			t.Errorf("hash collision")
		}

	}

	// Collission Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		// arity1 := rand.Intn(100) + 1
		arity := r.Intn(20) + 1

		var temp1 []int
		var temp2 []int

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, r.Intn(100))
		}

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, r.Intn(100))
		}

		if reflect.DeepEqual(temp1, temp2) {
			continue
		}

		edge := Edge{Name: 0, Vertices: temp1}

		edge2 := Edge{Name: 0, Vertices: temp2}

		hash1 := edge.Hash()

		hash2 := edge2.Hash()

		if hash1 == hash2 {
			fmt.Println("Collission", temp1, temp2)
			t.Errorf("hash collision")
		}
	}

}

func TestEdgesHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 100; x++ {

		length := r.Intn(20) + 1
		var temp []Edge

		for c := 0; c < length; c++ {

			arity := r.Intn(100) + 1

			var vertices []int

			name := r.Intn(1000)

			for i := 0; i < arity; i++ {
				vertices = append(vertices, r.Intn(1000)+i)
			}

			edge := Edge{Name: name, Vertices: vertices}

			temp = append(temp, edge)

		}

		edges := NewEdges(temp)

		hash1 := edges.Hash()

		r.Shuffle(len(temp), func(i, j int) { temp[i], temp[j] = temp[j], temp[i] })

		edges = NewEdges(temp)

		hash2 := edges.Hash()

		index := r.Intn(len(temp))
		index2 := r.Intn(len(temp[index].Vertices))
		temp[index].Vertices[index2] = temp[index].Vertices[index2] + 1

		edges = NewEdges(temp)

		hash3 := edges.Hash()

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if hash3 == hash2 {
			t.Errorf("hash collision")
		}

	}

	// Collission Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		length := r.Intn(20) + 1

		var temp []Edge
		var temp2 []Edge

		for j := 0; j < length; j++ {

			// arity1 := rand.Intn(100) + 1
			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp = append(temp, Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		for j := 0; j < length; j++ {

			// arity1 := rand.Intn(100) + 1
			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp2 = append(temp2, Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		if reflect.DeepEqual(temp, temp2) {
			continue
		}

		edges := NewEdges(temp)

		edges2 := NewEdges(temp2)

		hash1 := edges.Hash()

		hash2 := edges2.Hash()

		if hash1 == hash2 {
			fmt.Println("Collission", temp, temp2)
			t.Errorf("hash collision")
		}
	}

}

func TestEdgesExtendedHash(t *testing.T) {

	var Sp []Special

	lengthSpeciale := rand.Intn(20) + 1

	for c := 0; c < lengthSpeciale; c++ {

		arity := rand.Intn(100) + 1

		var vertices []int

		for i := 0; i < arity; i++ {
			vertices = append(vertices, rand.Intn(1000)+i)
		}

		special := Special{Vertices: vertices}

		Sp = append(Sp, special)

	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 100; x++ {

		length := r.Intn(20) + 1
		var temp []Edge

		for c := 0; c < length; c++ {

			arity := r.Intn(100) + 1

			var vertices []int

			name := r.Intn(1000)

			for i := 0; i < arity; i++ {
				vertices = append(vertices, r.Intn(1000)+i)
			}

			edge := Edge{Name: name, Vertices: vertices}

			temp = append(temp, edge)

		}

		edges := NewEdges(temp)

		hash1 := edges.HashExtended(Sp)

		r.Shuffle(len(temp), func(i, j int) { temp[i], temp[j] = temp[j], temp[i] })

		edges = NewEdges(temp)

		hash2 := edges.HashExtended(Sp)

		index := r.Intn(len(temp))
		index2 := r.Intn(len(temp[index].Vertices))
		temp[index].Vertices[index2] = temp[index].Vertices[index2] + 1

		edges = NewEdges(temp)

		hash3 := edges.HashExtended(Sp)

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if hash3 == hash2 {
			t.Errorf("hash collision")
		}

	}

	// Collission Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		length := r.Intn(20) + 1

		var temp []Edge
		var temp2 []Edge

		for j := 0; j < length; j++ {

			// arity1 := r.Intn(100) + 1
			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp = append(temp, Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		for j := 0; j < length; j++ {

			// arity1 := r.Intn(100) + 1
			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp2 = append(temp2, Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		if reflect.DeepEqual(temp, temp2) {
			continue
		}

		edges := NewEdges(temp)

		edges2 := NewEdges(temp2)

		hash1 := edges.HashExtended(Sp)

		hash2 := edges2.HashExtended(Sp)

		if hash1 == hash2 {
			fmt.Println("Collission", temp, temp2)
			t.Errorf("hash collision")
		}
	}

}
