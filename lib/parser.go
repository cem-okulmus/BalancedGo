package lib

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

//hook for the json-iterator library
var json = jsoniter.ConfigCompatibleWithStandardLibrary

var m map[int]string // stores the encoding of vertices for last file parsed (bit of a hack)
var mutex = sync.RWMutex{}
var encode int // stores the encoding of the highest int used

type ParseEdge struct {
	Name     string   ` @(Number|Ident|String)`
	Vertices []string `"(" ( @(Number|Ident|String)  ","? )* ")"`
}

type ParseGraph struct {
	Edges    []ParseEdge `( @@ ","?)* (".")?`
	Encoding map[string]int
}

func GetGraph(s string) (Graph, ParseGraph) {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (digit| alpha | "_") { Punct |  "_" | alpha | digit } .
    String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
    Number = [ "-" | "+" ] ("." | digit) { "." | digit } .
    Punct = "." | ";"  | "_" | ":" | "!" | "?" | "\\" | "/" | "=" | "[" | "]" | "'" | "$" | "<" | ">" | "-" | "+" | "~" | "@" | "*" | "\""  .
    Paranthesis = "(" | ")"  | "," .
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

type NiceGraph struct {
	Width    int         `"⭐️" @(Number|Ident|String) `
	Edges    []ParseEdge `( @@ ","?)* (".")?`
	Encoding map[string]int
}

func GetNiceGraph(s string) (Graph, ParseGraph, int) {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (digit| alpha | "_") { Punct |  "_" | alpha | digit } .
    String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
    Number = [ "-" | "+" ] ("." | digit) { "." | digit } .
    Punct = "." | ";"  | "_" | ":" | "!" | "?" | "\\" | "/" | "=" | "[" | "]" | "'" | "$" | "<" | ">" | "-" | "+" | "~" | "@" | "*" | "\""  .
    Paranthesis = "(" | ")"  | "," .
    Whitespace = " " | "\t" | "\n" | "\r" .
    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .
    any = "\u0000"…"\uffff" .
    WidthPromise = "⭐️" .
    `))

	var parser = participle.MustBuild(&NiceGraph{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	var output Graph
	var edges []Edge
	pgraph := NiceGraph{}
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

	return output, ParseGraph{Edges: pgraph.Edges, Encoding: pgraph.Encoding}, pgraph.Width
}

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

	var parser = participle.MustBuild(&ParseEdge{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
		participle.Elide("Comment", "Whitespace"))
	pEdge := ParseEdge{}
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

func GetGraphPACE(s string) Graph {

	graphLexer := lexer.Must(ebnf.New(`
    Comment = ("c" | "//") { "\u0000"…"\uffff"-"\n" } Newline.
    Begin = "p htd" .
    Number = ("." | digit | "_"){"." | digit | stuff } .
    Whitespace = " " | "\t" | "\n" | "\r" .
    stuff = ":" | "@" | ";" | "-" | "_" .
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
	//  log.Println(e.FullString())
	// }

	// log.Panicln("")
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

// Implement Decomp parsing  (via JSON format)

type DecompJson struct {
	Root NodeJson
}

type NodeJson struct {
	Bag      []string
	Cover    []string
	Children []NodeJson
	Star     bool
}

func (d Decomp) IntoJson() DecompJson {
	var output DecompJson

	output.Root = d.Root.IntoJson()

	return output
}

func (n Node) IntoJson() NodeJson {
	var output NodeJson

	for _, i := range n.Bag {
		output.Bag = append(output.Bag, m[i])
	}

	for i := range n.Cover.Slice() {
		output.Cover = append(output.Cover, m[n.Cover.Slice()[i].Name])
	}

	for i := range n.Children {
		output.Children = append(output.Children, n.Children[i].IntoJson())
	}

	output.Star = n.Star

	return output
}

func (d DecompJson) IntoDecomp(graph Graph, encoding map[string]int) Decomp {

	var output Decomp

	output.Graph = graph

	output.Root = d.Root.IntoNode(graph, encoding)

	return output
}

func (n NodeJson) IntoNode(graph Graph, encoding map[string]int) Node {
	var output Node
	var cover []Edge

	for i := range n.Bag {
		output.Bag = append(output.Bag, encoding[n.Bag[i]])
	}

	for i := range n.Cover {
		cover = append(cover, extractEdge(graph.Edges.Slice(), encoding[n.Cover[i]]))
	}
	output.Cover = NewEdges(cover)

	output.Star = n.Star

	for i := range n.Children {
		output.Children = append(output.Children, n.Children[i].IntoNode(graph, encoding))
	}

	return output
}

func GetDecomp(input []byte, graph Graph, encoding map[string]int) Decomp {

	var jason DecompJson

	err := json.Unmarshal(input, &jason)
	if err != nil {
		fmt.Println("error:", err)
		log.Panicln("decomp couldn't be parased")
	}

	return jason.IntoDecomp(graph, encoding)
}

func WriteDecomp(input Decomp) []byte {
	out, err := json.Marshal(input.IntoJson())

	if err != nil {
		fmt.Println("error:", err)
		log.Panicln("decomp couldn't be marshalled")
	}

	return out
}

type Arc struct {
	Source int
	Target int
}

type ParseGMLValue struct {
	FlatVal string       ` @(Ident | Number) | "\"" @(Number | Ident | Punct)* "\""    `
	List    ParseGMLList `| "[" @@ "]"`
}

type ParseGMLListEntry struct {
	Key   string        ` @(Ident|Number) `
	Value ParseGMLValue ` @@ `
}

type ParseGMLList struct {
	Entries []ParseGMLListEntry `( @@ )*`
}

type ParseGML struct {
	GML ParseGMLList `@@`
}

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

	var parser = participle.MustBuild(&ParseGML{}, participle.UseLookahead(1), participle.Lexer(graphLexer), participle.Elide("Comment", "Whitespace"))
	pDecomp := ParseGML{}
	err := parser.ParseString(input, &pDecomp)
	if err != nil {
		fmt.Println("Couldn't parse input: ")
		panic(err)
	}

	// Check if GML file consists of single graph node
	if len(pDecomp.GML.Entries) != 1 && pDecomp.GML.Entries[0].Key != "graph" {
		log.Panicln("Valid GML file, but does not contain a unique graph element")
	}

	var graphEntry ParseGMLListEntry

	graphEntry = pDecomp.GML.Entries[0]

	var arcs []Arc
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

			re_cover := regexp.MustCompile(`{(.*)}{.*}`)
			re_bag := regexp.MustCompile(`{.*}{(.*)}`)

			match_cover := re_cover.FindStringSubmatch(nodeLabels["label"])
			match_bag := re_bag.FindStringSubmatch(nodeLabels["label"])
			if match_cover == nil || match_bag == nil {
				log.Panicln("Label of node ", nodeLabels["id"], " not properly formatted: ", nodeLabels["label"], ".")
			}

			var bag []int
			for _, v := range strings.Split(match_bag[1], ",") {
				bag = append(bag, encoding[v])
			}

			var cover []Edge
			for _, e := range strings.Split(match_cover[1], ",") {
				out := extractEdge(edges, encoding[e])
				if reflect.DeepEqual(out, Edge{}) {
					log.Panicln("Can't find edge ", e)
				}
				cover = append(cover, out)
			}

			node.num, _ = strconv.Atoi(nodeLabels["id"])

			encode = Max(node.num, encode) + 1 // ensure encode will never collide with num values in parsed GMl

			node.Bag = bag
			node.Cover = NewEdges(cover)
			if v, ok := nodeLabels["magic"]; ok {
				node.Star, _ = strconv.ParseBool(v)
			}

			nodes = append(nodes, node)

			IDtoIndex[node.num] = len(nodes) - 1
		case "edge":
			var arc Arc

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

			arc.Source, _ = strconv.Atoi(arcLabels["source"])
			arc.Target, _ = strconv.Atoi(arcLabels["target"])

			arcs = append(arcs, arc)
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

func GetCache(input []byte) Cache {

	var jason Cache

	err := json.Unmarshal(input, &jason)
	if err != nil {
		fmt.Println("error:", err)
		log.Panicln("Oh noes, can't part JSON")
	}

	return jason

}

// Updated PACE 2019 format, with initial Special Edges

// type ParseEdgeUpdate struct {
//  Name     string   ` @(Ident|Number)`
//  Vertices []string `"(" ( @(Ident|Number)  ","? )* ")"`
// }

// type ParseSpecialEdgeUpdate struct {
//  Name     string   ` @(Ident|Number)`
//  Vertices []string `"(" ( @(Ident|Number)  ","? )* ")"`
// }

// type ParseGhostEdgeUpdate struct {
//  Name     string   ` @(Ident|Number)`
//  Vertices []string `"(" ( @(Ident|Number)  ","? )* ")"`
// }

// type ParseGraphUpdate struct {
//  Edges   []ParseEdgeUpdate        `( @@ ","?)* "."`
//  Ghost   []ParseSpecialEdgeUpdate `("👻" ( @@ ","?)*)?`
//  Special []ParseGhostEdgeUpdate   `"✨" ( @@ ","?)*`
//  m       map[string]int
// }

// type ParseGraphUpdateInfo struct {
//  Vertices int `"p htd":Begin @(Number) `
//  Edges    int `@(Number) "\n"`
// }

// type ParseGraphUpdate struct {
//  Info         ParseGraphUpdateInfo     `@@`
//  Edges        []ParseEdgeUpdate        `(@@) *`
//  SpecialEdges []ParseSpecialEdgeUpdate `(@@) *`
//  GhostEdges   []ParseGhostEdgeUpdate   `(@@) *`
//  m            map[int]int
// }

// func GetGraphUpdate(s string) (Graph, Graph, []Special) {

//  graphLexer := lexer.Must(ebnf.New(`
//     Comment = ("%" | "//") { "\u0000"…"\uffff"-"\n" } .
//     Ident = (alpha | "_") { "_" | alpha | digit | stuff } .
//     Number = ("." | digit | "_"){"." | digit | "_"} .
//     Whitespace = " " | "\t" | "\n" | "\r" .
//     stuff = ":" | "@" | ";" | "-" .
//     Punct = "!"…"/"  .
//     alpha = "a"…"z" | "A"…"Z" .
//     SpecialSep = "✨" .
//     GhostSep  = "👻" .
//     digit = "0"…"9" .`))

//  var parser = participle.MustBuild(&ParseGraphUpdate{}, participle.UseLookahead(1), participle.Lexer(graphLexer),
//      participle.Elide("Comment", "Whitespace"))
//  var output Graph
//  var ghostGraph Graph
//  var edges []Edge
//  var ghostEdges []Edge
//  var special []Special

//  pgraph := ParseGraphUpdate{}
//  err := parser.ParseString(s, &pgraph)
//  if err != nil {
//      fmt.Println("Couldn't parse input: ")
//      panic(err)
//  }
//  encode = 1 // initialize to 1

//  encoding := make(map[int]string)
//  pgraph.m = make(map[string]int)

//  for _, e := range pgraph.Edges {
//      _, ok := pgraph.m[e.Name]
//      if ok {
//          log.Panicln("Edge names not unique, not a vald hypergraph!")
//      }
//      pgraph.m[e.Name] = encode
//      encoding[encode] = e.Name
//      encode++
//  }

//  for _, e := range pgraph.Ghost {
//      encoding[encode] = "👻" + e.Name
//      pgraph.m["👻"+e.Name] = encode
//      encode++
//  }

//  for _, e := range pgraph.Special {
//      encoding[encode] = e.Name + " ✨"
//      pgraph.m[e.Name+" ✨"] = encode
//      encode++
//  }

//  for _, e := range pgraph.Edges {
//      var outputEdge []int
//      for _, n := range e.Vertices {
//          i, ok := pgraph.m[n]
//          if ok {
//              outputEdge = append(outputEdge, i)
//          } else {
//              pgraph.m[n] = encode
//              encoding[encode] = n
//              outputEdge = append(outputEdge, encode)
//              encode++
//          }
//      }
//      edges = append(edges, Edge{Name: pgraph.m[e.Name], Vertices: outputEdge})
//  }

//  for _, e := range pgraph.Ghost {
//      var outputEdge []int
//      for _, n := range e.Vertices {
//          i, ok := pgraph.m[n]
//          if ok {
//              outputEdge = append(outputEdge, i)
//          } else {
//              pgraph.m[n] = encode
//              encoding[encode] = n
//              outputEdge = append(outputEdge, encode)
//              encode++

//          }
//      }
//      ghostEdges = append(ghostEdges, Edge{Name: pgraph.m["👻"+e.Name], Vertices: outputEdge})
//  }

//  for _, s := range pgraph.Special {
//      var outputSpecialEdge []int
//      for _, n := range s.Vertices {
//          i, ok := pgraph.m[n]
//          if ok {
//              outputSpecialEdge = append(outputSpecialEdge, i)
//          } else {
//              pgraph.m[n] = encode
//              encoding[encode] = n
//              outputSpecialEdge = append(outputSpecialEdge, encode)
//              encode++
//          }
//      }
//      dummyEdges := NewEdges([]Edge{Edge{Name: pgraph.m[s.Name+" ✨"], Vertices: outputSpecialEdge}})
//      special = append(special, Special{Vertices: outputSpecialEdge, Edges: dummyEdges})
//  }

//  m = encoding

//  output.Edges = NewEdges(edges)
//  ghostGraph.Edges = NewEdges(append(edges, ghostEdges...))

//  // log.Println("Edges", pgraph.Info.Edges)
//  // log.Println("Vertices", pgraph.Info.Vertices)

//  // for _, e := range output.Edges.Slice() {
//  //  log.Println(e.FullString())
//  // }

//  // for _, e := range special {
//  //  log.Println(e)
//  // }

//  return ghostGraph, output, special
// }
