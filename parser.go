package main

import "github.com/alecthomas/participle"

var m map[int]string // stores the encoding of vertices for last file parsed (bit of a hack)
var encode int       // stores the encoding of the highest int used

type ParseEdge struct {
	Name     string   `(Int)? @Ident`
	Vertices []string `"(" ( @(Ident|Int)  ","? )* ")"`
}

type ParseGraph struct {
	Edges []ParseEdge `( @@ ","?)*`
	m     map[string]int
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
		pgraph.m[e.Name] = encode
		encode++
	}
	for _, e := range pgraph.Edges {
		var outputEdges []int
		for _, n := range e.Vertices {
			i, ok := pgraph.m[n]
			if ok {
				outputEdges = append(outputEdges, i)
			} else {
				pgraph.m[n] = encode
				encoding[encode] = n
				outputEdges = append(outputEdges, encode)
				encode++
			}
		}
		output.edges = append(output.edges, Edge{name: e.Name, vertices: outputEdges})
	}
	m = encoding
	return output
}
