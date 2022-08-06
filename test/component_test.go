package tests

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/cem-okulmus/BalancedGo/lib"
	"github.com/cem-okulmus/disjoint"
)

func connected(g lib.Graph) bool {

	var vertices = make(map[int]*disjoint.Element)
	var comps = make(map[*disjoint.Element][]lib.Edge)

	//  Set up the disjoint sets for each node
	for _, i := range g.Vertices() {
		vertices[i] = disjoint.NewElement()
	}

	// Merge together the connected components
	for k := range g.Edges.Slice() {
		for i := 0; i < len(g.Edges.Slice()[k].Vertices); i++ {
			for j := i + 1; j < len(g.Edges.Slice()[k].Vertices); j++ {
				disjoint.Union(vertices[g.Edges.Slice()[k].Vertices[i]], vertices[g.Edges.Slice()[k].Vertices[j]])
				// j = i-1
				break
			}
		}
	}

	for i := range g.Edges.Slice() {
		vertexRep := g.Edges.Slice()[0].Vertices[0] // pick arbitrary vertex of this edge

		slice, ok := comps[vertices[vertexRep].Find()]
		if !ok {
			newslice := make([]lib.Edge, 0, g.Edges.Len())
			comps[vertices[vertexRep].Find()] = newslice
			slice = newslice
		}

		comps[vertices[vertexRep].Find()] = append(slice, g.Edges.Slice()[i])
	}

	return len(comps) == 1 // graph is connected iff there's only one connected component in it
}

// TestComponent makes sure the calculation of connected components works. This is done by  generating a random instance of a graph and a separator, and making sure the produced components fulfill th properties of being components of the seperator.
func TestComponents(t *testing.T) {

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	graphInitial, _ := getRandomGraph(30)

	width := r.Intn(6) + 1
	sep := getRandomSep(graphInitial, width)
	var Vertices = make(map[int]*disjoint.Element)

	// get components
	comps, _, _ := graphInitial.GetComponents(sep, Vertices)

	// number of components must be >1

	if len(comps) == 0 && !lib.Subset(graphInitial.Vertices(), sep.Vertices()) {
		t.Error(" Compnent calculation error: empty number of components")
	}

	// Edges of sep must not occur in any of its components

	for _, c := range comps {
		for _, e := range c.Edges.Slice() {
			for _, s := range sep.Slice() {
				if reflect.DeepEqual(e, s) {
					t.Error("Component calculation error: contains edge of seperator")
				}
			}
		}
	}

	//  Check if components are vertex distinct, save for intersection with sep
	for i, c := range comps {
		for j, c2 := range comps {
			if i == j {
				continue
			}

			intersection := lib.Inter(c.Vertices(), c2.Vertices())

			if !lib.Subset(intersection, sep.Vertices()) {
				t.Error("Component calculation error: intersection components")
			}

		}
	}

	// Check if all components are connected

	for _, c := range comps {
		if !connected(c) {
			t.Error("Component calculation error: found component that's not connected")
		}
	}

}
