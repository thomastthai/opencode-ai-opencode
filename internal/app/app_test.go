package app

import (
	"context"
	"database/sql"
	"testing"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/agent"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/permission"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/opencode-ai/opencode/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services
type mockSessionService struct {
	mock.Mock
}

func (m *mockSessionService) Create(ctx context.Context, title string) (session.Session, error) {
	args := m.Called(ctx, title)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *mockSessionService) CreateTaskSession(ctx context.Context, toolCallID, parentSessionID, title string) (session.Session, error) {
	args := m.Called(ctx, toolCallID, parentSessionID, title)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *mockSessionService) CreateTitleSession(ctx context.Context, parentSessionID string) (session.Session, error) {
	args := m.Called(ctx, parentSessionID)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *mockSessionService) Save(ctx context.Context, s session.Session) (session.Session, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *mockSessionService) List(ctx context.Context) ([]session.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]session.Session), args.Error(1)
}

func (m *mockSessionService) Get(ctx context.Context, id string) (session.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *mockSessionService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSessionService) Subscribe(ctx context.Context) <-chan pubsub.Event[session.Session] {
	args := m.Called(ctx)
	return args.Get(0).(<-chan pubsub.Event[session.Session])
}

func (m *mockSessionService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockAgentService struct {
	mock.Mock
}

func (m *mockAgentService) Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error) {
	args := m.Called(agentName, modelID)
	return args.Get(0).(models.Model), args.Error(1)
}

func (m *mockAgentService) Summarize(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockAgentService) Model() models.Model {
	args := m.Called()
	return args.Get(0).(models.Model)
}

func (m *mockAgentService) IsSessionBusy(sessionID string) bool {
	args := m.Called(sessionID)
	return args.Bool(0)
}

func (m *mockAgentService) IsBusy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockAgentService) Run(ctx context.Context, sessionID string, prompt string, attachments ...message.Attachment) (<-chan agent.AgentEvent, error) {
	args := m.Called(ctx, sessionID, prompt, attachments)
	return args.Get(0).(<-chan agent.AgentEvent), args.Error(1)
}

type mockPermissionService struct {
	mock.Mock
}

func (m *mockPermissionService) AutoApproveSession(sessionID string) {
	m.Called(sessionID)
}

func (m *mockPermissionService) Request(req permission.CreatePermissionRequest) bool {
	args := m.Called(req)
	return args.Bool(0)
}

func (m *mockPermissionService) GrantPersistant(req permission.PermissionRequest) {
	m.Called(req)
}

func (m *mockPermissionService) Grant(req permission.PermissionRequest) {
	m.Called(req)
}

func (m *mockPermissionService) Deny(req permission.PermissionRequest) {
	m.Called(req)
}

func (m *mockPermissionService) Subscribe(ctx context.Context) <-chan pubsub.Event[permission.PermissionRequest] {
	args := m.Called(ctx)
	return args.Get(0).(<-chan pubsub.Event[permission.PermissionRequest])
}

func (m *mockAgentService) Cancel(sessionID string) {
	m.Called(sessionID)
}

func (m *mockAgentService) Subscribe(ctx context.Context) <-chan pubsub.Event[agent.AgentEvent] {
	args := m.Called(ctx)
	return args.Get(0).(<-chan pubsub.Event[agent.AgentEvent])
}

func TestRunNonInteractive(t *testing.T) {
	// Load config
	_, err := config.Load(".", true)
	assert.NoError(t, err)

	// Create a new App with mock services
	app, err := New(context.Background(), &sql.DB{}, true)
	assert.NoError(t, err)

	mockSessions := new(mockSessionService)
	mockAgent := new(mockAgentService)
	mockPermissions := new(mockPermissionService)

	app.Sessions = mockSessions
	app.CoderAgent = mockAgent
	app.Permissions = mockPermissions

	// Test case
	prompt := "test prompt"
	expectedSession := session.Session{ID: "test-session"}
	expectedMessage := message.Message{
		ID:        "test-message",
		SessionID: "test-session",
		Role:      message.User,
		Parts:     []message.ContentPart{message.TextContent{Text: "test response"}},
		Model:     "test-model",
	}

	// Set up mock expectations
	mockSessions.On("Create", mock.Anything, mock.Anything).Return(expectedSession, nil)
	mockPermissions.On("AutoApproveSession", expectedSession.ID).Return()
	mockAgent.On("Model").Return(models.Model{}).Maybe()
	mockAgent.On("Run", mock.Anything, expectedSession.ID, prompt, mock.Anything).Return(createResultChannel(agent.AgentEvent{Message: expectedMessage}), nil)

	// Run the function
	err = app.RunNonInteractive(context.Background(), prompt, "text", true)

	// Assertions
	assert.NoError(t, err)
	mockSessions.AssertExpectations(t)
	mockPermissions.AssertExpectations(t)
	mockAgent.AssertExpectations(t)
}


func createResultChannel(result agent.AgentEvent) <-chan agent.AgentEvent {
	ch := make(chan agent.AgentEvent, 1)
	ch <- result
	close(ch)
	return ch
}
