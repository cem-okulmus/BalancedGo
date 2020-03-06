// Components with two connecting sets, used for DivideK
package lib

import (
	"bytes"
	"fmt"
)

type DivideComp struct {
	Up           []int
	Length       int
	Edges        Edges
	Low          []int
	UpConnecting bool
}

func (c DivideComp) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Up: ")
	buffer.WriteString(Edge{Vertices: c.Up}.String())

	buffer.WriteString(" Edges: ")
	buffer.WriteString(Graph{Edges: c.Edges}.String())

	buffer.WriteString(" Low: ")
	buffer.WriteString(Edge{Vertices: c.Low}.String())

	buffer.WriteString(" UpConnecting: ")
	buffer.WriteString(fmt.Sprintln(c.UpConnecting))

	return buffer.String()
}

func (comp DivideComp) GetComponents(sep Edges) ([]DivideComp, bool) {
	var output []DivideComp

	// log.Println("Edges coming in, ", comp.Edges)

	//special case
	if Subset(comp.Edges.Vertices(), sep.Vertices()) {
		return []DivideComp{}, true
	}

	var mustCover = Inter(comp.Up, comp.Low)
	if !Subset(mustCover, sep.Vertices()) {
		return output, false
	}

	// log.Println("Testing ", sep)

	var Up []int
	var Low []int
	if !Subset(comp.Up, sep.Vertices()) {
		Up = comp.Up // ignore Up connection if sep already fully covers it
	}
	if !Subset(comp.Low, sep.Vertices()) {
		Low = comp.Low // ignore Low connection if sep already fully covers it
	}

	sepBoth := NewEdges(sep.Both(comp.Edges))

	comps, UpIndex, LowIndex, isolatedComp := Graph{Edges: comp.Edges}.GetComponentsIsolated(sep, sepBoth, Up, Low)

	// log.Println("Edges coming out, ", len(comps), " many comps:")

	// for i := range comps {
	//  log.Println("comp ", comps[i].Edges)
	// }

	// log.Println("UpIndex ", UpIndex, " LowIndex", LowIndex)

	if UpIndex != -1 && UpIndex == LowIndex { // reject case, Up and Low not seperated
		return output, false
	}

	unionOfUps := []int{}

	// take care of component, calculating connection sets
	for i := range comps {
		c := DivideComp{}
		c.Length = comps[i].Edges.Len() - len(sep.Both(comps[i].Edges))

		if i == UpIndex { // Upper component
			c.UpConnecting = true
			compEdges := comps[i].Edges.Slice()
			c.Low = Inter(sep.Vertices(), comps[i].Edges.Vertices())

			c.Edges = NewEdges(append(compEdges, sepBoth.Slice()...))
			c.Edges.RemoveDuplicates()

			c.Up = comp.Up
		} else if i == LowIndex { // Lower component
			c.Edges = comps[i].Edges
			c.Low = Inter(comp.Low, c.Edges.Vertices())
			c.Up = Inter(sep.Vertices(), c.Edges.Vertices())
			unionOfUps = append(unionOfUps, c.Up...)
		} else {
			c.Edges = comps[i].Edges
			c.Up = Inter(sep.Vertices(), c.Edges.Vertices())
			unionOfUps = append(unionOfUps, c.Up...)

		}

		output = append(output, c)
	}
	if UpIndex != -1 && len(unionOfUps) > 0 {
		output[UpIndex].Low = append(output[UpIndex].Low, unionOfUps...)
		output[UpIndex].Low = Inter(output[UpIndex].Edges.Vertices(), output[UpIndex].Low)
		output[UpIndex].Low = RemoveDuplicates(output[UpIndex].Low)

	}
	// //Add isolated comp to Low set
	if UpIndex != -1 && len(isolatedComp) > 0 {
		coveredEdges := NewEdges(isolatedComp)
		output[UpIndex].Low = append(output[UpIndex].Low, coveredEdges.Vertices()...)
		output[UpIndex].Low = Inter(output[UpIndex].Edges.Vertices(), output[UpIndex].Low)
		output[UpIndex].Low = RemoveDuplicates(output[UpIndex].Low)

	}

	return output, true
}
