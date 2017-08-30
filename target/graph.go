package target

import "github.com/antha-lang/antha/graph"

// Graph is a view of instructions as a graph
type Graph struct {
	Insts []Inst
}

// NumNodes implements a Graph
func (a *Graph) NumNodes() int {
	return len(a.Insts)
}

// Node implements a Graph
func (a *Graph) Node(i int) graph.Node {
	return a.Insts[i].(graph.Node)
}

// NumOuts implements a Graph
func (a *Graph) NumOuts(n graph.Node) int {
	return len(n.(Inst).DependsOn())
}

// Out implements a Graph
func (a *Graph) Out(n graph.Node, i int) graph.Node {
	return n.(Inst).DependsOn()[i].(graph.Node)
}
