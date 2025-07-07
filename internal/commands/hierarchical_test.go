package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHierarchicalRegistry(t *testing.T) {
	t.Run("register and retrieve topics", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Register a topic
		topic := &Topic{
			ID:          "test",
			Name:        "Test",
			Description: "Test commands",
			Icon:        "🧪",
		}
		err := registry.RegisterTopic(topic)
		assert.NoError(t, err)
		
		// Retrieve topic
		retrieved, exists := registry.GetTopic("test")
		assert.True(t, exists)
		assert.Equal(t, "test", retrieved.ID)
		assert.Equal(t, "Test", retrieved.Name)
		
		// Non-existent topic
		_, exists = registry.GetTopic("nonexistent")
		assert.False(t, exists)
	})
	
	t.Run("register and retrieve commands", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Register topic first
		topic := &Topic{
			ID:   "test",
			Name: "Test",
		}
		registry.RegisterTopic(topic)
		
		// Register command
		command := &HierCommand{
			ID:          "run",
			Name:        "Run",
			Description: "Run test",
			MinArgs:     0,
			MaxArgs:     1,
		}
		err := registry.RegisterCommand("test", command)
		assert.NoError(t, err)
		
		// Retrieve command
		retrieved, exists := registry.GetCommand("test", "run")
		assert.True(t, exists)
		assert.Equal(t, "run", retrieved.ID)
		assert.Equal(t, "Run", retrieved.Name)
		
		// Non-existent command
		_, exists = registry.GetCommand("test", "nonexistent")
		assert.False(t, exists)
		
		// Verb in non-existent topic
		_, exists = registry.GetCommand("nonexistent", "run")
		assert.False(t, exists)
	})
	
	t.Run("duplicate registration errors", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Register topic
		topic := &Topic{ID: "test", Name: "Test"}
		err := registry.RegisterTopic(topic)
		assert.NoError(t, err)
		
		// Duplicate topic
		err = registry.RegisterTopic(topic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
		
		// Register command
		command := &HierCommand{ID: "run", Name: "Run"}
		err = registry.RegisterCommand("test", command)
		assert.NoError(t, err)
		
		// Duplicate verb
		err = registry.RegisterCommand("test", command)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
		
		// Verb in non-existent topic
		err = registry.RegisterCommand("nonexistent", command)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("list topics", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Register multiple topics
		topics := []*Topic{
			{ID: "session", Name: "Session"},
			{ID: "config", Name: "Config"},
			{ID: "auth", Name: "Auth"},
		}
		
		for _, topic := range topics {
			registry.RegisterTopic(topic)
		}
		
		// List topics
		listed := registry.ListTopics()
		assert.Len(t, listed, 3)
		
		// Check all topics are present
		topicMap := make(map[string]bool)
		for _, t := range listed {
			topicMap[t.ID] = true
		}
		
		assert.True(t, topicMap["session"])
		assert.True(t, topicMap["config"])
		assert.True(t, topicMap["auth"])
	})
	
	t.Run("get completions for topic", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Set up topic with commands
		topic := &Topic{ID: "test", Name: "Test"}
		registry.RegisterTopic(topic)
		
		commands := []*HierCommand{
			{ID: "run", Name: "Run", Description: "Run test"},
			{ID: "stop", Name: "Stop", Description: "Stop test"},
			{ID: "status", Name: "Status", Description: "Show status"},
		}
		
		for _, command := range commands {
			registry.RegisterCommand("test", command)
		}
		
		// Get completions
		completions := registry.GetCompletionsForTopic("test")
		assert.Len(t, completions, 3)
		
		// Check completions
		for _, comp := range completions {
			assert.Contains(t, comp.Complete, "/test ")
			assert.NotEmpty(t, comp.Display)
			assert.NotEmpty(t, comp.Description)
		}
		
		// Non-existent topic
		completions = registry.GetCompletionsForTopic("nonexistent")
		assert.Empty(t, completions)
	})
	
	t.Run("command execution", func(t *testing.T) {
		registry := NewHierarchicalRegistry()
		
		// Track handler execution
		handlerCalled := false
		handlerArgs := map[string]interface{}{}
		
		// Set up topic and verb with handler
		topic := &Topic{ID: "test", Name: "Test"}
		registry.RegisterTopic(topic)
		
		command := &HierCommand{
			ID:      "run",
			Name:    "Run",
			MinArgs: 1,
			MaxArgs: 2,
			Handler: func(ctx context.Context, args map[string]interface{}) error {
				handlerCalled = true
				handlerArgs = args
				return nil
			},
		}
		registry.RegisterCommand("test", command)
		
		// Execute valid command
		cmd := SlashCommand{
			Raw:     "/test run arg1 arg2",
			Topic:   "test",
			Command: "run",
			Args:    []string{"arg1", "arg2"},
			Options: NewParsedOptions(),
		}
		
		err := registry.Execute(context.Background(), cmd)
		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, []string{"arg1", "arg2"}, handlerArgs["args"])
		assert.Equal(t, "/test run arg1 arg2", handlerArgs["raw"])
		
		// Test argument validation
		cmd.Args = []string{} // Too few args
		err = registry.Execute(context.Background(), cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient arguments")
		
		cmd.Args = []string{"1", "2", "3"} // Too many args
		err = registry.Execute(context.Background(), cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many arguments")
		
		// Non-existent command
		cmd = SlashCommand{
			Topic:   "nonexistent",
			Command: "command",
			Options: NewParsedOptions(),
		}
		err = registry.Execute(context.Background(), cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not found")
	})
}

func TestInitializeBuiltinCommands(t *testing.T) {
	registry := NewHierarchicalRegistry()
	err := InitializeBuiltinCommands(registry)
	assert.NoError(t, err)
	
	// Check that expected topics exist
	expectedTopics := []string{"session", "config", "project", "auth", "system", "help"}
	topics := registry.ListTopics()
	
	topicMap := make(map[string]bool)
	for _, topic := range topics {
		topicMap[topic.ID] = true
	}
	
	for _, expected := range expectedTopics {
		assert.True(t, topicMap[expected], "Expected topic %s to be registered", expected)
	}
	
	// Check some specific verbs
	verb, exists := registry.GetCommand("session", "new")
	assert.True(t, exists)
	assert.Equal(t, "New", verb.Name)
	
	verb, exists = registry.GetCommand("auth", "login")
	assert.True(t, exists)
	assert.Equal(t, 1, verb.MinArgs, "Login should require provider argument")
	
	verb, exists = registry.GetCommand("system", "help")
	assert.True(t, exists)
	assert.NotNil(t, verb.Handler)
}