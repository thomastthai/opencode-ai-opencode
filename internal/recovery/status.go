package recovery

import (
	"context"
	"sync"
	"time"

	"github.com/opencode-ai/opencode/internal/pubsub"
)

// RecoveryState represents the current state of post-sleep recovery
type RecoveryState int

const (
	RecoveryIdle RecoveryState = iota
	RecoveryInProgress
	RecoveryCompleted
	RecoveryFailed
)

func (s RecoveryState) String() string {
	switch s {
	case RecoveryIdle:
		return "idle"
	case RecoveryInProgress:
		return "in_progress"
	case RecoveryCompleted:
		return "completed"
	case RecoveryFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// RecoveryStep represents a single step in the recovery process
type RecoveryStep struct {
	Name        string
	Description string
	Status      StepStatus
	Error       error
	StartTime   time.Time
	EndTime     time.Time
}

type StepStatus int

const (
	StepPending StepStatus = iota
	StepInProgress
	StepCompleted
	StepFailed
)

func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepInProgress:
		return "in_progress"
	case StepCompleted:
		return "completed"
	case StepFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// RecoveryStatus represents the current recovery progress
type RecoveryStatus struct {
	State       RecoveryState
	Steps       []RecoveryStep
	StartTime   time.Time
	EndTime     time.Time
	TotalSteps  int
	CompletedSteps int
}

// Service manages recovery status and progress
type Service interface {
	pubsub.Suscriber[RecoveryStatus]
	StartRecovery(ctx context.Context, steps []string) error
	UpdateStep(stepName string, status StepStatus, err error)
	CompleteRecovery()
	FailRecovery(err error)
	GetStatus() RecoveryStatus
	IsRecovering() bool
}

type service struct {
	*pubsub.Broker[RecoveryStatus]
	mu     sync.RWMutex
	status RecoveryStatus
}

// NewService creates a new recovery service
func NewService() Service {
	return &service{
		Broker: pubsub.NewBroker[RecoveryStatus](),
		status: RecoveryStatus{
			State: RecoveryIdle,
			Steps: []RecoveryStep{},
		},
	}
}

func (s *service) StartRecovery(ctx context.Context, stepNames []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	steps := make([]RecoveryStep, len(stepNames))
	for i, name := range stepNames {
		steps[i] = RecoveryStep{
			Name:        name,
			Description: getStepDescription(name),
			Status:      StepPending,
			StartTime:   time.Time{},
			EndTime:     time.Time{},
		}
	}

	s.status = RecoveryStatus{
		State:          RecoveryInProgress,
		Steps:          steps,
		StartTime:      time.Now(),
		TotalSteps:     len(steps),
		CompletedSteps: 0,
	}

	s.Publish(pubsub.UpdatedEvent, s.status)
	return nil
}

func (s *service) UpdateStep(stepName string, status StepStatus, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.status.Steps {
		if s.status.Steps[i].Name == stepName {
			step := &s.status.Steps[i]
			
			if status == StepInProgress && step.Status == StepPending {
				step.StartTime = time.Now()
			}
			
			step.Status = status
			step.Error = err
			
			if status == StepCompleted || status == StepFailed {
				step.EndTime = time.Now()
				if status == StepCompleted {
					s.status.CompletedSteps++
				}
			}
			
			break
		}
	}

	s.Publish(pubsub.UpdatedEvent, s.status)
}

func (s *service) CompleteRecovery() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.State = RecoveryCompleted
	s.status.EndTime = time.Now()
	s.Publish(pubsub.UpdatedEvent, s.status)
}

func (s *service) FailRecovery(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.State = RecoveryFailed
	s.status.EndTime = time.Now()
	s.Publish(pubsub.UpdatedEvent, s.status)
}

func (s *service) GetStatus() RecoveryStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *service) IsRecovering() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.State == RecoveryInProgress
}

// getStepDescription returns a user-friendly description for recovery steps
func getStepDescription(stepName string) string {
	descriptions := map[string]string{
		"sessions":    "Checking session database connectivity...",
		"messages":    "Validating message storage...",
		"history":     "Verifying file history access...",
		"shell":       "Restarting shell processes...",
		"lsp":         "Reconnecting language servers...",
		"permissions": "Updating permissions...",
		"agent":       "Initializing AI agent...",
	}
	
	if desc, exists := descriptions[stepName]; exists {
		return desc
	}
	return "Processing " + stepName + "..."
}