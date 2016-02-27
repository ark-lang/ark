package ast

import (
	"bytes"
)

type DependencyNode struct {
	Module  *ModuleName
	index   int
	lowlink int
	onStack bool
}

type Dependency struct{ Src, Dst *DependencyNode }
type NodeSet []*DependencyNode

type DependencyGraph struct {
	Nodes       NodeSet
	NodeIndices map[string]int
	EdgesFrom   map[string][]Dependency

	index int
	stack []*DependencyNode
	out   []NodeSet
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:       make(NodeSet, 0),
		NodeIndices: make(map[string]int),
		EdgesFrom:   make(map[string][]Dependency),
	}
}

func (v *DependencyGraph) getOrCreate(modname *ModuleName) *DependencyNode {
	idx, ok := v.NodeIndices[modname.String()]
	if !ok {
		idx = len(v.Nodes)
		v.Nodes = append(v.Nodes, &DependencyNode{Module: modname})
		v.NodeIndices[modname.String()] = idx
	}
	return v.Nodes[idx]
}

func (v *DependencyGraph) AddDependency(source, dependency *ModuleName) {
	srcNode := v.getOrCreate(source)
	dstNode := v.getOrCreate(dependency)
	dep := Dependency{Src: srcNode, Dst: dstNode}
	v.EdgesFrom[source.String()] = append(v.EdgesFrom[source.String()], dep)
}

func (d *DependencyGraph) DetectCycles() []string {
	scgs := d.tarjan()

	var errs []string
	for _, scg := range scgs {
		if len(scg) == 1 {
			continue
		}

		buf := new(bytes.Buffer)
		for idx, v := range scg {
			buf.WriteString(v.Module.String())
			if idx != len(scg)-1 {
				buf.WriteString(", ")
			}
		}
		errs = append(errs, buf.String())
	}

	return errs
}

func (v *DependencyGraph) tarjan() []NodeSet {
	// Tarjan's strongly connected components algorithm, as per:
	// https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
	v.out = nil

	v.index = 0
	v.stack = nil

	// Initial clear
	for _, node := range v.Nodes {
		node.index, node.lowlink = -1, -1
	}

	// Actual algorithm run
	for _, node := range v.Nodes {
		if node.index == -1 {
			v.strongConnect(node)
		}
	}

	return v.out
}

func (d *DependencyGraph) strongConnect(v *DependencyNode) {
	v.index = d.index
	v.lowlink = d.index
	d.index++

	d.stack = append(d.stack, v)
	v.onStack = true

	edges := d.EdgesFrom[v.Module.String()]
	for _, edge := range edges {
		w := edge.Dst
		if w.index == -1 {
			d.strongConnect(w)
			v.lowlink = min(v.lowlink, w.lowlink)
		} else if w.onStack {
			v.lowlink = min(v.lowlink, w.index)
		}
	}

	if v.lowlink == v.index {
		out := make(NodeSet, 0)
		for {
			idx := len(d.stack) - 1
			w := d.stack[idx]
			d.stack[idx] = nil
			d.stack = d.stack[:idx]

			w.onStack = false

			out = append(out, w)

			if w == v {
				break
			}
		}
		d.out = append(d.out, out)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
