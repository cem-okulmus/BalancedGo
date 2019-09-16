// transforms a Decomp into a .gml file
package lib

import (
	"bytes"
	"fmt"
	"log"
)

func (n Node) toGML() string {

	var buffer bytes.Buffer

	current := "  node [\n    id " + fmt.Sprint(n.num) +
		"\n    label \"" + n.Cover.String() + " " + PrintVertices(n.Bag) +
		"\"\n    vgj [\n      labelPosition \"in\"\n      shape \"Rectangle\"\n    ]\n  ]\n\n"

	buffer.WriteString(current)

	for i := range n.Children {
		buffer.WriteString(n.Children[i].toGML())
	}

	return buffer.String()
}

func (e Edge) toGML() string {
	if len(e.Vertices) != 2 {
		log.Panicln("can't convert proper hyperedge to GML!")
	}
	return "  edge [\n    source " + fmt.Sprint(e.Vertices[0]) +
		"\n    target " + fmt.Sprint(e.Vertices[1]) + "\n  ]\n\n"
}

func (d Decomp) ToGML() string {
	var buffer bytes.Buffer

	buffer.WriteString("graph [\n\n  directed 0\n\n")

	edges := d.Root.getConGraph(false).Slice()

	buffer.WriteString(d.Root.toGML())

	for i := range edges {
		buffer.WriteString(edges[i].toGML())
	}

	buffer.WriteString("\n]\n")

	return buffer.String()
}
