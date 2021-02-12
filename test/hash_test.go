package tests

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
	"github.com/google/go-cmp/cmp"
)

func check(e error) {
	if e != nil {
		panic(e)
	}

}

// TestIntHash provides a basic test for hashes of integers
func TestIntHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 1000; x++ {
		arity := r.Intn(100) + 1

		var vertices []int

		for i := 0; i < arity; i++ {
			vertices = append(vertices, r.Intn(1000)+i)
		}

		hash1 := lib.IntHash(vertices)

		r.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

		hash2 := lib.IntHash(vertices)

		newVal := r.Intn(100) + len(vertices)
		different := vertices[len(vertices)/2] != newVal
		vertices[len(vertices)/2] = newVal

		hash3 := lib.IntHash(vertices)

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if different && hash3 == hash2 {

			fmt.Println("vertex", vertices)
			t.Errorf("hash collision 1")
		}

	}

	// Collision Test
	// generate two different integers and see if their hashs collide

	for x := 0; x < 1000; x++ {

		arity := r.Intn(20) + 1

		var temp1 []int
		var temp2 []int

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, r.Intn(100))
		}

		for i := 0; i < arity; i++ {
			temp2 = append(temp2, r.Intn(100))
		}


		hash1 := lib.IntHash(temp1)
		hash2 := lib.IntHash(temp2)


		temp1 = lib.RemoveDuplicates(temp1)
		temp2 = lib.RemoveDuplicates(temp2)

		if cmp.Equal(temp1, temp2) {
			continue
		}


		if hash1 == hash2 {
			fmt.Println("Collision", temp1, temp2)
			t.Errorf("hash collision 2")
		}
	}

}

// BenchmarkSeparator uses a specific hypergraph
func BenchmarkSeparator(b *testing.B) {

	// Get the data
	resp, err := http.Get("http://hyperbench.dbai.tuwien.ac.at/download/hypergraph/655")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	parsedGraph, _ := lib.GetGraph(buf.String())

	pred := lib.BalancedCheck{}

	for i := 0; i < b.N; i++ {

		var edges []int

		k := 20

		for i := 0; i < k; i++ {
			edges = append(edges, r.Intn(parsedGraph.Edges.Len()))
		}

		sep := lib.GetSubset(parsedGraph.Edges, edges)

		pred.Check(&parsedGraph, &sep, 1)
	}
}

// TestEdgeHash tests the hash function of Edge against collisions and stability under permutation
func TestEdgeHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 100; x++ {
		arity := r.Intn(100) + 1

		var vertices []int

		name := r.Intn(1000)

		for i := 0; i < arity; i++ {
			vertices = append(vertices, r.Intn(1000)+i)
		}

		edge := lib.Edge{Name: name, Vertices: vertices}

		hash1 := edge.Hash()

		r.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

		edge2 := lib.Edge{Name: name, Vertices: vertices}

		hash2 := edge2.Hash()

		newVal := r.Intn(100) + len(vertices)
		different := vertices[len(vertices)/2] != newVal
		vertices[len(vertices)/2] = newVal

		edge3 := lib.Edge{Name: name, Vertices: vertices}

		hash3 := edge3.Hash()

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if different && hash3 == hash2 {

			fmt.Println("vertex", vertices)
			t.Errorf("hash collision 3")
		}

	}

	// Collision Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		arity := r.Intn(20) + 1

		var temp1 []int
		var temp2 []int

		for i := 0; i < arity; i++ {
			temp1 = append(temp1, r.Intn(100))
		}

		for i := 0; i < arity; i++ {
			temp2 = append(temp2, r.Intn(100))
		}

		if cmp.Equal(temp1, temp2) {
			continue
		}

		edge := lib.Edge{Name: 0, Vertices: lib.RemoveDuplicates(temp1)}

		edge2 := lib.Edge{Name: 0, Vertices: lib.RemoveDuplicates(temp2)}

		hash1 := edge.Hash()

		hash2 := edge2.Hash()

		if hash1 == hash2 {
			fmt.Println("Collision", temp1, temp2)
			t.Errorf("hash collision 4")
		}
	}

}

// TestEdgesHash tests the hash function of Edges against collisions and stability under permutation
func TestEdgesHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for x := 0; x < 100; x++ {

		length := r.Intn(20) + 1
		var temp []lib.Edge

		for c := 0; c < length; c++ {

			arity := r.Intn(100) + 1

			var vertices []int

			name := r.Intn(1000)

			for i := 0; i < arity; i++ {
				vertices = append(vertices, r.Intn(1000)+i)
			}

			edge := lib.Edge{Name: name, Vertices: vertices}

			temp = append(temp, edge)

		}

		edges := lib.NewEdges(temp)

		hash1 := edges.Hash()

		r.Shuffle(len(temp), func(i, j int) { temp[i], temp[j] = temp[j], temp[i] })

		edges = lib.NewEdges(temp)

		hash2 := edges.Hash()

		index := r.Intn(len(temp))
		index2 := r.Intn(len(temp[index].Vertices))
		temp[index].Vertices[index2] = temp[index].Vertices[index2] + 1

		edges = lib.NewEdges(temp)

		hash3 := edges.Hash()

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if hash3 == hash2 {
			t.Errorf("hash collision 5")
		}

	}

	// Collision Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		length := r.Intn(20) + 1

		var temp []lib.Edge
		var temp2 []lib.Edge

		for j := 0; j < length; j++ {

			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp = append(temp, lib.Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		for j := 0; j < length; j++ {

			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp2 = append(temp2, lib.Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		if cmp.Equal(temp, temp2) {
			continue
		}

		edges := lib.NewEdges(temp)

		edges2 := lib.NewEdges(temp2)

		hash1 := edges.Hash()

		hash2 := edges2.Hash()

		if hash1 == hash2 {
			fmt.Println("Collision", temp, temp2)
			t.Errorf("hash collision 6")
		}
	}

}

// TestGraphHash tests the hash function of Graph against collisions and stability under permutation
func TestGraphHash(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	var Sp []lib.Edges

	lengthSpeciale := r.Intn(20) + 1

	lengthK := r.Intn(5) + 1

	for c := 0; c < lengthSpeciale; c++ {

		var slice []lib.Edge

		for o := 0; o < lengthK; o++ {

			arity := r.Intn(100) + 1

			var vertices []int

			for i := 0; i < arity; i++ {
				vertices = append(vertices, r.Intn(1000)+i)
			}

			slice = append(slice, lib.Edge{Vertices: vertices})
		}

		Sp = append(Sp, lib.NewEdges(slice))

	}

	for x := 0; x < 100; x++ {

		length := r.Intn(20) + 1
		var temp []lib.Edge

		for c := 0; c < length; c++ {

			arity := r.Intn(100) + 1

			var vertices []int

			name := r.Intn(1000)

			for i := 0; i < arity; i++ {
				vertices = append(vertices, r.Intn(1000)+i)
			}

			edge := lib.Edge{Name: name, Vertices: vertices}

			temp = append(temp, edge)

		}

		edges := lib.NewEdges(temp)

		graph := lib.Graph{Edges: edges, Special: Sp}

		hash1 := graph.Hash()

		r.Shuffle(len(temp), func(i, j int) { temp[i], temp[j] = temp[j], temp[i] })

		edges = lib.NewEdges(temp)
		graph2 := lib.Graph{Edges: edges, Special: Sp}

		hash2 := graph2.Hash()

		index := r.Intn(len(temp))
		index2 := r.Intn(len(temp[index].Vertices))
		temp[index].Vertices[index2] = temp[index].Vertices[index2] + 1

		edges = lib.NewEdges(temp)
		graph3 := lib.Graph{Edges: edges, Special: Sp}

		hash3 := graph3.Hash()

		if hash1 != hash2 {
			t.Errorf("hash not stable under permutation")
		}

		if hash3 == hash2 {
			t.Errorf("hash collision 7")
		}

	}

	// Collision Test
	// generate two different edges and see if their hashs collide

	for x := 0; x < 1000; x++ {

		length := r.Intn(20) + 1

		var temp []lib.Edge
		var temp2 []lib.Edge

		for j := 0; j < length; j++ {

			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp = append(temp, lib.Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		for j := 0; j < length; j++ {

			arity := r.Intn(20) + 1

			var temp1a []int

			for i := 0; i < arity; i++ {
				temp1a = append(temp1a, r.Intn(100))
			}
			temp2 = append(temp2, lib.Edge{Name: r.Intn(100) + 1, Vertices: temp1a})

		}

		if cmp.Equal(temp, temp2) {
			continue
		}

		edges := lib.NewEdges(temp)

		graph4 := lib.Graph{Edges: edges, Special: Sp}
		edges2 := lib.NewEdges(temp2)
		graph5 := lib.Graph{Edges: edges2, Special: Sp}

		hash1 := graph4.Hash()

		hash2 := graph5.Hash()

		if hash1 == hash2 {
			fmt.Println("Collision", temp, temp2)
			t.Errorf("hash collision 9")
		}
	}

}
