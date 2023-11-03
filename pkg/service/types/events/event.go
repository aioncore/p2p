package events

import "fmt"

type Event interface {
	GetType() string
	GetID() string
	String() string
}

type EventHead struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (e *EventHead) GetType() string {
	return e.Type
}

func (e *EventHead) GetID() string {
	return e.ID
}

func (e *EventHead) String() string {
	return fmt.Sprintf("%s event %s", e.Type, e.ID)
}
