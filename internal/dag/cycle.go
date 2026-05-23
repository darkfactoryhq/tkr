package dag

import "sort"

type CycleResult struct {
	HasCycle   bool
	CycleNodes []string
	TopoOrder  []string
}

func DetectCycles(g *Graph) CycleResult {
	inDegree := make(map[string]int)
	for _, id := range g.Nodes() {
		inDegree[id] = len(g.edges[id])
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, dep := range g.Dependents(node) {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(order) == len(g.nodes) {
		return CycleResult{
			HasCycle:  false,
			TopoOrder: order,
		}
	}

	var cycleNodes []string
	processed := make(map[string]bool)
	for _, id := range order {
		processed[id] = true
	}
	for id := range g.nodes {
		if !processed[id] {
			cycleNodes = append(cycleNodes, id)
		}
	}
	sort.Strings(cycleNodes)

	return CycleResult{
		HasCycle:   true,
		CycleNodes: cycleNodes,
	}
}

// DetectParentCycles walks the parent chain for every node and returns a
// sorted list of node IDs that participate in a parent-hierarchy cycle.
func DetectParentCycles(g *Graph) []string {
	globalVisited := make(map[string]bool)
	inCycle := make(map[string]bool)

	for id := range g.nodes {
		if g.parents[id] == "" || globalVisited[id] {
			continue
		}

		// Walk the parent chain, tracking the path.
		chainVisited := make(map[string]bool)
		var chain []string
		cur := id
		for cur != "" && !globalVisited[cur] {
			if chainVisited[cur] {
				// Found a cycle — mark everything from cur onward.
				for i := len(chain) - 1; i >= 0; i-- {
					inCycle[chain[i]] = true
					if chain[i] == cur {
						break
					}
				}
				inCycle[cur] = true
				break
			}
			chainVisited[cur] = true
			chain = append(chain, cur)
			cur = g.parents[cur]
		}
		for _, n := range chain {
			globalVisited[n] = true
		}
	}

	var result []string
	for id := range inCycle {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
