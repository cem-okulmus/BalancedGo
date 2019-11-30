package lib

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

var m map[int]string // stores the encoding of vertices for last file parsed (bit of a hack)
var mutex = sync.RWMutex{}
var encode int // stores the encoding of the highest int used

type ParseEdge struct {
	Name     string   ` @(Ident|Number)`
	Vertices []string `"(" ( @(Ident|Number)  ","? )* ")"`
}

type ParseGraph struct {
	Edges []ParseEdge `( @@ ","?)*`
	m     map[string]int
}

// Implement PACE 2019 format
//TODO Look up how to match multiple values in same line to different fields

type ParseEdgePACE struct {
	Name     int   ` @Number`
	Vertices []int ` ( @Number   )* "\n" `
}

type ParseGraphPACEInfo struct {
	Vertices int `"p htd":Begin @(Number) `
	Edges    int `@(Number) "\n"`
}

type ParseGraphPACE struct {
	Info  ParseGraphPACEInfo `@@`
	Edges []ParseEdgePACE    `(@@) *`
	m     map[int]int
}

func GetGraph(s string) (Graph, ParseGraph) {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (alpha | "_") { "_" | alpha | digit | stuff } .
    Number = ("." | digit) {"." | digit} .
    Whitespace = " " | "\t" | "\n" | "\r" .
    stuff = ":" | "@" | ";" | "-" .
    Punct = "!"…"/"  .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&ParseGraph{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	var output Graph
	var edges []Edge
	pgraph := ParseGraph{}
	parser.ParseString(s, &pgraph)
	// if err != nil {
	// 	fmt.Println("Couldn't parse input: ")
	// 	panic(err)
	// }
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
		edges = append(edges, Edge{Name: pgraph.m[e.Name], Vertices: outputEdges})
	}
	output.Edges = NewEdges(edges)
	m = encoding
	return output, pgraph
}

func (p *ParseGraph) GetEdge(input string) Edge {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (alpha | "_") { "_" | alpha | digit | stuff } .
    Number = ("." | digit) {"." | digit} .
    Whitespace = " " | "\t" | "\n" | "\r" .
    stuff = ":" | "@" | ";" | "-" .
    Punct = "!"…"/"  .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&ParseEdge{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
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

func GetGraphPACE(s string) Graph {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("c" | "//") { "\u0000"…"\uffff"-"\n" } Newline.
    Begin = "p htd" .
    Number = ("." | digit) {"." | digit} .
    Whitespace = " " | "\t" | "\r" .
    stuff = ":" | "@" | ";" | "-" .
    Punct = "!"…"/"  .
    Newline = "\n" .

    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&ParseGraphPACE{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	var output Graph
	var edges []Edge
	pgraph := ParseGraphPACE{}
	err := parser.ParseString(s, &pgraph)
	if err != nil {
		fmt.Println("Couldn't parse input: ")
		panic(err)
	}
	encode = 1 // initialize to 1

	encoding := make(map[int]string)
	pgraph.m = make(map[int]int)

	for _, e := range pgraph.Edges {
		encoding[encode] = "E" + strconv.Itoa(e.Name)
		pgraph.m[e.Name] = encode
		encode++
	}

	for _, e := range pgraph.Edges {
		var outputEdges []int
		for _, n := range e.Vertices {
			i, ok := pgraph.m[n+pgraph.Info.Edges]
			if ok {
				outputEdges = append(outputEdges, i)
			} else {
				pgraph.m[n+pgraph.Info.Edges] = encode
				encoding[encode] = "V" + strconv.Itoa(n)
				outputEdges = append(outputEdges, encode)
				encode++

			}
		}
		edges = append(edges, Edge{Name: pgraph.m[e.Name], Vertices: outputEdges})
	}

	m = encoding

	output.Edges = NewEdges(edges)

	// log.Println("Edges", pgraph.Info.Edges)
	// log.Println("Vertices", pgraph.Info.Vertices)

	// for _, e := range output.Edges.Slice() {
	// 	log.Println(e.FullString())
	// }

	// log.Panicln("")
	return output
}
