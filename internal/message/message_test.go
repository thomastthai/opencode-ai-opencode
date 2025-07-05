package message

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/opencode-ai/opencode/internal/db"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock database querier
type MockQuerier struct {
	mock.Mock
}

// Message methods
func (m *MockQuerier) CreateMessage(ctx context.Context, params db.CreateMessageParams) (db.Message, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Message), args.Error(1)
}

func (m *MockQuerier) UpdateMessage(ctx context.Context, params db.UpdateMessageParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Message), args.Error(1)
}

func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]db.Message), args.Error(1)
}

func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Required interface methods (stubs)
func (m *MockQuerier) CreateFile(ctx context.Context, arg db.CreateFileParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) { return db.Session{}, nil }
func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error { return nil }
func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error { return nil }
func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error { return nil }
func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, arg db.GetFileByPathAndSessionParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) { return db.Session{}, nil }
func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) { return []db.Session{}, nil }
func (m *MockQuerier) UpdateFile(ctx context.Context, arg db.UpdateFileParams) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) UpdateSession(ctx context.Context, arg db.UpdateSessionParams) (db.Session, error) { return db.Session{}, nil }

func TestNewService(t *testing.T) {
	mockQuerier := new(MockQuerier)
	
	service := NewService(mockQuerier)
	
	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

func TestService_Create_UserMessage(t *testing.T) {
	mockQuerier := new(MockQuerier)
	service := NewService(mockQuerier)
	
	sessionID := uuid.New().String()
	params := CreateMessageParams{
		Role: User,
		Parts: []ContentPart{
			TextContent{Text: "Hello, world!"},
		},
		Model: models.GPT41,
	}
	
	expectedDBMessage := db.Message{
		ID:        "test-message-id",
		SessionID: sessionID,
		Role:      string(User),
		Parts:     `[{"type":"text","data":{"text":"Hello, world!"}},{"type":"finish","data":{"reason":"stop","time":0}}]`,
		Model:     sql.NullString{String: string(models.GPT41), Valid: true},
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}
	
	mockQuerier.On("CreateMessage", mock.Anything, mock.Anything).Return(expectedDBMessage, nil)
	
	message, err := service.Create(context.Background(), sessionID, params)
	
	assert.NoError(t, err)
	assert.Equal(t, "test-message-id", message.ID)
	assert.Equal(t, sessionID, message.SessionID)
	assert.Equal(t, User, message.Role)
	assert.Equal(t, models.GPT41, message.Model)
	
	mockQuerier.AssertExpectations(t)
}