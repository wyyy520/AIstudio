// Package eventbus provides a publish/subscribe event system
// that decouples all AIStudio modules. Every module communicates
// through events — no direct dependencies between modules.
package eventbus

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Topic represents a unique event topic identifier.
type Topic string

// EventHandler is a function that processes an event.
type EventHandler func(Event)

// Subscription represents a registered event handler.
type Subscription struct {
	ID      string
	Topic   Topic
	handler EventHandler
}

// Event is the base event structure for all system events.
type Event struct {
	ID        string      `json:"id"`
	Topic     Topic       `json:"topic"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"` // Module name
	Data      interface{} `json:"data,omitempty"`
}

// EventBus provides publish/subscribe messaging between modules.
// It is the backbone of AIStudio's decoupled architecture.
type EventBus struct {
	mu            sync.RWMutex
	handlers      map[Topic][]Subscription
	history       []Event
	historySize   int
	traceEnabled  bool
}

// New creates a new EventBus.
func New(opts ...Option) *EventBus {
	eb := &EventBus{
		handlers:     make(map[Topic][]Subscription),
		history:      make([]Event, 0, 100),
		historySize:  100,
		traceEnabled: false,
	}
	for _, opt := range opts {
		opt(eb)
	}
	return eb
}

// Option configures the EventBus.
type Option func(*EventBus)

// WithHistorySize sets the maximum number of events to keep in history.
func WithHistorySize(size int) Option {
	return func(eb *EventBus) {
		eb.historySize = size
	}
}

// WithTrace enables event tracing for debugging.
func WithTrace(enabled bool) Option {
	return func(eb *EventBus) {
		eb.traceEnabled = enabled
	}
}

// Publish publishes an event to all subscribers of the topic.
// This is non-blocking — handlers are called in separate goroutines.
func (eb *EventBus) Publish(topic Topic, data interface{}) {
	eb.PublishEvent(Event{
		ID:        uuid.New().String(),
		Topic:     topic,
		Timestamp: time.Now(),
		Source:    "unknown",
		Data:      data,
	})
}

// PublishEvent publishes a pre-built event to all subscribers.
func (eb *EventBus) PublishEvent(event Event) {
	eb.mu.RLock()
	subs, exists := eb.handlers[event.Topic]
	eb.mu.RUnlock()

	if eb.traceEnabled {
		log.Printf("[eventbus] publish: %s (source=%s, subscribers=%d)",
			event.Topic, event.Source, len(subs))
	}

	// Store in history
	eb.mu.Lock()
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.historySize {
		eb.history = eb.history[1:]
	}
	eb.mu.Unlock()

	if !exists {
		return
	}

	// Call handlers asynchronously
	for _, sub := range subs {
		sub := sub // Capture for closure
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[eventbus] handler panic: topic=%s, subscription=%s, error=%v",
						event.Topic, sub.ID, r)
				}
			}()
			sub.handler(event)
		}()
	}
}

// Subscribe registers a handler for a topic.
// Returns a Subscription that can be used to unsubscribe.
func (eb *EventBus) Subscribe(topic Topic, handler EventHandler) Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := Subscription{
		ID:      uuid.New().String(),
		Topic:   topic,
		handler: handler,
	}

	eb.handlers[topic] = append(eb.handlers[topic], sub)

	if eb.traceEnabled {
		log.Printf("[eventbus] subscribe: %s (id=%s)", topic, sub.ID)
	}

	return sub
}

// SubscribeAll registers a handler for all topics.
func (eb *EventBus) SubscribeAll(handler EventHandler) Subscription {
	// Subscribe to a special wildcard by subscribing to each topic
	// For simplicity, we track all-topic subscriptions separately
	allSub := Subscription{
		ID:      uuid.New().String(),
		Topic:   "*",
		handler: handler,
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Subscribe to all existing topics
	for topic := range eb.handlers {
		sub := Subscription{
			ID:      uuid.New().String(),
			Topic:   topic,
			handler: handler,
		}
		eb.handlers[topic] = append(eb.handlers[topic], sub)
	}

	return allSub
}

// Unsubscribe removes a subscription.
func (eb *EventBus) Unsubscribe(sub Subscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, exists := eb.handlers[sub.Topic]
	if !exists {
		return
	}

	for i, s := range subs {
		if s.ID == sub.ID {
			eb.handlers[sub.Topic] = append(subs[:i], subs[i+1:]...)
			if eb.traceEnabled {
				log.Printf("[eventbus] unsubscribe: %s (id=%s)", sub.Topic, sub.ID)
			}
			return
		}
	}

	// Check all topics for wildcard subscription
	for topic, subs := range eb.handlers {
		for i, s := range subs {
			if s.ID == sub.ID {
				eb.handlers[topic] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

// HasSubscribers checks if a topic has any subscribers.
func (eb *EventBus) HasSubscribers(topic Topic) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, exists := eb.handlers[topic]
	return exists && len(subs) > 0
}

// SubscriberCount returns the number of subscribers for a topic.
func (eb *EventBus) SubscriberCount(topic Topic) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, exists := eb.handlers[topic]
	if !exists {
		return 0
	}
	return len(subs)
}

// History returns the event history.
func (eb *EventBus) History() []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make([]Event, len(eb.history))
	copy(result, eb.history)
	return result
}

// Clear clears all subscribers and history.
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers = make(map[Topic][]Subscription)
	eb.history = make([]Event, 0, eb.historySize)
}

// Close shuts down the event bus.
func (eb *EventBus) Close() {
	eb.Clear()
	log.Println("[eventbus] closed")
}

// String returns a string representation of the EventBus.
func (eb *EventBus) String() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	totalSubs := 0
	for _, subs := range eb.handlers {
		totalSubs += len(subs)
	}

	return fmt.Sprintf("EventBus{topics=%d, subscribers=%d, history=%d}",
		len(eb.handlers), totalSubs, len(eb.history))
}