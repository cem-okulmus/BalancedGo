package lib

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

// Heuristics to order the edges by

func GetMSCOrder(edges Edges) Edges {

	rand.Seed(time.Now().UTC().UnixNano())
	if edges.Len() <= 1 {
		return edges
	}
	var selected []Edge
	chosen := make([]bool, edges.Len())

	//randomly select last edge in the ordering
	i := rand.Intn(edges.Len())
	chosen[i] = true
	selected = append(selected, edges.Slice()[i])

	for len(selected) < edges.Len() {
		var candidates []int
		maxcard := 0

		for current := range edges.Slice() {
			currentCard := edges.Slice()[current].numNeighboursOrder(edges, chosen)
			if !chosen[current] && currentCard >= maxcard {
				if currentCard > maxcard {
					candidates = []int{}
					maxcard = currentCard
				}

				candidates = append(candidates, current)
			}
		}

		//randomly select one of the edges with equal connectivity
		nextInOrder := candidates[rand.Intn(len(candidates))]
		//nextInOrder := candidates[0]

		selected = append(selected, edges.Slice()[nextInOrder])
		chosen[nextInOrder] = true
	}

	// //reverse order of selected
	// for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
	//  selected[i], selected[j] = selected[j], selected[i]
	// }

	return NewEdges(selected)
}

//Order the edges by how much  they increase shortest paths within the hypergraph

//basic Floyd-Warschall (using the primal graph)

func order(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func isInf(a int) bool {
	return a == math.MaxInt64
}

func addEdgeDistances(order map[int]int, output [][]int, e Edge) [][]int {

	for _, n := range e.Vertices {
		for _, m := range e.Vertices {
			n_index, _ := order[n]
			m_index, _ := order[m]
			if n_index != m_index {
				output[n_index][m_index] = 1
			}
		}
	}

	return output
}

func getMinDistances(vertices []int, edges Edges) ([][]int, map[int]int) {
	var output [][]int
	order := make(map[int]int)

	for i, n := range vertices {
		order[n] = i
	}

	row := make([]int, len(vertices))
	for j := 0; j < len(vertices); j++ {
		row[j] = math.MaxInt64
	}

	for j := 0; j < len(vertices); j++ {
		newRow := make([]int, len(vertices))
		copy(newRow, row)
		output = append(output, newRow)
	}

	for _, e := range edges.Slice() {
		output = addEdgeDistances(order, output, e)
	}

	for j := 0; j < edges.Len(); j++ {
		changed := false
		for k := range vertices {
			for l := range vertices {
				if isInf(output[k][l]) {
					continue
				}
				for m := range vertices {
					if isInf(output[l][m]) {
						continue
					}
					newdist := output[k][l] + output[l][m]
					if output[k][m] > newdist {
						output[k][m] = newdist
						changed = true
					}
				}
			}
		}
		if !changed {
			break
		}

	}

	return output, order
}

//  weight of each edge = (sum of path disconnected)*SepWeight  +  (sum of each path made longer * diff)
func diffDistances(old, new [][]int) int {
	var output int

	SepWeight := len(old) * len(old)

	for j := 0; j < len(old); j++ {
		for i := 0; i < len(old[j]); i++ {
			if isInf(old[j][i]) && !isInf(new[j][i]) { // disconnected a path
				output = output + SepWeight
			} else if !isInf(old[j][i]) && !isInf(new[j][i]) { // check if parth shortened
				diff := old[j][i] - new[j][i]
				output = output + diff
			}
		}
	}

	return output
}

func GetMaxSepOrder(edges Edges) Edges {
	if edges.Len() <= 1 {
		return edges
	}
	vertices := edges.Vertices()
	weights := make([]int, edges.Len())

	initialDiff, order := getMinDistances(vertices, edges)

	for i, e := range edges.Slice() {
		edgesWihoutE := DiffEdges(edges, e)
		newDiff, _ := getMinDistances(vertices, edgesWihoutE)
		newDiffPrep := addEdgeDistances(order, newDiff, e)
		weights[i] = diffDistances(initialDiff, newDiffPrep)
	}

	sort.Slice(edges.Slice(), func(i, j int) bool { return weights[i] > weights[j] })

	return edges
}

func edgeVertexDegree(edges Edges, edge Edge) int {
	var output int

	for _, v := range edge.Vertices {
		output = output + getDegree(edges, v)
	}

	return output - len(edge.Vertices)
}

func GetDegreeOrder(edges Edges) Edges {
	if edges.Len() <= 1 {
		return edges
	}
	sort.Slice(edges.Slice(), func(i, j int) bool {
		return edgeVertexDegree(edges, edges.Slice()[i]) > edgeVertexDegree(edges, edges.Slice()[j])
	})
	return edges
}

func edgeDegree(edges Edges, edge Edge) int {
	output := 0

	for i := range edges.Slice() {
		if edges.Slice()[i].areNeighbours(edge) {
			output++
		}
	}

	return output
}

func GetEdgeDegreeOrder(edges Edges) Edges {
	if edges.Len() <= 1 {
		return edges
	}
	sort.Slice(edges.Slice(), func(i, j int) bool {
		return edgeDegree(edges, edges.Slice()[i]) > edgeDegree(edges, edges.Slice()[j])
	})
	return edges

}
