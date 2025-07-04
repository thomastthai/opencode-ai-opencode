
package agent

import (
	"context"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/provider"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/opencode-ai/opencode/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) Create(ctx context.Context, title string) (session.Session, error) {
	args := m.Called(ctx, title)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockSessionService) CreateTaskSession(ctx context.Context, toolCallID, parentSessionID, title string) (session.Session, error) {
	args := m.Called(ctx, toolCallID, parentSessionID, title)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockSessionService) CreateTitleSession(ctx context.Context, parentID string) (session.Session, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockSessionService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionService) Get(ctx context.Context, id string) (session.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockSessionService) Save(ctx context.Context, s session.Session) (session.Session, error) {
	args := m.Called(ctx, s)
	return s, args.Error(1)
}

func (m *MockSessionService) List(ctx context.Context) ([]session.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]session.Session), args.Error(1)
}

func (m *MockSessionService) Subscribe(ctx context.Context) <-chan pubsub.Event[session.Session] {
	args := m.Called(ctx)
	return args.Get(0).(<-chan pubsub.Event[session.Session])
}

type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) Create(ctx context.Context, sessionID string, params message.CreateMessageParams) (message.Message, error) {
	args := m.Called(ctx, sessionID, params)
	return args.Get(0).(message.Message), args.Error(1)
}

func (m *MockMessageService) Update(ctx context.Context, msg message.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageService) Get(ctx context.Context, id string) (message.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(message.Message), args.Error(1)
}

func (m *MockMessageService) List(ctx context.Context, sessionID string) ([]message.Message, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]message.Message), args.Error(1)
}

func (m *MockMessageService) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockMessageService) Subscribe(ctx context.Context) <-chan pubsub.Event[message.Message] {
	args := m.Called(ctx)
	return args.Get(0).(<-chan pubsub.Event[message.Message])
}

type MockTool struct {
	mock.Mock
}

func (m *MockTool) Info() tools.ToolInfo {
	args := m.Called()
	return args.Get(0).(tools.ToolInfo)
}

func (m *MockTool) Run(ctx context.Context, call tools.ToolCall) (tools.ToolResponse, error) {
	args := m.Called(ctx, call)
	return args.Get(0).(tools.ToolResponse), args.Error(1)
}

func TestNewAgent(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)

	// Mock the provider registry
	provider.RegisterProvider(models.ProviderOpenAI, func(cfg provider.ProviderConfig) (provider.Provider, error) {
		return &provider.MockProvider{}, nil
	}, provider.ProviderInfo{})

	models.SupportedModels["gpt-4"] = models.Model{
		ID:       "gpt-4",
		Provider: models.ProviderOpenAI,
	}
	models.SupportedModels["gpt-3.5-turbo"] = models.Model{
		ID:       "gpt-3.5-turbo",
		Provider: models.ProviderOpenAI,
	}

	cfg := &config.Config{
		Agents: map[config.AgentName]config.Agent{
			config.AgentCoder: {
				Model:     "gpt-4",
				MaxTokens: 1024,
			},
			config.AgentTitle: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
			config.AgentSummarizer: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
		},
		Providers: map[models.ModelProvider]config.Provider{
			models.ProviderOpenAI: {
				APIKey: "test-key",
			},
		},
	}
	config.Set(cfg)

	agent, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, nil)
	assert.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestAgent_Model(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)

	// Mock the provider registry
	provider.RegisterProvider(models.ProviderOpenAI, func(cfg provider.ProviderConfig) (provider.Provider, error) {
		return &provider.MockProvider{}, nil
	}, provider.ProviderInfo{})

	models.SupportedModels["gpt-4"] = models.Model{
		ID:       "gpt-4",
		Provider: models.ProviderOpenAI,
	}
	models.SupportedModels["gpt-3.5-turbo"] = models.Model{
		ID:       "gpt-3.5-turbo",
		Provider: models.ProviderOpenAI,
	}

	cfg := &config.Config{
		Agents: map[config.AgentName]config.Agent{
			config.AgentCoder: {
				Model:     "gpt-4",
				MaxTokens: 1024,
			},
			config.AgentTitle: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
			config.AgentSummarizer: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
		},
		Providers: map[models.ModelProvider]config.Provider{
			models.ProviderOpenAI: {
				APIKey: "test-key",
			},
		},
	}
	config.Set(cfg)

	agent, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, nil)
	assert.NoError(t, err)
	assert.NotNil(t, agent)

	model := agent.Model()
	assert.Equal(t, models.ModelID("gpt-4"), model.ID)
}

func TestAgent_IsBusy(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)

	// Mock the provider registry
	provider.RegisterProvider(models.ProviderOpenAI, func(cfg provider.ProviderConfig) (provider.Provider, error) {
		return &provider.MockProvider{}, nil
	}, provider.ProviderInfo{})

	models.SupportedModels["gpt-4"] = models.Model{
		ID:       "gpt-4",
		Provider: models.ProviderOpenAI,
	}
	models.SupportedModels["gpt-3.5-turbo"] = models.Model{
		ID:       "gpt-3.5-turbo",
		Provider: models.ProviderOpenAI,
	}

	cfg := &config.Config{
		Agents: map[config.AgentName]config.Agent{
			config.AgentCoder: {
				Model:     "gpt-4",
				MaxTokens: 1024,
			},
			config.AgentTitle: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
			config.AgentSummarizer: {
				Model:     "gpt-3.5-turbo",
				MaxTokens: 1024,
			},
		},
		Providers: map[models.ModelProvider]config.Provider{
			models.ProviderOpenAI: {
				APIKey: "test-key",
			},
		},
	}
	config.Set(cfg)

	agentSvc, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, nil)
	assert.NoError(t, err)
	assert.NotNil(t, agentSvc)

	assert.False(t, agentSvc.IsBusy())
	assert.False(t, agentSvc.IsSessionBusy("session-1"))

	// Simulate a busy session
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	agentImpl := agentSvc.(*agent)
	agentImpl.ActiveRequests.Store("session-1", cancel)

	assert.True(t, agentSvc.IsBusy())
	assert.True(t, agentSvc.IsSessionBusy("session-1"))
	assert.False(t, agentSvc.IsSessionBusy("session-2"))

	// Cancel the request and check again
	agentSvc.Cancel("session-1")
	assert.False(t, agentSvc.IsBusy())
	assert.False(t, agentSvc.IsSessionBusy("session-1"))
}

func TestAgent_Run(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)

	mockProviderConfig := &provider.MockConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			Model: models.Model{
				ID: "gpt-4",
			},
			MaxTokens: 1024,
		},
		StreamSupport: true,
		StreamEvents: []provider.ProviderEvent{
			{
				Type: provider.EventComplete,
				Response: &provider.ProviderResponse{
					FinishReason: message.FinishReasonEndTurn,
				},
			},
		},
	}

	mockProvider, err := provider.NewMockProvider(mockProviderConfig)
	assert.NoError(t, err)

	agentSvc, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, nil, mockProvider, mockProvider, mockProvider)
	assert.NoError(t, err)
	assert.NotNil(t, agentSvc)

	sessionID := "session-1"
	prompt := "Hello"

	mockSessions.On("Get", mock.Anything, sessionID).Return(session.Session{ID: sessionID}, nil)
	mockSessions.On("Save", mock.Anything, mock.Anything).Return(nil, nil)
	mockMessages.On("List", mock.Anything, sessionID).Return([]message.Message{}, nil)
	mockMessages.On("Create", mock.Anything, sessionID, mock.Anything).Return(message.Message{ID: "msg-1", Role: message.User}, nil).Once()
	mockMessages.On("Create", mock.Anything, sessionID, mock.Anything).Return(message.Message{ID: "msg-2", Role: message.Assistant}, nil).Once()
	mockMessages.On("Update", mock.Anything, mock.Anything).Return(nil)

	events, err := agentSvc.Run(context.Background(), sessionID, prompt)
	assert.NoError(t, err)

	select {
	case event := <-events:
		assert.Equal(t, AgentEventTypeResponse, event.Type)
		assert.NoError(t, event.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for agent event")
	}
}

func TestAgent_Run_ProviderError(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)

	mockProviderConfig := &provider.MockConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			Model: models.Model{
				ID: "gpt-4",
			},
			MaxTokens: 1024,
		},
		StreamSupport: true,
		ErrorToReturn: "provider error",
	}

	mockProvider, err := provider.NewMockProvider(mockProviderConfig)
	assert.NoError(t, err)

	agentSvc, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, nil, mockProvider, mockProvider, mockProvider)
	assert.NoError(t, err)
	assert.NotNil(t, agentSvc)

	sessionID := "session-1"
	prompt := "Hello"

	mockSessions.On("Get", mock.Anything, sessionID).Return(session.Session{ID: sessionID}, nil)
	mockMessages.On("List", mock.Anything, sessionID).Return([]message.Message{}, nil)
	mockMessages.On("Create", mock.Anything, sessionID, mock.Anything).Return(message.Message{ID: "msg-1", Role: message.User}, nil).Once()
	mockMessages.On("Create", mock.Anything, sessionID, mock.Anything).Return(message.Message{ID: "msg-2", Role: message.Assistant}, nil).Once()
	mockMessages.On("Update", mock.Anything, mock.Anything).Return(nil)

	events, err := agentSvc.Run(context.Background(), sessionID, prompt)
	assert.NoError(t, err)

	select {
	case event := <-events:
		assert.Equal(t, AgentEventTypeError, event.Type)
		assert.Error(t, event.Error)
		assert.Contains(t, event.Error.Error(), "provider error")
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for agent event")
	}
}

func TestAgent_Run_ToolCall(t *testing.T) {
	mockSessions := new(MockSessionService)
	mockMessages := new(MockMessageService)
	mockTool := new(MockTool)

	mockTool.On("Info").Return(tools.ToolInfo{Name: "test-tool"})
	mockTool.On("Run", mock.Anything, mock.Anything).Return(tools.ToolResponse{Content: "tool result"}, nil)

	mockProviderConfig := &provider.MockConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			Model: models.Model{
				ID: "gpt-4",
			},
			MaxTokens: 1024,
		},
		StreamSupport: true,
		StreamEvents: []provider.ProviderEvent{
			{
				Type: provider.EventToolUseStart,
				ToolCall: &message.ToolCall{
					ID:   "tool-1",
					Name: "test-tool",
				},
			},
			{
				Type: provider.EventComplete,
				Response: &provider.ProviderResponse{
					FinishReason: message.FinishReasonToolUse,
					ToolCalls: []message.ToolCall{
						{
							ID:   "tool-1",
							Name: "test-tool",
						},
					},
				},
			},
		},
	}

	// This mock provider will be used for the second call to the LLM, after the tool result has been processed.
	mockProviderAfterTool := &provider.MockConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			Model: models.Model{
				ID: "gpt-4",
			},
			MaxTokens: 1024,
		},
		StreamSupport: true,
		StreamEvents: []provider.ProviderEvent{
			{
				Type: provider.EventComplete,
				Response: &provider.ProviderResponse{
					FinishReason: message.FinishReasonEndTurn,
				},
			},
		},
	}

	mockProvider, err := provider.NewMockProvider(mockProviderConfig)
	assert.NoError(t, err)

	mockProvider2, err := provider.NewMockProvider(mockProviderAfterTool)
	assert.NoError(t, err)

	agentSvc, err := NewAgent(config.AgentCoder, mockSessions, mockMessages, []tools.BaseTool{mockTool}, mockProvider, mockProvider2, mockProvider)
	assert.NoError(t, err)
	assert.NotNil(t, agentSvc)

	sessionID := "session-1"
	prompt := "Hello"

	mockSessions.On("Get", mock.Anything, sessionID).Return(session.Session{ID: sessionID}, nil)
	mockSessions.On("Save", mock.Anything, mock.Anything).Return(nil, nil)
	mockMessages.On("List", mock.Anything, sessionID).Return([]message.Message{}, nil).Once()
	mockMessages.On("List", mock.Anything, sessionID).Return([]message.Message{
		{ID: "msg-1", Role: message.User},
		{ID: "msg-2", Role: message.Assistant, Parts: []message.ContentPart{message.ToolCall{ID: "tool-1", Name: "test-tool"}}},
		{ID: "msg-3", Role: message.Tool, Parts: []message.ContentPart{message.ToolResult{ToolCallID: "tool-1", Content: "tool result"}}},
	}, nil).Once()

	mockMessages.On("Create", mock.Anything, sessionID, mock.MatchedBy(func(params message.CreateMessageParams) bool {
		return params.Role == message.User
	})).Return(message.Message{ID: "msg-1", Role: message.User}, nil).Once()

	mockMessages.On("Create", mock.Anything, sessionID, mock.MatchedBy(func(params message.CreateMessageParams) bool {
		return params.Role == message.Assistant
	})).Return(message.Message{ID: "msg-2", Role: message.Assistant}, nil).Twice()

	mockMessages.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(params message.CreateMessageParams) bool {
		return params.Role == message.Tool
	})).Return(message.Message{ID: "msg-3", Role: message.Tool}, nil).Once()

	mockMessages.On("Update", mock.Anything, mock.Anything).Return(nil)

	events, err := agentSvc.Run(context.Background(), sessionID, prompt)
	assert.NoError(t, err)

	// We expect two events: one for the tool call, and one for the final response
	eventCount := 0
	for event := range events {
		eventCount++
		if eventCount == 1 {
			assert.Equal(t, AgentEventTypeToolCall, event.Type)
			assert.NoError(t, event.Error)
		} else if eventCount == 2 {
			assert.Equal(t, AgentEventTypeResponse, event.Type)
			assert.NoError(t, event.Error)
		}
	}
	assert.Equal(t, 2, eventCount)
}
