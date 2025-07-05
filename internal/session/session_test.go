package session

import (
	"context"
	"database/sql"
	"testing"

	"github.com/opencode-ai/opencode/internal/db"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock database querier for session package
type MockQuerier struct {
	mock.Mock
}

// Session methods
func (m *MockQuerier) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) UpdateSession(ctx context.Context, params db.UpdateSessionParams) (db.Session, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Session), args.Error(1)
}

func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Stub methods for db.Querier interface compliance
func (m *MockQuerier) CreateFile(ctx context.Context, arg db.CreateFileParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) CreateMessage(ctx context.Context, arg db.CreateMessageParams) (db.Message, error) { return db.Message{}, nil }
func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error { return nil }
func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error { return nil }
func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error { return nil }
func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error { return nil }
func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, arg db.GetFileByPathAndSessionParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) { return db.Message{}, nil }
func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) { return []db.Message{}, nil }
func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) UpdateFile(ctx context.Context, arg db.UpdateFileParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) UpdateMessage(ctx context.Context, arg db.UpdateMessageParams) error { return nil }

func TestNewService(t *testing.T) {
	mockQuerier := new(MockQuerier)
	
	service := NewService(mockQuerier)
	
	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

func TestService_Create(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	title := "Test Session"
	
	expectedDBSession := db.Session{
		ID:    "test-session-id",
		Title: title,
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}
	
	mockQuerier.On("CreateSession", mock.Anything, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.Title == title && params.ID != ""
	})).Return(expectedDBSession, nil)
	
	session, err := service.Create(context.Background(), title)
	
	assert.NoError(t, err)
	assert.Equal(t, "test-session-id", session.ID)
	assert.Equal(t, title, session.Title)
	assert.Equal(t, int64(1234567890), session.CreatedAt)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_CreateTaskSession(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	toolCallID := "tool-call-123"
	parentSessionID := "parent-session-123"
	title := "Task Session"
	
	expectedDBSession := db.Session{
		ID:              toolCallID,
		ParentSessionID: sql.NullString{String: parentSessionID, Valid: true},
		Title:           title,
		CreatedAt:       1234567890,
		UpdatedAt:       1234567890,
	}
	
	mockQuerier.On("CreateSession", mock.Anything, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.ID == toolCallID && 
			   params.ParentSessionID.String == parentSessionID &&
			   params.ParentSessionID.Valid &&
			   params.Title == title
	})).Return(expectedDBSession, nil)
	
	session, err := service.CreateTaskSession(context.Background(), toolCallID, parentSessionID, title)
	
	assert.NoError(t, err)
	assert.Equal(t, toolCallID, session.ID)
	assert.Equal(t, parentSessionID, session.ParentSessionID)
	assert.Equal(t, title, session.Title)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_CreateTitleSession(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	parentSessionID := "parent-session-123"
	expectedTitleSessionID := "title-" + parentSessionID
	
	expectedDBSession := db.Session{
		ID:              expectedTitleSessionID,
		ParentSessionID: sql.NullString{String: parentSessionID, Valid: true},
		Title:           "Generate a title",
		CreatedAt:       1234567890,
		UpdatedAt:       1234567890,
	}
	
	mockQuerier.On("CreateSession", mock.Anything, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.ID == expectedTitleSessionID &&
			   params.ParentSessionID.String == parentSessionID &&
			   params.Title == "Generate a title"
	})).Return(expectedDBSession, nil)
	
	session, err := service.CreateTitleSession(context.Background(), parentSessionID)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedTitleSessionID, session.ID)
	assert.Equal(t, parentSessionID, session.ParentSessionID)
	assert.Equal(t, "Generate a title", session.Title)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_Get(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	sessionID := "test-session-id"
	
	dbSession := db.Session{
		ID:               sessionID,
		Title:            "Test Session",
		MessageCount:     5,
		PromptTokens:     100,
		CompletionTokens: 200,
		Cost:             0.05,
		CreatedAt:        1234567890,
		UpdatedAt:        1234567890,
	}
	
	mockQuerier.On("GetSessionByID", mock.Anything, sessionID).Return(dbSession, nil)
	
	session, err := service.Get(context.Background(), sessionID)
	
	assert.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, "Test Session", session.Title)
	assert.Equal(t, int64(5), session.MessageCount)
	assert.Equal(t, int64(100), session.PromptTokens)
	assert.Equal(t, int64(200), session.CompletionTokens)
	assert.Equal(t, 0.05, session.Cost)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_Save(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	session := Session{
		ID:               "test-session-id",
		Title:            "Updated Session",
		PromptTokens:     150,
		CompletionTokens: 250,
		SummaryMessageID: "summary-msg-123",
		Cost:             0.10,
	}
	
	updatedDBSession := db.Session{
		ID:               session.ID,
		Title:            session.Title,
		PromptTokens:     session.PromptTokens,
		CompletionTokens: session.CompletionTokens,
		SummaryMessageID: sql.NullString{String: session.SummaryMessageID, Valid: true},
		Cost:             session.Cost,
		UpdatedAt:        1234567891,
	}
	
	mockQuerier.On("UpdateSession", mock.Anything, mock.MatchedBy(func(params db.UpdateSessionParams) bool {
		return params.ID == session.ID &&
			   params.Title == session.Title &&
			   params.PromptTokens == session.PromptTokens &&
			   params.CompletionTokens == session.CompletionTokens &&
			   params.SummaryMessageID.String == session.SummaryMessageID &&
			   params.SummaryMessageID.Valid &&
			   params.Cost == session.Cost
	})).Return(updatedDBSession, nil)
	
	savedSession, err := service.Save(context.Background(), session)
	
	assert.NoError(t, err)
	assert.Equal(t, session.ID, savedSession.ID)
	assert.Equal(t, session.Title, savedSession.Title)
	assert.Equal(t, int64(1234567891), savedSession.UpdatedAt)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_List(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	dbSessions := []db.Session{
		{
			ID:    "session-1",
			Title: "First Session",
			CreatedAt: 1234567890,
		},
		{
			ID:    "session-2",
			Title: "Second Session", 
			CreatedAt: 1234567891,
		},
	}
	
	mockQuerier.On("ListSessions", mock.Anything).Return(dbSessions, nil)
	
	sessions, err := service.List(context.Background())
	
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "session-1", sessions[0].ID)
	assert.Equal(t, "First Session", sessions[0].Title)
	assert.Equal(t, "session-2", sessions[1].ID)
	assert.Equal(t, "Second Session", sessions[1].Title)
	
	mockQuerier.AssertExpectations(t)
}

func TestService_Delete(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	sessionID := "test-session-id"
	
	dbSession := db.Session{
		ID:    sessionID,
		Title: "Session to Delete",
		CreatedAt: 1234567890,
	}
	
	mockQuerier.On("GetSessionByID", mock.Anything, sessionID).Return(dbSession, nil)
	mockQuerier.On("DeleteSession", mock.Anything, sessionID).Return(nil)
	
	err := service.Delete(context.Background(), sessionID)
	
	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

func TestService_PubSubIntegration(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	// Subscribe to events
	ctx := context.Background()
	eventsChan := service.Subscribe(ctx)
	
	// Create a session and verify event is published
	title := "Test Session"
	
	expectedDBSession := db.Session{
		ID:    "test-session-id",
		Title: title,
		CreatedAt: 1234567890,
	}
	
	mockQuerier.On("CreateSession", mock.Anything, mock.Anything).Return(expectedDBSession, nil)
	
	session, err := service.Create(ctx, title)
	assert.NoError(t, err)
	
	// Check that event was published
	select {
	case event := <-eventsChan:
		assert.Equal(t, pubsub.CreatedEvent, event.Type)
		assert.Equal(t, session.ID, event.Payload.ID)
		assert.Equal(t, title, event.Payload.Title)
	default:
		t.Fatal("Expected to receive a created event")
	}
	
	mockQuerier.AssertExpectations(t)
}