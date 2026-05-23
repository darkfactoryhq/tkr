package store

import (
	"strconv"
	"strings"

	"github.com/tkr-cli/tkr/internal/ticket"
)

type Op int

const (
	OpEq Op = iota
	OpIn
	OpGt
	OpLt
	OpRange
	OpContains
)

type Clause struct {
	Field    string
	Operator Op
	Values   []string
	Negate   bool
}

type Filter struct {
	Clauses []Clause
}

func ParseFilter(args []string) (*Filter, error) {
	f := &Filter{}

	for _, arg := range args {
		c, err := parseClause(arg)
		if err != nil {
			return nil, err
		}
		f.Clauses = append(f.Clauses, c)
	}

	return f, nil
}

func parseClause(arg string) (Clause, error) {
	if strings.HasPrefix(arg, "+") {
		return Clause{
			Field:    "labels",
			Operator: OpContains,
			Values:   []string{arg[1:]},
		}, nil
	}

	if strings.HasPrefix(arg, "-") && !strings.Contains(arg, ":") {
		return Clause{
			Field:    "status",
			Operator: OpEq,
			Values:   []string{arg[1:]},
			Negate:   true,
		}, nil
	}

	idx := strings.Index(arg, ":")
	if idx == -1 {
		return Clause{
			Field:    "status",
			Operator: OpEq,
			Values:   []string{arg},
		}, nil
	}

	field := arg[:idx]
	value := arg[idx+1:]

	if strings.HasPrefix(value, ">") {
		return Clause{
			Field:    field,
			Operator: OpGt,
			Values:   []string{value[1:]},
		}, nil
	}

	if strings.HasPrefix(value, "<") {
		return Clause{
			Field:    field,
			Operator: OpLt,
			Values:   []string{value[1:]},
		}, nil
	}

	if strings.Contains(value, "..") {
		parts := strings.SplitN(value, "..", 2)
		return Clause{
			Field:    field,
			Operator: OpRange,
			Values:   parts,
		}, nil
	}

	if strings.Contains(value, ",") {
		return Clause{
			Field:    field,
			Operator: OpIn,
			Values:   strings.Split(value, ","),
		}, nil
	}

	return Clause{
		Field:    field,
		Operator: OpEq,
		Values:   []string{value},
	}, nil
}

func (f *Filter) Match(t ticket.Ticket) bool {
	for _, c := range f.Clauses {
		result := matchClause(c, t)
		if c.Negate {
			result = !result
		}
		if !result {
			return false
		}
	}
	return true
}

func matchClause(c Clause, t ticket.Ticket) bool {
	fieldVal := fieldValue(c.Field, t)

	switch c.Operator {
	case OpEq:
		return eqInsensitive(fieldVal, c.Values[0])

	case OpIn:
		for _, v := range c.Values {
			if eqInsensitive(fieldVal, v) {
				return true
			}
		}
		return false

	case OpContains:
		return containsInsensitive(fieldSlice(c.Field, t), c.Values[0])

	case OpGt:
		a, err1 := strconv.Atoi(fieldVal)
		b, err2 := strconv.Atoi(c.Values[0])
		if err1 != nil || err2 != nil {
			return strings.ToLower(fieldVal) > strings.ToLower(c.Values[0])
		}
		return a > b

	case OpLt:
		a, err1 := strconv.Atoi(fieldVal)
		b, err2 := strconv.Atoi(c.Values[0])
		if err1 != nil || err2 != nil {
			return strings.ToLower(fieldVal) < strings.ToLower(c.Values[0])
		}
		return a < b

	case OpRange:
		a, err1 := strconv.Atoi(fieldVal)
		lo, err2 := strconv.Atoi(c.Values[0])
		hi, err3 := strconv.Atoi(c.Values[1])
		if err1 != nil || err2 != nil || err3 != nil {
			return false
		}
		return a >= lo && a <= hi
	}

	return false
}

func fieldValue(field string, t ticket.Ticket) string {
	switch strings.ToLower(field) {
	case "id":
		return t.ID
	case "title":
		return t.Title
	case "status":
		return string(t.Status)
	case "priority":
		return string(t.Priority)
	case "actor":
		return string(t.Actor)
	case "type":
		return string(t.Type)
	case "complexity":
		return strconv.Itoa(t.Complexity)
	case "branch":
		return t.Branch
	case "pr":
		return t.PR
	case "depends", "dependencies":
		return strings.Join(t.Dependencies, ",")
	case "labels":
		return strings.Join(t.Labels, ",")
	case "parent", "parent_id":
		return t.ParentID
	case "assignee":
		return t.Assignee
	case "estimate":
		return t.Estimate
	case "blocks":
		return strings.Join(t.Blocks, ",")
	case "related", "related_to":
		return strings.Join(t.RelatedTo, ",")
	case "duplicates":
		return strings.Join(t.Duplicates, ",")
	default:
		return ""
	}
}

func fieldSlice(field string, t ticket.Ticket) []string {
	switch strings.ToLower(field) {
	case "labels":
		return t.Labels
	case "depends", "dependencies":
		return t.Dependencies
	case "blocks":
		return t.Blocks
	case "related", "related_to":
		return t.RelatedTo
	case "duplicates":
		return t.Duplicates
	default:
		return nil
	}
}

func eqInsensitive(a, b string) bool {
	return strings.EqualFold(a, b)
}

func containsInsensitive(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}
