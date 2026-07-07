package task

import "fmt"

// stateTransitions defines valid status transitions.
var stateTransitions = map[Status][]Status{
	StatusPending:   {StatusRunning, StatusCancelled},
	StatusRunning:   {StatusSuccess, StatusFailed, StatusCancelled},
	StatusSuccess:   {}, // terminal
	StatusFailed:    {}, // terminal
	StatusCancelled: {}, // terminal
}

// ValidateTransition checks if a status transition is valid.
// Returns an error if the transition is not allowed.
func ValidateTransition(current, next Status) error {
	allowed, exists := stateTransitions[current]
	if !exists {
		return fmt.Errorf("unknown current status: %s", current)
	}
	for _, s := range allowed {
		if s == next {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition: %s -> %s", current, next)
}

// IsTerminal returns true if the status is a terminal state.
func IsTerminal(s Status) bool {
	return s == StatusSuccess || s == StatusFailed || s == StatusCancelled
}

// IsActive returns true if the task is still active (pending or running).
func IsActive(s Status) bool {
	return s == StatusPending || s == StatusRunning
}