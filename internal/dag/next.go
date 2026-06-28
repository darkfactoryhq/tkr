package dag

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

type NextOptions struct {
	Actor    string
	Assignee string
	Labels   []string
	Type     string
}

func NextTicket(tickets []ticket.Ticket, g *Graph, opts NextOptions) (*ticket.Ticket, error) {
	result := DetectCycles(g)
	if result.HasCycle {
		return nil, fmt.Errorf("dependency cycle detected among: %s", strings.Join(result.CycleNodes, ", "))
	}

	done := make(map[string]bool)
	for _, t := range tickets {
		if t.Status == ticket.StatusDone {
			done[t.ID] = true
		}
	}

	var candidates []ticket.Ticket
	for _, t := range tickets {
		if t.Status != ticket.StatusTodo {
			continue
		}

		allDepsDone := true
		for _, dep := range g.Dependencies(t.ID) {
			if !done[dep] {
				allDepsDone = false
				break
			}
		}
		if !allDepsDone {
			continue
		}

		if opts.Actor != "" && t.Actor != "" && string(t.Actor) != opts.Actor {
			continue
		}
		if opts.Type != "" && t.Type != "" && string(t.Type) != opts.Type {
			continue
		}

		if opts.Assignee != "" && t.Assignee != "" && t.Assignee != opts.Assignee {
			continue
		}

		if len(opts.Labels) > 0 {
			labelSet := make(map[string]bool)
			for _, l := range t.Labels {
				labelSet[l] = true
			}
			hasAll := true
			for _, l := range opts.Labels {
				if !labelSet[l] {
					hasAll = false
					break
				}
			}
			if !hasAll {
				continue
			}
		}

		candidates = append(candidates, t)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		ri, rj := candidates[i].Priority.Rank(), candidates[j].Priority.Rank()
		if ri != rj {
			return ri < rj
		}
		if candidates[i].Complexity != candidates[j].Complexity {
			return candidates[i].Complexity < candidates[j].Complexity
		}
		ei, ej := candidates[i].EstimateDuration(), candidates[j].EstimateDuration()
		if ei != ej {
			return ei < ej
		}
		return naturalLess(candidates[i].ID, candidates[j].ID)
	})

	return &candidates[0], nil
}

func naturalLess(a, b string) bool {
	ai, an := splitID(a)
	bi, bn := splitID(b)
	if ai != bi {
		return ai < bi
	}
	return an < bn
}

// splitID splits an ID like "TKR-10" into prefix "TKR-" and numeric part 10.
func splitID(id string) (string, int) {
	idx := strings.LastIndex(id, "-")
	if idx < 0 {
		return id, 0
	}
	n, err := strconv.Atoi(id[idx+1:])
	if err != nil {
		return id, 0
	}
	return id[:idx+1], n
}
