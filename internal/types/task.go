package types

// TaskWithID represents a task with an ID for scheduler interface compatibility
type TaskWithID interface {
	GetDescription() string
	GetID() string
	GetStatus() string
}
