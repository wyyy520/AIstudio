// Package event provides the publish/subscribe EventBus — the nervous system
// of AIStudio.
//
// All inter-module communication flows through the EventBus. Compiler progress,
// runtime logs, task status changes, plugin events — everything is a typed event
// published to named Topics.
//
// Key design decisions:
//   - In-memory: No external broker dependency for the single-process case.
//   - Sync + Async: Subscribers choose between synchronous handler calls
//     (guaranteed delivery, blocks publisher) or async channels (non-blocking,
//     drops on overflow).
//   - Wildcard matching: Subscribing to "runtime.*" receives all runtime
//     sub-topics (log, progress, started, etc).
//   - History ring: Last N events are retained for late joiners / debugging.
//   - Panic recovery: Each handler is wrapped in a recover() to prevent a
//     broken subscriber from crashing the bus.
//   - Thread-safe: All operations are protected by sync.RWMutex.
//
// Usage:
//
//	bus := event.New(event.WithHistorySize(1000))
//	bus.Subscribe(event.TopicRuntimeLog, func(e event.Event) { ... })
//	bus.Publish(event.TopicRuntimeLog, event.LogEventData{...})
//
// EngStudio.md §8 — Event System
package event

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// EventBus — the central publish/subscribe engine
// ============================================================================

// EventBus manages topic subscriptions and event delivery.
// It is the ONLY mechanism for cross-module communication.
type EventBus struct {
	mu           sync.RWMutex
	handlers     map[Topic][]Subscription
	wildcards    []Subscription
	asyncBufSize int
	history      []Event
	historySize  int
	traceEnabled bool
}

// ============================================================================
// Constructor Options
// ============================================================================

// Option is a functional option for configuring an EventBus.
type Option func(*EventBus)

// WithAsyncBufferSize sets the buffer size for async subscriber channels.
func WithAsyncBufferSize(size int) Option {
	return func(eb *EventBus) {
		eb.asyncBufSize = size
	}
}

func WithHistorySize(size int) Option {
	return func(eb *EventBus) {
		eb.historySize = size
	}
}

func WithTrace(enabled bool) Option {
	return func(eb *EventBus) {
		eb.traceEnabled = enabled
	}
}

// New creates a new EventBus with the given options.
// Defaults: asyncBufSize=100, historySize=100, traceEnabled=false.
func New(opts ...Option) *EventBus {
	eb := &EventBus{
		handlers:     make(map[Topic][]Subscription),
		wildcards:    make([]Subscription, 0),
		asyncBufSize: 100,
		history:      make([]Event, 0, 100),
		historySize:  100,
		traceEnabled: false,
	}
	for _, opt := range opts {
		opt(eb)
	}
	return eb
}

// ============================================================================
// Publishing — synchronously deliver events to all matching subscribers
// ============================================================================

// Publish creates an Event from the topic and data, then delivers it.
// Convenience wrapper around PublishEvent.
func (eb *EventBus) Publish(topic Topic, data interface{}) {
	eb.PublishEvent(Event{
		ID:        uuid.New().String(),
		Topic:     topic,
		Timestamp: time.Now(),
		Source:    "unknown",
		Data:      data,
	})
}

// PublishEvent delivers a pre-built Event to all matching subscribers.
// Steps:
//  1. Match subscribers (exact + wildcard)
//  2. Append to history ring
//  3. Deliver: async → channel send (may drop), sync → direct call
func (eb *EventBus) PublishEvent(event Event) {
	eb.mu.RLock()
	subs := eb.matchSubscribers(event.Topic)
	eb.mu.RUnlock()

	if eb.traceEnabled {
		log.Printf("[eventbus] publish: %s (source=%s, subscribers=%d)",
			event.Topic, event.Source, len(subs))
	}

	eb.mu.Lock()
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.historySize {
		eb.history = eb.history[1:]
	}
	eb.mu.Unlock()

	for _, sub := range subs {
		if sub.async {
			select {
			case sub.ch <- event:
			default:
				log.Printf("[eventbus] drop event: %s (subscriber %s buffer full)", event.Topic, sub.ID)
			}
		} else {
			eb.callHandler(sub.h, event)
		}
	}
}

func (eb *EventBus) matchSubscribers(topic Topic) []Subscription {
	var matched []Subscription
	for _, sub := range eb.wildcards {
		if matchTopic(sub.Topic, topic) {
			matched = append(matched, sub)
		}
	}
	if subs, ok := eb.handlers[topic]; ok {
		matched = append(matched, subs...)
	}
	return matched
}

func matchTopic(pattern Topic, topic Topic) bool {
	if pattern == "*" {
		return true
	}
	p := string(pattern)
	t := string(topic)
	if idx := strings.IndexByte(p, '*'); idx >= 0 {
		prefix := p[:idx]
		if idx > 0 && (p[idx-1] == '.' || p[idx-1] == ':') {
			prefix = p[:idx-1]
		}
		return strings.HasPrefix(t, prefix)
	}
	return p == t
}

// ============================================================================
// Subscribing — register handlers for topics
// ============================================================================

// Subscribe registers a synchronous handler for the given topic.
// The handler is called in the publisher's goroutine — use for simple
// in-process work like logging or state updates.
func (eb *EventBus) Subscribe(topic Topic, handler EventHandler) Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := Subscription{
		ID:    uuid.New().String(),
		Topic: topic,
		h:     handler,
	}

	if isWildcard(topic) {
		eb.wildcards = append(eb.wildcards, sub)
	} else {
		eb.handlers[topic] = append(eb.handlers[topic], sub)
	}

	return sub
}

// SubscribeAsync registers an async handler: events are buffered via channel
// and processed in a dedicated goroutine. Use when the handler does I/O
// or heavy work that shouldn't block the publisher.
func (eb *EventBus) SubscribeAsync(topic Topic, handler EventHandler) Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := Subscription{
		ID:    uuid.New().String(),
		Topic: topic,
		h:     handler,
		ch:    make(chan Event, eb.asyncBufSize),
		async: true,
	}

	if strings.HasSuffix(string(topic), ":*") || topic == "*" {
		eb.wildcards = append(eb.wildcards, sub)
	} else {
		eb.handlers[topic] = append(eb.handlers[topic], sub)
	}

	go eb.subLoop(sub)

	return sub
}

func isWildcard(t Topic) bool {
	s := string(t)
	return s == "*" || strings.HasSuffix(s, ":*") || strings.HasSuffix(s, ".*")
}

func (eb *EventBus) subLoop(sub Subscription) {
	for event := range sub.ch {
		eb.callHandler(sub.h, event)
	}
}

// callHandler invokes an EventHandler with panic recovery.
// A panicking handler never crashes the bus.
func (eb *EventBus) callHandler(h EventHandler, event Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[eventbus] handler panic: topic=%s, error=%v", event.Topic, r)
		}
	}()
	h(event)
}

// ============================================================================
// Lifecycle — unsubscribe, query, clear, close
// ============================================================================

// Unsubscribe removes a subscription and closes its channel if async.
func (eb *EventBus) Unsubscribe(sub Subscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for i, s := range eb.wildcards {
		if s.ID == sub.ID {
			eb.wildcards = append(eb.wildcards[:i], eb.wildcards[i+1:]...)
			if sub.async && sub.ch != nil {
				close(sub.ch)
			}
			return
		}
	}

	for topic, subs := range eb.handlers {
		for i, s := range subs {
			if s.ID == sub.ID {
				eb.handlers[topic] = append(subs[:i], subs[i+1:]...)
				if sub.async && sub.ch != nil {
					close(sub.ch)
				}
				return
			}
		}
	}
}

func (eb *EventBus) HasSubscribers(topic Topic) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if _, exists := eb.handlers[topic]; exists {
		return len(eb.handlers[topic]) > 0
	}
	for _, sub := range eb.wildcards {
		if matchTopic(sub.Topic, topic) {
			return true
		}
	}
	return false
}

func (eb *EventBus) SubscriberCount(topic Topic) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	count := 0
	if subs, ok := eb.handlers[topic]; ok {
		count += len(subs)
	}
	for _, sub := range eb.wildcards {
		if matchTopic(sub.Topic, topic) {
			count++
		}
	}
	return count
}

func (eb *EventBus) History() []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make([]Event, len(eb.history))
	copy(result, eb.history)
	return result
}

func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, sub := range eb.wildcards {
		if sub.async && sub.ch != nil {
			close(sub.ch)
		}
	}
	for _, subs := range eb.handlers {
		for _, sub := range subs {
			if sub.async && sub.ch != nil {
				close(sub.ch)
			}
		}
	}

	eb.handlers = make(map[Topic][]Subscription)
	eb.wildcards = make([]Subscription, 0)
	eb.history = make([]Event, 0, eb.historySize)
}

func (eb *EventBus) Close() {
	eb.Clear()
}

func (eb *EventBus) String() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	totalSubs := 0
	for _, subs := range eb.handlers {
		totalSubs += len(subs)
	}
	totalSubs += len(eb.wildcards)

	return fmt.Sprintf("EventBus{topics=%d, subscribers=%d, history=%d}",
		len(eb.handlers), totalSubs, len(eb.history))
}
