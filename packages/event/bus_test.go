package event

import (
	"sync"
	"testing"
	"time"
)

func TestSubscribeAndPublish(t *testing.T) {
	bus := New()
	defer bus.Close()

	received := make(chan Event, 1)
	bus.Subscribe("test.topic", func(e Event) {
		received <- e
	})

	bus.Publish("test.topic", "hello")

	select {
	case e := <-received:
		if e.Topic != "test.topic" {
			t.Errorf("expected topic test.topic, got %s", e.Topic)
		}
		if e.Data != "hello" {
			t.Errorf("expected data hello, got %v", e.Data)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for event")
	}
}

func TestSubscribeAsync(t *testing.T) {
	bus := New(WithAsyncBufferSize(10))
	defer bus.Close()

	var mu sync.Mutex
	count := 0
	bus.SubscribeAsync("test.async", func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Publish("test.async", 1)
	bus.Publish("test.async", 2)
	bus.Publish("test.async", 3)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if count != 3 {
		t.Errorf("expected 3 events, got %d", count)
	}
	mu.Unlock()
}

func TestWildcardSubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received []string
	var mu sync.Mutex

	bus.Subscribe("workflow:*", func(e Event) {
		mu.Lock()
		received = append(received, string(e.Topic))
		mu.Unlock()
	})

	bus.Publish("workflow.created", nil)
	bus.Publish("workflow.updated", nil)
	bus.Publish("compile.started", nil)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(received) != 2 {
		t.Errorf("expected 2 events matching workflow:*, got %d", len(received))
	}
	mu.Unlock()
}

func TestSubscribeAll(t *testing.T) {
	bus := New()
	defer bus.Close()

	var mu sync.Mutex
	count := 0

	bus.Subscribe("*", func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Publish("any.topic", nil)
	bus.Publish("another.topic", nil)
	bus.Publish("yet.another", nil)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if count != 3 {
		t.Errorf("expected 3 events, got %d", count)
	}
	mu.Unlock()
}

func TestUnsubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	received := make(chan Event, 1)
	sub := bus.Subscribe("test.unsub", func(e Event) {
		received <- e
	})

	bus.Publish("test.unsub", "first")

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Error("timed out waiting for first event")
	}

	bus.Unsubscribe(sub)

	bus.Publish("test.unsub", "second")

	select {
	case <-received:
		t.Error("received event after unsubscribe")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestHasSubscribers(t *testing.T) {
	bus := New()
	defer bus.Close()

	if bus.HasSubscribers("test.topic") {
		t.Error("expected no subscribers initially")
	}

	bus.Subscribe("test.topic", func(e Event) {})

	if !bus.HasSubscribers("test.topic") {
		t.Error("expected subscribers after subscribe")
	}
}

func TestSubscriberCount(t *testing.T) {
	bus := New()
	defer bus.Close()

	bus.Subscribe("test.count", func(e Event) {})
	bus.Subscribe("test.count", func(e Event) {})

	if count := bus.SubscriberCount("test.count"); count != 2 {
		t.Errorf("expected 2 subscribers, got %d", count)
	}
}

func TestConcurrentPublish(t *testing.T) {
	bus := New()
	defer bus.Close()

	var mu sync.Mutex
	count := 0

	bus.SubscribeAsync("test.concurrent", func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish("test.concurrent", i)
		}()
	}
	wg.Wait()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if count != 100 {
		t.Errorf("expected 100 events, got %d", count)
	}
	mu.Unlock()
}

func TestClear(t *testing.T) {
	bus := New()
	defer bus.Close()

	bus.Subscribe("test.clear", func(e Event) {})
	bus.Subscribe("test.clear", func(e Event) {})

	if count := bus.SubscriberCount("test.clear"); count != 2 {
		t.Errorf("expected 2 subscribers before clear, got %d", count)
	}

	bus.Clear()

	if count := bus.SubscriberCount("test.clear"); count != 0 {
		t.Errorf("expected 0 subscribers after clear, got %d", count)
	}
}
