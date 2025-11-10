package storage

import (
	"testing"
)

func TestUserSettingsValidation(t *testing.T) {
	tests := []struct {
		name         string
		reminderTime string
		timezone     string
		wantValid    bool
	}{
		{
			name:         "Valid time and timezone",
			reminderTime: "09:00",
			timezone:     "UTC",
			wantValid:    true,
		},
		{
			name:         "Valid time with different timezone",
			reminderTime: "14:30",
			timezone:     "America/New_York",
			wantValid:    true,
		},
		{
			name:         "Valid time with Europe timezone",
			reminderTime: "08:15",
			timezone:     "Europe/London",
			wantValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &UserSettings{
				ChatID:       12345,
				UserID:       67890,
				ReminderTime: tt.reminderTime,
				Timezone:     tt.timezone,
			}

			if settings.ReminderTime != tt.reminderTime {
				t.Errorf("ReminderTime = %v, want %v", settings.ReminderTime, tt.reminderTime)
			}
			if settings.Timezone != tt.timezone {
				t.Errorf("Timezone = %v, want %v", settings.Timezone, tt.timezone)
			}
		})
	}
}
