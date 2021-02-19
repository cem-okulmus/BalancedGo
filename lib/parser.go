package lib

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

var m map[int]string // stores the encoding of vertices for last file parsed (bit of a hack)
var mutex = sync.RWMutex{}
var encode int // stores the encoding of the highest int used

type parseEdge struct {
	Name     string   ` @(Number|Ident|String)`
	Vertices []string `"(" ( @(Number|Ident|String)  ","? )* ")"`
}

// ParseGraph contains data used to parse a graph, potentially useful for testing
type ParseGraph struct {
	Edges    []parseEdge `( @@ ","?)* (".")?`
	Encoding map[string]int
}

// GetGraph parses a string in HyperBench format into a graph
func GetGraph(s string) (Graph, ParseGraph) {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (digit| alpha | "_") { Punct |  "_" | alpha | digit } .
    String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
    Number = [ "-" | "+" ] ("." | digit) { "." | digit } .
    Punct = "." | ";"  | "_" | ":" | "!" | "?" | "\\" | "/" | "=" | "[" | "]" | "'" | "$" | "<" | ">" | "-" | "+" | "~" | "@" | "*" | "\""  .
    Parenthesis = "(" | ")"  | "," .
    Whitespace = " " | "\t" | "\n" | "\r" .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .
    any = "\u0000"…"\uffff" .
    `))

	var parser = participle.MustBuild(&ParseGraph{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	var output Graph
	var edges []Edge
	pgraph := ParseGraph{}
	err := parser.ParseString(s, &pgraph)
	if err != nil {
		fmt.Println("Couldn't parse input: ")
		panic(err)
	}
	encoding := make(map[int]string)
	encode = 1 // initialize to 1

	pgraph.Encoding = make(map[string]int)
	//fix first numbers for edge names
	for _, e := range pgraph.Edges {
		_, ok := pgraph.Encoding[e.Name]
		if ok {
			log.Panicln("Edge names not unique, not a vald hypergraph!")
		}

		pgraph.Encoding[e.Name] = encode
		encoding[encode] = e.Name
		encode++
	}
	for _, e := range pgraph.Edges {
		var outputEdges []int
		for _, n := range e.Vertices {
			i, ok := pgraph.Encoding[n]
			if ok {
				outputEdges = append(outputEdges, i)
			} else {
				pgraph.Encoding[n] = encode
				encoding[encode] = n
				outputEdges = append(outputEdges, encode)
				encode++
			}
		}
		edges = append(edges, Edge{Name: pgraph.Encoding[e.Name], Vertices: outputEdges})
	}
	output.Edges = NewEdges(edges)
	m = encoding
	return output, pgraph
}

// GetEdge can be used parse additional hyperedges. Useful for testing purposes
func (p *ParseGraph) GetEdge(input string) Edge {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (alpha | "_") { "_" | alpha | digit | stuff } .
    Number = ("." | digit | "_"){"." | digit | stuff } .
    Whitespace = " " | "\t" | "\n" | "\r" .
    stuff = ":" | "@" | ";" | "-" | "_" .
    Punct = "!"…"/"  .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&parseEdge{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	pEdge := parseEdge{}
	parser.ParseString(input, &pEdge)
	var vertices []int
	for _, v := range pEdge.Vertices {
		val, ok := p.Encoding[v]
		if ok {
			vertices = append(vertices, val)
		} else {
			p.Encoding[v] = encode
			m[encode] = v
			vertices = append(vertices, encode)
			encode++
		}
	}
	m[encode] = pEdge.Name
	encode++
	return Edge{Vertices: vertices, Name: encode - 1}
}

// Implement PACE 2019 format

type parseEdgePACE struct {
	Name     int   ` @Number`
	Vertices []int ` ( @Number   )* "\n" `
}

type parseGraphPACEInfo struct {
	Vertices int `"p htd" @(Number) `
	Edges    int `@(Number) "\n"`
}

type parseGraphPACE struct {
	Info  parseGraphPACEInfo `@@`
	Edges []parseEdgePACE    `(@@) *`
	m     map[int]int
}

// GetGraphPACE parses a string in PACE 2019 format into a graph
func GetGraphPACE(s string) Graph {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("c" | "//") { "\u0000"…"\uffff"-"\n" } Newline.
    Begin = "p htd" .
    Number = ("." | digit | "_"){"." | digit | stuff } .
    Whitespace = " " | "\t" | "\r" .
    stuff = ":" | "@" | ";" | "-" | "_" .
    Punct = "!"…"/"  .
    Newline = "\n" .

    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&parseGraphPACE{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	var output Graph
	var edges []Edge
	pgraph := parseGraphPACE{}
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

	return output
}

func extractEdge(edges []Edge, edge int) Edge {
	for i := range edges {
		if edges[i].Name == edge {
			return edges[i]
		}
	}

	return Edge{} // return empty struct for fail case
}

func (n Node) attachChild(target int, child Node) Node {
	if n.num == target {
		child.parPointer = &n
		n.Children = append(n.Children, child)
		return n
	}

	for i := range n.Children {
		out := n.Children[i].attachChild(target, child)

		if !reflect.DeepEqual(out, Node{}) {
			out.parPointer = &n
			n.Children[i] = out
			return n
		}
	}

	return Node{}
}

type arc struct {
	Source int
	Target int
}

type parseGMLValue struct {
	FlatVal string       ` @(Ident | Number) | "\"" @(Number | Ident | Punct)* "\""    `
	List    parseGMLList `| "[" @@ "]"`
}

type parseGMLListEntry struct {
	Key   string        ` @(Ident|Number) `
	Value parseGMLValue ` @@ `
}

type parseGMLList struct {
	Entries []parseGMLListEntry `( @@ )*`
}

type parseGML struct {
	GML parseGMLList `@@`
}

// GetDecompGML can parse an input string in GML format to produce a decomp
func GetDecompGML(input string, graph Graph, encoding map[string]int) Decomp {

	graphLexer := lexer.Must(ebnf.New(`
    Quote = "\"" .
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (alpha | "_") { "_" | alpha | digit | stuff } .
    Number = ("." | digit | "_"){"." | digit | stuff } .
    Whitespace = " " | "\t" | "\n" | "\r" .
    stuff = ":" | "@" | ";" | "-" | "_" .
    Punct = "!"…"}"-"\""  .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .`))

	var parser = participle.MustBuild(&parseGML{}, participle.UseLookahead(1), participle.Lexer(graphLexer), participle.Elide("Comment", "Whitespace"))
	pDecomp := parseGML{}
	err := parser.ParseString(input, &pDecomp)
	if err != nil {
		fmt.Println("Couldn't parse input: ")
		panic(err)
	}

	// Check if GML file consists of single graph node
	if len(pDecomp.GML.Entries) != 1 && pDecomp.GML.Entries[0].Key != "graph" {
		log.Panicln("Valid GML file, but does not contain a unique graph element")
	}

	var graphEntry parseGMLListEntry

	graphEntry = pDecomp.GML.Entries[0]

	var arcs []arc
	var nodes []Node

	IDtoIndex := make(map[int]int)

	edges := graph.Edges.Slice()

	for _, n := range graphEntry.Value.List.Entries {
		switch n.Key {
		case "node":
			var node Node

			var nodeLabels map[string]string
			nodeLabels = make(map[string]string)

			for _, e := range n.Value.List.Entries {
				if e.Value.FlatVal != "" {
					nodeLabels[e.Key] = e.Value.FlatVal
				}
			}

			// check for necessary fields, id and label
			if _, ok := nodeLabels["id"]; !ok {
				log.Println("Node without id present in GML file.")
			}
			if _, ok := nodeLabels["label"]; !ok {
				log.Println("Node without label present in GML file.")
			}

			// extract edge cover and bag from label

			reCover := regexp.MustCompile(`{(.*)}{.*}`)
			reBag := regexp.MustCompile(`{.*}{(.*)}`)

			matchCover := reCover.FindStringSubmatch(nodeLabels["label"])
			matchBag := reBag.FindStringSubmatch(nodeLabels["label"])
			if matchCover == nil || matchBag == nil {
				log.Panicln("Label of node ", nodeLabels["id"], " not properly formatted: ", nodeLabels["label"], ".")
			}

			var bag []int
			for _, v := range strings.Split(matchBag[1], ",") {
				bag = append(bag, encoding[v])
			}

			var cover []Edge
			for _, e := range strings.Split(matchCover[1], ",") {
				out := extractEdge(edges, encoding[e])
				if reflect.DeepEqual(out, Edge{}) {
					log.Panicln("Can't find edge ", e)
				}
				cover = append(cover, out)
			}

			node.num, _ = strconv.Atoi(nodeLabels["id"])

			encode = max(node.num, encode) + 1 // ensure encode will never collide with num values in parsed GMl

			node.Bag = bag
			node.Cover = NewEdges(cover)

			nodes = append(nodes, node)

			IDtoIndex[node.num] = len(nodes) - 1
		case "edge":
			var Arc arc

			var arcLabels map[string]string
			arcLabels = make(map[string]string)

			for _, e := range n.Value.List.Entries {
				if e.Value.FlatVal != "" {
					arcLabels[e.Key] = e.Value.FlatVal
				}
			}

			// check for necessary fields, id and label
			if _, ok := arcLabels["source"]; !ok {
				log.Println("Edge without source present in GML file.")
			}
			if _, ok := arcLabels["target"]; !ok {
				log.Println("Edge without target present in GML file.")
			}

			Arc.Source, _ = strconv.Atoi(arcLabels["source"])
			Arc.Target, _ = strconv.Atoi(arcLabels["target"])

			arcs = append(arcs, Arc)
		}
	}

	var root int
	if len(arcs) != 0 {
		root = arcs[0].Source
	} else {
		root = nodes[0].num
	}

	changed := true

	for changed {
		changed = false
		for _, arc := range arcs {
			if arc.Target == root { //determine global ancestor
				root = arc.Source
				changed = true
			}
		}
	}

	for _, arc := range arcs {
		source := nodes[IDtoIndex[arc.Source]]
		target := nodes[IDtoIndex[arc.Target]]

		nodes[IDtoIndex[arc.Source]] = source.attachChild(arc.Source, target)

		IDtoIndex[arc.Target] = IDtoIndex[arc.Source] // update reference to target
	}

	return Decomp{Graph: graph, Root: nodes[IDtoIndex[root]]}
}
