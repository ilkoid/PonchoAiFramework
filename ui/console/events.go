package console

import (
	"time"
)

// FlowEvent represents an event that occurs during flow execution
type FlowEvent struct {
	Time   time.Time `json:"time"`
	Step   string    `json:"step"`
	Status string    `json:"status"`
	Detail string    `json:"detail,omitempty"`
}

// FlowObserver defines the interface for observing flow events
type FlowObserver interface {
	OnEvent(event FlowEvent)
}