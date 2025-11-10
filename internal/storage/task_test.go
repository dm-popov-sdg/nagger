package storage

import (
	"testing"
)

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected string
	}{
		{
			name:     "Active status",
			status:   TaskStatusActive,
			expected: "active",
		},
		{
			name:     "CompletedToday status",
			status:   TaskStatusCompletedToday,
			expected: "completed_today",
		},
		{
			name:     "Closed status",
			status:   TaskStatusClosed,
			expected: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("TaskStatus = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestTaskStatusConstants(t *testing.T) {
	// Verify that status constants are distinct
	statuses := map[TaskStatus]bool{
		TaskStatusActive:         true,
		TaskStatusCompletedToday: true,
		TaskStatusClosed:         true,
	}

	if len(statuses) != 3 {
		t.Errorf("Expected 3 distinct task statuses, got %d", len(statuses))
	}
}
