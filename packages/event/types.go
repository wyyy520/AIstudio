package event

import "time"

type Topic string

type Event struct {
	ID        string      `json:"id"`
	Topic     Topic       `json:"topic"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Data      interface{} `json:"data,omitempty"`
}

type EventHandler func(Event)

type EventData interface {
	EventTopic() Topic
}

type Subscription struct {
	ID    string
	Topic Topic
	ch    chan Event
	h     EventHandler
	async bool
}
