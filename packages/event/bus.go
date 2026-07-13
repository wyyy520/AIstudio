package event

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventBus struct {
	mu           sync.RWMutex
	handlers     map[Topic][]Subscription
	wildcards    []Subscription
	asyncBufSize int
}

type Option func(*EventBus)

func WithAsyncBufferSize(size int) Option {
	return func(eb *EventBus) {
		eb.asyncBufSize = size
	}
}

func New(opts ...Option) *EventBus {
	eb := &EventBus{
		handlers:     make(map[Topic][]Subscription),
		wildcards:    make([]Subscription, 0),
		asyncBufSize: 100,
	}
	for _, opt := range opts {
		opt(eb)
	}
	return eb
}

func (eb *EventBus) Publish(topic Topic, data interface{}) {
	eb.PublishEvent(Event{
		ID:        uuid.New().String(),
		Topic:     topic,
		Timestamp: time.Now(),
		Data:      data,
	})
}

func (eb *EventBus) PublishEvent(event Event) {
	eb.mu.RLock()
	subs := eb.matchSubscribers(event.Topic)
	eb.mu.RUnlock()

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

func (eb *EventBus) callHandler(h EventHandler, event Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[eventbus] handler panic: topic=%s, error=%v", event.Topic, r)
		}
	}()
	h(event)
}

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

	return fmt.Sprintf("EventBus{topics=%d, subscribers=%d}",
		len(eb.handlers), totalSubs)
}
