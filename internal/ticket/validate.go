package ticket

import "time"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func Validate(t *Ticket) []ValidationError {
	var errs []ValidationError

	if t.ID == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "must not be empty"})
	}

	if t.Title == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "must not be empty"})
	}

	if !t.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: "invalid status: " + string(t.Status)})
	}

	if !t.Priority.IsValid() {
		errs = append(errs, ValidationError{Field: "priority", Message: "invalid priority: " + string(t.Priority)})
	}

	if !t.Actor.IsValid() {
		errs = append(errs, ValidationError{Field: "actor", Message: "invalid actor: " + string(t.Actor) + " (use agent or human)"})
	}

	if !t.Type.IsValid() {
		errs = append(errs, ValidationError{Field: "type", Message: "invalid type: " + string(t.Type) + " (use feature, bug, spike, chore, docs)"})
	}

	if t.Complexity < 0 || t.Complexity > 10 {
		errs = append(errs, ValidationError{Field: "complexity", Message: "must be between 0 and 10"})
	}

	if t.Created.Equal(time.Time{}) {
		errs = append(errs, ValidationError{Field: "created", Message: "must not be zero"})
	}

	if t.Estimate != "" {
		if _, err := ParseEstimate(t.Estimate); err != nil {
			errs = append(errs, ValidationError{Field: "estimate", Message: "invalid format: use 30m, 2h, 1d, 1w"})
		}
	}

	if t.ParentID != "" && t.ParentID == t.ID {
		errs = append(errs, ValidationError{Field: "parent_id", Message: "ticket cannot be its own parent"})
	}

	for _, id := range t.Blocks {
		if id == t.ID {
			errs = append(errs, ValidationError{Field: "blocks", Message: "ticket cannot block itself"})
			break
		}
	}

	for _, id := range t.RelatedTo {
		if id == t.ID {
			errs = append(errs, ValidationError{Field: "related_to", Message: "ticket cannot be related to itself"})
			break
		}
	}

	for _, id := range t.Duplicates {
		if id == t.ID {
			errs = append(errs, ValidationError{Field: "duplicates", Message: "ticket cannot duplicate itself"})
			break
		}
	}

	return errs
}
