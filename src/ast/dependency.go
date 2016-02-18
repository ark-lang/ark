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
type NodeSet map[string]*DependencyNode

type DependencyGraph struct {
	Nodes     NodeSet
	EdgesFrom map[string][]Dependency

	index int
	stack []*DependencyNode
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:     make(NodeSet),
		EdgesFrom: make(map[string][]Dependency),
	}
}

func (v *DependencyGraph) AddDependency(source, dependency *ModuleName) {
	srcNode, ok := v.Nodes[source.String()]
	if !ok {
		srcNode = &DependencyNode{Module: source}
		v.Nodes[source.String()] = srcNode
	}

	dstNode, ok := v.Nodes[dependency.String()]
	if !ok {
		dstNode = &DependencyNode{Module: dependency}
		v.Nodes[dependency.String()] = dstNode
	}

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
		for _, v := range scg {
			edges := d.EdgesFrom[v.Module.String()]
			for _, edge := range edges {
				w := edge.Dst
				if _, ok := scg[w.Module.String()]; !ok {
					continue
				}

				buf.WriteString(v.Module.String())
				buf.WriteString(" => ")
				buf.WriteString(w.Module.String())
				buf.WriteString(", ")
			}
			buf.WriteString("\n")
		}
		errs = append(errs, buf.String())
	}

	return errs
}

func (v *DependencyGraph) tarjan() []NodeSet {
	// Tarjan's strongly connected components algorithm, as per:
	// https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
	var out []NodeSet

	v.index = 0
	v.stack = nil

	// Initial clear
	for _, node := range v.Nodes {
		node.index, node.lowlink = -1, -1
	}

	// Actual algorithm run
	for _, node := range v.Nodes {
		if node.index == -1 {
			res := v.strongConnect(node)
			if res == nil {
				panic("Illegal state in depencency check")
			}
			out = append(out, res)
		}
	}

	return out
}

func (d *DependencyGraph) strongConnect(v *DependencyNode) NodeSet {
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
		out := make(NodeSet)
		for {
			idx := len(d.stack) - 1
			w := d.stack[idx]
			d.stack[idx] = nil
			d.stack = d.stack[:idx]
			w.onStack = false

			out[w.Module.String()] = w

			if w == v {
				break
			}
		}
		return out
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
