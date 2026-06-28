package dag

import (
	"sort"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

// EdgeType describes the relationship between two nodes.
type EdgeType int

const (
	EdgeDependency EdgeType = iota
	EdgeBlocks
	EdgeRelated
	EdgeDuplicate
	EdgeParent
)

type Graph struct {
	edges     map[string][]string
	reverse   map[string][]string
	nodes     map[string]bool
	edgeTypes map[string]map[string]EdgeType
	parents   map[string]string
	children  map[string][]string
}

func New() *Graph {
	return &Graph{
		edges:     make(map[string][]string),
		reverse:   make(map[string][]string),
		nodes:     make(map[string]bool),
		edgeTypes: make(map[string]map[string]EdgeType),
		parents:   make(map[string]string),
		children:  make(map[string][]string),
	}
}

func (g *Graph) AddNode(id string) {
	g.nodes[id] = true
}

func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.edges[from] = append(g.edges[from], to)
	g.reverse[to] = append(g.reverse[to], from)
}

func (g *Graph) setEdgeType(from, to string, et EdgeType) {
	if g.edgeTypes[from] == nil {
		g.edgeTypes[from] = make(map[string]EdgeType)
	}
	g.edgeTypes[from][to] = et
}

// GetEdgeType returns the edge type between two nodes.
// If no explicit type is recorded, EdgeDependency is returned.
func (g *Graph) GetEdgeType(from, to string) EdgeType {
	if m, ok := g.edgeTypes[from]; ok {
		if et, ok := m[to]; ok {
			return et
		}
	}
	return EdgeDependency
}

// SetParent records a parent-child relationship (not a DAG edge).
func (g *Graph) SetParent(child, parent string) {
	g.parents[child] = parent
	g.children[parent] = append(g.children[parent], child)
}

// Parent returns the parent ID for a node, or "" if none.
func (g *Graph) Parent(id string) string {
	return g.parents[id]
}

// Children returns the child IDs for a node.
func (g *Graph) Children(id string) []string {
	c := make([]string, len(g.children[id]))
	copy(c, g.children[id])
	sort.Strings(c)
	return c
}

func (g *Graph) Dependencies(id string) []string {
	deps := make([]string, len(g.edges[id]))
	copy(deps, g.edges[id])
	sort.Strings(deps)
	return deps
}

func (g *Graph) Dependents(id string) []string {
	deps := make([]string, len(g.reverse[id]))
	copy(deps, g.reverse[id])
	sort.Strings(deps)
	return deps
}

func (g *Graph) Nodes() []string {
	result := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func BuildFromTickets(tickets []ticket.Ticket) *Graph {
	g := New()
	for _, t := range tickets {
		g.AddNode(t.ID)
	}
	for _, t := range tickets {
		// Direct dependencies: t depends on dep.
		for _, dep := range t.Dependencies {
			g.AddEdge(t.ID, dep)
			g.setEdgeType(t.ID, dep, EdgeDependency)
		}
		// Blocks: if t blocks X, then X depends on t.
		for _, blocked := range t.Blocks {
			g.AddEdge(blocked, t.ID)
			g.setEdgeType(blocked, t.ID, EdgeBlocks)
		}
		// RelatedTo: informational only, no DAG edge.
		for _, rel := range t.RelatedTo {
			g.setEdgeType(t.ID, rel, EdgeRelated)
		}
		// Duplicates: informational only, no DAG edge.
		for _, dup := range t.Duplicates {
			g.setEdgeType(t.ID, dup, EdgeDuplicate)
		}
		// Parent-child: hierarchy, not a DAG edge.
		if t.ParentID != "" {
			g.SetParent(t.ID, t.ParentID)
		}
	}
	return g
}
