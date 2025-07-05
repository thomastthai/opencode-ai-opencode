package permission

import (
	"context"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestNewPermissionService(t *testing.T) {
	service := NewPermissionService()

	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

func TestPermissionService_AutoApproveSession(t *testing.T) {
	service := NewPermissionService()
	sessionID := "test-session-123"

	// Test auto-approve functionality
	service.AutoApproveSession(sessionID)

	req := CreatePermissionRequest{
		SessionID:   sessionID,
		ToolName:    "test-tool",
		Description: "Test action",
		Action:      "read",
		Path:        "/test/path",
	}

	granted := service.Request(req)
	assert.True(t, granted, "Auto-approved session should grant permission")
}

func TestPermissionService_Request_PersistentPermission(t *testing.T) {
	// Setup config for working directory
	config.Init(config.Options{})

	service := NewPermissionService()
	sessionID := "test-session-123"

	// First request should require approval
	req := CreatePermissionRequest{
		SessionID:   sessionID,
		ToolName:    "file-tool",
		Description: "Read file",
		Action:      "read",
		Path:        "/test/file.txt",
	}

	// Grant persistent permission
	permissionReq := PermissionRequest{
		ID:          "perm-123",
		SessionID:   sessionID,
		ToolName:    "file-tool",
		Action:      "read",
		Path:        "/test",
	}
	service.GrantPersistant(permissionReq)

	// Same request should now be automatically granted
	granted := service.Request(req)
	assert.True(t, granted, "Persistent permission should grant future requests")
}

func TestPermissionService_Grant_Deny(t *testing.T) {
	service := NewPermissionService()

	tests := []struct {
		name             string
		grantPermission  bool
		expectedResponse bool
	}{
		{
			name:             "grant permission",
			grantPermission:  true,
			expectedResponse: true,
		},
		{
			name:             "deny permission",
			grantPermission:  false,
			expectedResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			config.Init(config.Options{})

			sessionID := "test-session-" + tt.name
			req := CreatePermissionRequest{
				SessionID:   sessionID,
				ToolName:    "test-tool",
				Description: "Test action",
				Action:      "write",
				Path:        "/test/path",
			}

			// Subscribe to permission events BEFORE starting the request
			ctx := context.Background()
			eventsChan := service.Subscribe(ctx)

			// Start the request in a goroutine since it blocks
			resultChan := make(chan bool, 1)
			go func() {
				result := service.Request(req)
				resultChan <- result
			}()

			// Get the permission request event
			select {
			case event := <-eventsChan:
				assert.Equal(t, pubsub.CreatedEvent, event.Type)
				permissionRequest := event.Payload

				// Grant or deny based on test case
				if tt.grantPermission {
					service.Grant(permissionRequest)
				} else {
					service.Deny(permissionRequest)
				}

			case <-time.After(500 * time.Millisecond):
				t.Fatal("Expected to receive a permission request event")
			}

			// Check the result
			select {
			case result := <-resultChan:
				assert.Equal(t, tt.expectedResponse, result)
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Expected to receive a response")
			}
		})
	}
}

func TestPermissionService_Request_PathHandling(t *testing.T) {
	// Setup config
	config.Init(config.Options{})

	service := NewPermissionService()
	sessionID := "test-session-path"

	tests := []struct {
		name         string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "absolute path",
			requestPath:  "/absolute/path/file.txt",
			expectedPath: "/absolute/path",
		},
		{
			name:         "relative path with dot",
			requestPath:  "./file.txt",
			expectedPath: config.WorkingDirectory(),
		},
		{
			name:         "just filename",
			requestPath:  "file.txt",
			expectedPath: "/tmp", // config.Init sets working dir to /tmp
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreatePermissionRequest{
				SessionID:   sessionID,
				ToolName:    "test-tool",
				Description: "Test path handling",
				Action:      "read",
				Path:        tt.requestPath,
			}

			// Start the request in a goroutine
			go func() {
				service.Request(req)
			}()

			// Check the published event
			ctx := context.Background()
			eventsChan := service.Subscribe(ctx)

			select {
			case event := <-eventsChan:
				permissionRequest := event.Payload
				if tt.name == "relative path with dot" {
					// For relative paths, it should use working directory
					assert.NotEmpty(t, permissionRequest.Path)
				} else {
					assert.Equal(t, tt.expectedPath, permissionRequest.Path)
				}

				// Grant to complete the test
				service.Grant(permissionRequest)

			case <-time.After(100 * time.Millisecond):
				t.Fatal("Expected to receive a permission request event")
			}
		})
	}
}

func TestPermissionService_PublishesEvents(t *testing.T) {
	config.Init(config.Options{})

	service := NewPermissionService()
	sessionID := "test-session-events"

	req := CreatePermissionRequest{
		SessionID:   sessionID,
		ToolName:    "event-tool",
		Description: "Test event publishing",
		Action:      "execute",
		Path:        "/test/path",
	}

	// Subscribe to events
	ctx := context.Background()
	eventsChan := service.Subscribe(ctx)

	// Start the request in a goroutine
	go func() {
		service.Request(req)
	}()

	// Verify event is published
	select {
	case event := <-eventsChan:
		assert.Equal(t, pubsub.CreatedEvent, event.Type)
		permissionRequest := event.Payload

		assert.Equal(t, sessionID, permissionRequest.SessionID)
		assert.Equal(t, "event-tool", permissionRequest.ToolName)
		assert.Equal(t, "Test event publishing", permissionRequest.Description)
		assert.Equal(t, "execute", permissionRequest.Action)
		assert.NotEmpty(t, permissionRequest.ID)

		// Grant to complete the test
		service.Grant(permissionRequest)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected to receive a permission request event")
	}
}

func TestPermissionService_ConcurrentRequests(t *testing.T) {
	config.Init(config.Options{})

	service := NewPermissionService()
	sessionID := "test-session-concurrent"

	numRequests := 5
	results := make(chan bool, numRequests)

	// Start multiple concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			req := CreatePermissionRequest{
				SessionID:   sessionID,
				ToolName:    "concurrent-tool",
				Description: "Concurrent test",
				Action:      "read",
				Path:        "/test/concurrent",
			}
			result := service.Request(req)
			results <- result
		}(i)
	}

	// Subscribe to events and grant all requests
	ctx := context.Background()
	eventsChan := service.Subscribe(ctx)

	// Grant all permission requests
	for i := 0; i < numRequests; i++ {
		select {
		case event := <-eventsChan:
			service.Grant(event.Payload)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timed out waiting for request %d", i)
		}
	}

	// Verify all requests were granted
	for i := 0; i < numRequests; i++ {
		select {
		case result := <-results:
			assert.True(t, result, "All concurrent requests should be granted")
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timed out waiting for result %d", i)
		}
	}
}

func TestErrorPermissionDenied(t *testing.T) {
	assert.Equal(t, "permission denied", ErrorPermissionDenied.Error())
}

func TestPermissionRequest_Fields(t *testing.T) {
	req := PermissionRequest{
		ID:          "test-id",
		SessionID:   "test-session",
		ToolName:    "test-tool",
		Description: "test description",
		Action:      "test-action",
		Params:      map[string]string{"key": "value"},
		Path:        "/test/path",
	}

	assert.Equal(t, "test-id", req.ID)
	assert.Equal(t, "test-session", req.SessionID)
	assert.Equal(t, "test-tool", req.ToolName)
	assert.Equal(t, "test description", req.Description)
	assert.Equal(t, "test-action", req.Action)
	assert.Equal(t, "/test/path", req.Path)
	assert.NotNil(t, req.Params)
}