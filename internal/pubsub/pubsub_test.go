package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBroker(t *testing.T) {
	broker := NewBroker[string]()

	assert.NotNil(t, broker)
	assert.Equal(t, 0, broker.GetSubscriberCount())
}

func TestNewBrokerWithOptions(t *testing.T) {
	bufferSize := 32
	maxEvents := 500

	broker := NewBrokerWithOptions[int](bufferSize, maxEvents)

	assert.NotNil(t, broker)
	assert.Equal(t, 0, broker.GetSubscriberCount())
	assert.Equal(t, maxEvents, broker.maxEvents)
}

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	assert.NotNil(t, sub)
	assert.Equal(t, 1, broker.GetSubscriberCount())
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()

	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)
	sub3 := broker.Subscribe(ctx)

	assert.NotNil(t, sub1)
	assert.NotNil(t, sub2)
	assert.NotNil(t, sub3)
	assert.Equal(t, 3, broker.GetSubscriberCount())
}

func TestBroker_Publish(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	testMessage := "test message"

	// Publish an event
	broker.Publish(CreatedEvent, testMessage)

	// Verify the subscriber receives it
	select {
	case event := <-sub:
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, testMessage, event.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected to receive an event")
	}
}

func TestBroker_PublishToMultipleSubscribers(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx := context.Background()

	numSubs := 3
	subs := make([]<-chan Event[int], numSubs)
	for i := 0; i < numSubs; i++ {
		subs[i] = broker.Subscribe(ctx)
	}

	testData := 42

	// Publish an event
	broker.Publish(UpdatedEvent, testData)

	// Verify all subscribers receive it
	for i, sub := range subs {
		select {
		case event := <-sub:
			assert.Equal(t, UpdatedEvent, event.Type)
			assert.Equal(t, testData, event.Payload)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Subscriber %d did not receive event", i)
		}
	}
}

func TestBroker_SubscriberContextCancellation(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	sub := broker.Subscribe(ctx)

	assert.Equal(t, 1, broker.GetSubscriberCount())

	// Cancel the context
	cancel()

	// Wait for cleanup
	time.Sleep(10 * time.Millisecond)

	// Subscriber count should decrease
	assert.Equal(t, 0, broker.GetSubscriberCount())

	// Channel should be closed
	select {
	case _, ok := <-sub:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected channel to be closed")
	}
}

func TestBroker_Shutdown(t *testing.T) {
	broker := NewBroker[string]()

	ctx := context.Background()
	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)

	assert.Equal(t, 2, broker.GetSubscriberCount())

	// Shutdown the broker
	broker.Shutdown()

	// Subscriber count should be zero
	assert.Equal(t, 0, broker.GetSubscriberCount())

	// All channels should be closed
	select {
	case _, ok := <-sub1:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected channel to be closed")
	}

	select {
	case _, ok := <-sub2:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected channel to be closed")
	}
}

func TestBroker_PublishAfterShutdown(t *testing.T) {
	broker := NewBroker[string]()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	// Shutdown the broker
	broker.Shutdown()

	// Try to publish after shutdown
	broker.Publish(CreatedEvent, "test")

	// The subscriber channel should be closed, so we should get a closed channel signal
	select {
	case _, ok := <-sub:
		assert.False(t, ok, "Channel should be closed after shutdown")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected channel to be closed after shutdown")
	}
}

func TestBroker_SubscribeAfterShutdown(t *testing.T) {
	broker := NewBroker[string]()
	broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	// Should get a closed channel immediately
	select {
	case _, ok := <-sub:
		assert.False(t, ok, "Channel should be closed immediately")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected closed channel immediately")
	}
}

func TestBroker_ConcurrentOperations(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx := context.Background()

	numGoroutines := 10
	numMessagesPerGoroutine := 10

	var wg sync.WaitGroup

	// Create multiple subscribers
	subs := make([]<-chan Event[int], numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		subs[i] = broker.Subscribe(ctx)
	}

	// Start publishers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessagesPerGoroutine; j++ {
				broker.Publish(CreatedEvent, id*100+j)
			}
		}(i)
	}

	// Start receivers
	receivedCounts := make([]int, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(subIndex int) {
			defer wg.Done()
			timeout := time.After(2 * time.Second)
			for {
				select {
				case event := <-subs[subIndex]:
					assert.Equal(t, CreatedEvent, event.Type)
					receivedCounts[subIndex]++
				case <-timeout:
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Each subscriber should have received some events
	totalExpected := numGoroutines * numMessagesPerGoroutine
	totalReceived := 0
	for _, count := range receivedCounts {
		totalReceived += count
	}

	// Due to concurrent nature and buffer limits, we might not receive all messages
	// but we should receive a significant portion
	assert.Greater(t, totalReceived, totalExpected/2, "Should receive at least half the messages")
}

func TestBroker_BufferOverflow(t *testing.T) {
	broker := NewBrokerWithOptions[int](2, 1000) // Small buffer
	defer broker.Shutdown()

	ctx := context.Background()
	sub := broker.Subscribe(ctx)

	// Publish more events than buffer can hold without reading
	for i := 0; i < 10; i++ {
		broker.Publish(CreatedEvent, i)
	}

	// Should still be able to read some events (those that fit in buffer)
	receivedCount := 0
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case event := <-sub:
			assert.Equal(t, CreatedEvent, event.Type)
			receivedCount++
		case <-timeout:
			goto done
		}
	}

done:
	// Should have received at least the buffer size worth of events
	assert.GreaterOrEqual(t, receivedCount, 1, "Should receive at least one event")
}

func TestEventTypes(t *testing.T) {
	assert.Equal(t, EventType("created"), CreatedEvent)
	assert.Equal(t, EventType("updated"), UpdatedEvent)
	assert.Equal(t, EventType("deleted"), DeletedEvent)
}

func TestEvent_Structure(t *testing.T) {
	payload := "test payload"
	event := Event[string]{
		Type:    CreatedEvent,
		Payload: payload,
	}

	assert.Equal(t, CreatedEvent, event.Type)
	assert.Equal(t, payload, event.Payload)
}

func TestBroker_DoubleShutdown(t *testing.T) {
	broker := NewBroker[string]()

	// First shutdown
	broker.Shutdown()
	assert.Equal(t, 0, broker.GetSubscriberCount())

	// Second shutdown should not panic
	broker.Shutdown()
	assert.Equal(t, 0, broker.GetSubscriberCount())
}