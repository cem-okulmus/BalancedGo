package main

import (
	"log"
	"math"
	"math/rand"
	"sort"
)

// Heuristics to order the edges by

func getMSCOrder(edges []Edge) []Edge {
	if len(edges) <= 1 {
		return edges
	}
	var selected []Edge
	chosen := make([]bool, len(edges))

	//randomly select last edge in the ordering
	i := rand.Intn(len(edges))
	chosen[i] = true
	selected = append(selected, edges[i])

	for len(selected) < len(edges) {
		var candidates []int
		maxcard := 0

		for current := range edges {
			current_card := edges[current].numNeighboursOrder(edges, chosen)
			if !chosen[current] && current_card >= maxcard {
				if current_card > maxcard {
					candidates = []int{}
					maxcard = current_card
				}

				candidates = append(candidates, current)
			}
		}

		//randomly select one of the edges with equal connectivity
		next_in_order := candidates[rand.Intn(len(candidates))]
		//next_in_order := candidates[0]

		selected = append(selected, edges[next_in_order])
		chosen[next_in_order] = true
	}

	// //reverse order of selected
	// for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
	// 	selected[i], selected[j] = selected[j], selected[i]
	// }

	return selected
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

	for _, n := range e.vertices {
		for _, m := range e.vertices {
			n_index, _ := order[n]
			m_index, _ := order[m]
			if n_index != m_index {
				output[n_index][m_index] = 1
			}
		}
	}

	return output
}

func getMinDistances(vertices []int, edges []Edge) ([][]int, map[int]int) {
	var output [][]int
	order := make(map[int]int)

	log.Println("Vertices: ", len(vertices))

	for i, n := range vertices {
		order[n] = i
	}

	row := make([]int, len(vertices))
	for j := 0; j < len(vertices); j++ {
		row[j] = math.MaxInt64
	}

	for j := 0; j < len(vertices); j++ {
		new_row := make([]int, len(vertices))
		copy(new_row, row)
		output = append(output, new_row)
	}

	for _, e := range edges {
		output = addEdgeDistances(order, output, e)
	}

	for j := 0; j < len(edges); j++ {
		changed := false
		for k := range vertices {
			for l := range vertices {
				for m := range vertices {
					if isInf(output[k][l]) || isInf(output[l][m]) {
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

func getMaxSepOrder(edges []Edge) []Edge {
	if len(edges) <= 1 {
		return edges
	}
	vertices := Vertices(edges)
	weights := make([]int, len(edges))

	initialDiff, order := getMinDistances(vertices, edges)

	for i, e := range edges {
		edges_wihout_e := diffEdges(edges, e)
		newDiff, _ := getMinDistances(vertices, edges_wihout_e)
		newDiffPrep := addEdgeDistances(order, newDiff, e)
		weights[i] = diffDistances(initialDiff, newDiffPrep)
	}

	sort.Slice(edges, func(i, j int) bool { return weights[i] > weights[j] })

	return edges
}

func edgeDegree(edges []Edge, edge Edge) int {
	var output int

	for _, v := range edge.vertices {
		output = output + getDegree(edges, v)
	}

	return output - len(edge.vertices)
}

func getDegreeOrder(edges []Edge) []Edge {
	if len(edges) <= 1 {
		return edges
	}
	sort.Slice(edges, func(i, j int) bool { return edgeDegree(edges, edges[i]) > edgeDegree(edges, edges[j]) })
	return edges
}
