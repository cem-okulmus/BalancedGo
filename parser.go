package main

import "github.com/alecthomas/participle"

type ParseEdge struct {
	Name     string   `(Int)? @Ident`
	Vertices []string `"(" ( @(Ident|Int)  ","? )* ")"`
}

type ParseGraph struct {
	Edges []ParseEdge `( @@ ","?)*`
	m     map[string]int
	count int
}

var parser = participle.MustBuild(&ParseGraph{}, participle.UseLookahead(1))

func getGraph(s string) Graph {
	var output Graph
	pgraph := ParseGraph{}
	parser.ParseString(s, &pgraph)
	encoding := make(map[int]string)

	pgraph.m = make(map[string]int)
	//fix first numbers for edge names
	for _, e := range pgraph.Edges {
		pgraph.m[e.Name] = pgraph.count
		pgraph.count++
	}
	for _, e := range pgraph.Edges {
		var outputEdges []int
		for _, n := range e.Vertices {
			i, ok := pgraph.m[n]
			if ok {
				outputEdges = append(outputEdges, i)
			} else {
				pgraph.m[n] = pgraph.count
				encoding[pgraph.count] = n
				outputEdges = append(outputEdges, pgraph.count)
				pgraph.count++
			}
		}
		output.edges = append(output.edges, Edge{vertices: outputEdges})
	}
	m = encoding
	return output
}
