package lib

import (
	"github.com/alecthomas/participle"
)

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

func GetGraph(s string) (Graph, ParseGraph) {
	var parser = participle.MustBuild(&ParseGraph{}, participle.UseLookahead(1))
	var output Graph
	pgraph := ParseGraph{}
	parser.ParseString(s, &pgraph)
	encoding := make(map[int]string)
	encode = 1 // initialize to 1

	pgraph.m = make(map[string]int)
	//fix first numbers for edge names
	for _, e := range pgraph.Edges {
		pgraph.m[e.Name] = encode
		encoding[encode] = e.Name
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
		output.Edges.Append(Edge{Name: pgraph.m[e.Name], Vertices: outputEdges})
	}
	m = encoding
	return output, pgraph
}

func (p *ParseGraph) GetEdge(input string) Edge {

	var parser = participle.MustBuild(&ParseEdge{}, participle.UseLookahead(1))
	pEdge := ParseEdge{}
	parser.ParseString(input, &pEdge)
	var vertices []int
	for _, v := range pEdge.Vertices {
		val, ok := p.m[v]
		if ok {
			vertices = append(vertices, val)
		} else {
			p.m[v] = encode
			m[encode] = v
			vertices = append(vertices, encode)
			encode++
		}
	}
	m[encode] = pEdge.Name
	encode++
	return Edge{Vertices: vertices, Name: encode - 1}
}
