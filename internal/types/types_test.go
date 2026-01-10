package types

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define static errors for tests
var (
	errNetworkTimeout       = errors.New("network timeout")
	errCustom               = errors.New("custom error")
	errDeliveryFailed       = errors.New("delivery failed")
	errInvalidConfiguration = errors.New("invalid configuration")
)

// Test ChannelType constants and string values
func TestChannelTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		channel  ChannelType
		expected string
	}{
		{"Slack channel", ChannelSlack, "slack"},
		{"Email channel", ChannelEmail, "email"},
		{"Webhook channel", ChannelWebhook, "webhook"},
		{"Teams channel", ChannelTeams, "teams"},
		{"Discord channel", ChannelDiscord, "discord"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.channel))
		})
	}
}

func TestChannelTypeJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		channel  ChannelType
		expected string
	}{
		{"Slack JSON", ChannelSlack, `"slack"`},
		{"Email JSON", ChannelEmail, `"email"`},
		{"Webhook JSON", ChannelWebhook, `"webhook"`},
		{"Teams JSON", ChannelTeams, `"teams"`},
		{"Discord JSON", ChannelDiscord, `"discord"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.channel)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled ChannelType
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.channel, unmarshaled)
		})
	}
}

// Test SeverityLevel constants and string values
func TestSeverityLevelConstants(t *testing.T) {
	tests := []struct {
		name     string
		severity SeverityLevel
		expected string
	}{
		{"Info severity", SeverityInfo, "info"},
		{"Warning severity", SeverityWarning, "warning"},
		{"Critical severity", SeverityCritical, "critical"},
		{"Emergency severity", SeverityEmergency, "emergency"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.severity))
		})
	}
}

func TestSeverityLevelJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		severity SeverityLevel
		expected string
	}{
		{"Info JSON", SeverityInfo, `"info"`},
		{"Warning JSON", SeverityWarning, `"warning"`},
		{"Critical JSON", SeverityCritical, `"critical"`},
		{"Emergency JSON", SeverityEmergency, `"emergency"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.severity)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled SeverityLevel
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.severity, unmarshaled)
		})
	}
}

// Test Priority constants and string values
func TestPriorityConstants(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{"Low priority", PriorityLow, "low"},
		{"Normal priority", PriorityNormal, "normal"},
		{"High priority", PriorityHigh, "high"},
		{"Urgent priority", PriorityUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.priority))
		})
	}
}

func TestPriorityJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{"Low JSON", PriorityLow, `"low"`},
		{"Normal JSON", PriorityNormal, `"normal"`},
		{"High JSON", PriorityHigh, `"high"`},
		{"Urgent JSON", PriorityUrgent, `"urgent"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.priority)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled Priority
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.priority, unmarshaled)
		})
	}
}

// Test Notification struct
func TestNotificationJSONSerialization(t *testing.T) {
	timestamp := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name         string
		notification Notification
	}{
		{
			name: "complete notification",
			notification: Notification{
				ID:         "notif-123",
				Subject:    "Coverage Alert",
				Message:    "Coverage dropped below threshold",
				Severity:   SeverityCritical,
				Priority:   PriorityHigh,
				Timestamp:  timestamp,
				Author:     "john.doe@example.com",
				Repository: "owner/repo",
				Branch:     "feature/test",
				PRNumber:   42,
				CommitSHA:  "abc123def456",
				CoverageData: &CoverageData{
					Current:  75.5,
					Previous: 80.0,
					Change:   -4.5,
					Target:   90.0,
				},
				TrendData: &TrendData{
					Direction:  "declining",
					Confidence: 0.85,
				},
				RichContent: &RichContent{
					Markdown: "## Coverage Alert\n\nCoverage has **dropped**.",
					HTML:     "<h2>Coverage Alert</h2><p>Coverage has <strong>dropped</strong>.</p>",
					Fields:   map[string]string{"project": "go-coverage", "environment": "production"},
				},
				Metadata: map[string]interface{}{
					"build_id":     float64(12345), // JSON unmarshaling converts int to float64
					"build_number": "v1.2.3",
					"duration":     300.5,
					"tags":         []interface{}{"ci", "coverage", "test"}, // JSON unmarshaling converts []string to []interface{}
				},
			},
		},
		{
			name: "minimal notification",
			notification: Notification{
				ID:        "min-123",
				Subject:   "Simple Alert",
				Message:   "Test message",
				Severity:  SeverityInfo,
				Priority:  PriorityNormal,
				Timestamp: timestamp,
			},
		},
		{
			name: "notification with empty nested structs",
			notification: Notification{
				ID:           "empty-123",
				Subject:      "Empty Nested",
				Message:      "Has empty nested structs",
				Severity:     SeverityWarning,
				Priority:     PriorityLow,
				Timestamp:    timestamp,
				CoverageData: &CoverageData{},
				TrendData:    &TrendData{},
				RichContent:  &RichContent{},
				Metadata:     nil, // Empty map becomes nil after JSON unmarshaling
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.notification)
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Test deserialization
			var unmarshaled Notification
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.notification, unmarshaled)
		})
	}
}

// Test CoverageData struct
func TestCoverageDataJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		data     CoverageData
		expected string
	}{
		{
			name: "complete coverage data",
			data: CoverageData{
				Current:  85.7,
				Previous: 82.3,
				Change:   3.4,
				Target:   90.0,
			},
			expected: `{"current":85.7,"previous":82.3,"change":3.4,"target":90}`,
		},
		{
			name: "coverage data without target",
			data: CoverageData{
				Current:  75.0,
				Previous: 78.5,
				Change:   -3.5,
			},
			expected: `{"current":75,"previous":78.5,"change":-3.5}`,
		},
		{
			name: "zero coverage data",
			data: CoverageData{
				Current:  0.0,
				Previous: 0.0,
				Change:   0.0,
				Target:   0.0,
			},
			expected: `{"current":0,"previous":0,"change":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.data)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled CoverageData
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.data, unmarshaled)
		})
	}
}

// Test TrendData struct
func TestTrendDataJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		data     TrendData
		expected string
	}{
		{
			name: "declining trend",
			data: TrendData{
				Direction:  "declining",
				Confidence: 0.95,
			},
			expected: `{"direction":"declining","confidence":0.95}`,
		},
		{
			name: "improving trend",
			data: TrendData{
				Direction:  "improving",
				Confidence: 0.87,
			},
			expected: `{"direction":"improving","confidence":0.87}`,
		},
		{
			name: "stable trend",
			data: TrendData{
				Direction:  "stable",
				Confidence: 1.0,
			},
			expected: `{"direction":"stable","confidence":1}`,
		},
		{
			name: "zero confidence",
			data: TrendData{
				Direction:  "unknown",
				Confidence: 0.0,
			},
			expected: `{"direction":"unknown","confidence":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.data)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled TrendData
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.data, unmarshaled)
		})
	}
}

// Test RichContent struct
func TestRichContentJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		content  RichContent
		expected string
	}{
		{
			name: "complete rich content",
			content: RichContent{
				Markdown: "# Header\n\n**Bold text**",
				HTML:     "<h1>Header</h1><p><strong>Bold text</strong></p>",
				Fields:   map[string]string{"author": "John Doe", "version": "1.0.0"},
			},
			expected: `{"markdown":"# Header\n\n**Bold text**","html":"<h1>Header</h1><p><strong>Bold text</strong></p>","fields":{"author":"John Doe","version":"1.0.0"}}`,
		},
		{
			name: "markdown only",
			content: RichContent{
				Markdown: "Simple *markdown* text",
			},
			expected: `{"markdown":"Simple *markdown* text"}`,
		},
		{
			name: "HTML only",
			content: RichContent{
				HTML: "<p>Simple <em>HTML</em> text</p>",
			},
			expected: `{"html":"<p>Simple <em>HTML</em> text</p>"}`,
		},
		{
			name: "fields only",
			content: RichContent{
				Fields: map[string]string{"status": "success", "duration": "2m30s"},
			},
			expected: `{"fields":{"duration":"2m30s","status":"success"}}`,
		},
		{
			name:     "empty rich content",
			content:  RichContent{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.content)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled RichContent
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.content, unmarshaled)
		})
	}
}

// Test DeliveryResult struct
func TestDeliveryResultJSONSerialization(t *testing.T) {
	timestamp := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	deliveryTime := 2 * time.Second

	tests := []struct {
		name   string
		result DeliveryResult
	}{
		{
			name: "successful delivery",
			result: DeliveryResult{
				Channel:      ChannelSlack,
				Success:      true,
				MessageID:    "msg-12345",
				Timestamp:    timestamp,
				DeliveryTime: deliveryTime,
				Error:        nil,
			},
		},
		{
			name: "failed delivery without error",
			result: DeliveryResult{
				Channel:      ChannelEmail,
				Success:      false,
				MessageID:    "",
				Timestamp:    timestamp,
				DeliveryTime: deliveryTime,
				Error:        nil, // Don't test with actual error due to JSON marshaling complexity
			},
		},
		{
			name: "partial delivery",
			result: DeliveryResult{
				Channel:      ChannelWebhook,
				Success:      true,
				MessageID:    "webhook-response-id",
				Timestamp:    timestamp,
				DeliveryTime: 500 * time.Millisecond,
				Error:        nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.result)
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Test deserialization - note that errors don't marshal/unmarshal perfectly
			var unmarshaled DeliveryResult
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			// Test deserialization
			assert.Equal(t, tt.result, unmarshaled)
		})
	}
}

// Test DeliveryResult with error separately since errors don't marshal/unmarshal well
func TestDeliveryResultWithError(t *testing.T) {
	testError := errDeliveryFailed
	timestamp := time.Now()

	result := DeliveryResult{
		Channel:      ChannelEmail,
		Success:      false,
		MessageID:    "",
		Timestamp:    timestamp,
		DeliveryTime: 2 * time.Second,
		Error:        testError,
	}

	// Test that we can create and access the error
	assert.Equal(t, ChannelEmail, result.Channel)
	assert.False(t, result.Success)
	assert.Empty(t, result.MessageID)
	assert.Equal(t, timestamp, result.Timestamp)
	assert.Equal(t, 2*time.Second, result.DeliveryTime)
	require.Error(t, result.Error)
	assert.Equal(t, "delivery failed", result.Error.Error())
}

// Test RateLimit struct
func TestRateLimitJSONSerialization(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit RateLimit
		expected  string
	}{
		{
			name: "complete rate limit",
			rateLimit: RateLimit{
				RequestsPerMinute: 60,
				RequestsPerHour:   3600,
				RequestsPerDay:    86400,
				BurstSize:         10,
				Window:            time.Minute,
			},
			expected: `{"requests_per_minute":60,"requests_per_hour":3600,"requests_per_day":86400,"burst_size":10,"window":60000000000}`,
		},
		{
			name: "minimal rate limit",
			rateLimit: RateLimit{
				RequestsPerMinute: 1,
				Window:            time.Second,
			},
			expected: `{"requests_per_minute":1,"requests_per_hour":0,"requests_per_day":0,"burst_size":0,"window":1000000000}`,
		},
		{
			name:      "zero rate limit",
			rateLimit: RateLimit{},
			expected:  `{"requests_per_minute":0,"requests_per_hour":0,"requests_per_day":0,"burst_size":0,"window":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.rateLimit)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled RateLimit
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.rateLimit, unmarshaled)
		})
	}
}

// Test zero values for all structs
func TestStructZeroValues(t *testing.T) {
	t.Run("Notification zero values", func(t *testing.T) {
		var n Notification
		assert.Empty(t, n.ID)
		assert.Empty(t, n.Subject)
		assert.Empty(t, n.Message)
		assert.Equal(t, SeverityLevel(""), n.Severity)
		assert.Equal(t, Priority(""), n.Priority)
		assert.True(t, n.Timestamp.IsZero())
		assert.Empty(t, n.Author)
		assert.Empty(t, n.Repository)
		assert.Empty(t, n.Branch)
		assert.Equal(t, 0, n.PRNumber)
		assert.Empty(t, n.CommitSHA)
		assert.Nil(t, n.CoverageData)
		assert.Nil(t, n.TrendData)
		assert.Nil(t, n.RichContent)
		assert.Nil(t, n.Metadata)
	})

	t.Run("CoverageData zero values", func(t *testing.T) {
		var c CoverageData
		assert.Zero(t, c.Current)
		assert.Zero(t, c.Previous)
		assert.Zero(t, c.Change)
		assert.Zero(t, c.Target)
	})

	t.Run("TrendData zero values", func(t *testing.T) {
		var trend TrendData
		assert.Empty(t, trend.Direction)
		assert.Zero(t, trend.Confidence)
	})

	t.Run("RichContent zero values", func(t *testing.T) {
		var r RichContent
		assert.Empty(t, r.Markdown)
		assert.Empty(t, r.HTML)
		assert.Nil(t, r.Fields)
	})

	t.Run("DeliveryResult zero values", func(t *testing.T) {
		var d DeliveryResult
		assert.Equal(t, ChannelType(""), d.Channel)
		assert.False(t, d.Success)
		assert.Empty(t, d.MessageID)
		assert.True(t, d.Timestamp.IsZero())
		assert.Equal(t, time.Duration(0), d.DeliveryTime)
		assert.NoError(t, d.Error)
	})

	t.Run("RateLimit zero values", func(t *testing.T) {
		var r RateLimit
		assert.Equal(t, 0, r.RequestsPerMinute)
		assert.Equal(t, 0, r.RequestsPerHour)
		assert.Equal(t, 0, r.RequestsPerDay)
		assert.Equal(t, 0, r.BurstSize)
		assert.Equal(t, time.Duration(0), r.Window)
	})
}

// Test edge cases with special characters and values
func TestSpecialCharactersAndEdgeCases(t *testing.T) {
	t.Run("Notification with unicode and special characters", func(t *testing.T) {
		notification := Notification{
			ID:         "unicode-test-123",
			Subject:    "Coverage Alert 🚨 Test",
			Message:    "Message with special chars: !@#$%^&*()_+-={}[]|\\:;\";'<>?,./ and emojis: 🎉",
			Severity:   SeverityEmergency,
			Priority:   PriorityUrgent,
			Timestamp:  time.Now(),
			Author:     "user@test.com",
			Repository: "owner/repo-with-special-chars",
			Branch:     "feature/test-branch",
			CommitSHA:  "abc123def456",
			Metadata: map[string]interface{}{
				"unicode_key":   "unicode_value",
				"special_chars": "!@#$%^&*()_+-={}[]|\\:;\";'<>?,./ ",
				"emoji":         "🚀🎉🚨",
			},
		}

		// Test JSON serialization/deserialization preserves special characters
		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		var unmarshaled Notification
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, notification.ID, unmarshaled.ID)
		assert.Equal(t, notification.Subject, unmarshaled.Subject)
		assert.Equal(t, notification.Message, unmarshaled.Message)
		assert.Equal(t, notification.Author, unmarshaled.Author)
		assert.Equal(t, notification.Repository, unmarshaled.Repository)
		assert.Equal(t, notification.Branch, unmarshaled.Branch)
		assert.Equal(t, notification.CommitSHA, unmarshaled.CommitSHA)
		assert.Equal(t, notification.Metadata, unmarshaled.Metadata)
	})

	t.Run("CoverageData with extreme values", func(t *testing.T) {
		coverage := CoverageData{
			Current:  100.0,
			Previous: 0.0,
			Change:   100.0,
			Target:   99.99,
		}

		jsonData, err := json.Marshal(coverage)
		require.NoError(t, err)

		var unmarshaled CoverageData
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, coverage, unmarshaled)
	})

	t.Run("Negative coverage changes", func(t *testing.T) {
		coverage := CoverageData{
			Current:  45.0,
			Previous: 95.0,
			Change:   -50.0,
			Target:   80.0,
		}

		jsonData, err := json.Marshal(coverage)
		require.NoError(t, err)

		var unmarshaled CoverageData
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, coverage, unmarshaled)
	})

	t.Run("Very long delivery time", func(t *testing.T) {
		result := DeliveryResult{
			Channel:      ChannelTeams,
			Success:      true,
			MessageID:    "very-long-delivery-id",
			Timestamp:    time.Now(),
			DeliveryTime: 24 * time.Hour, // Very long delivery time
			Error:        nil,
		}

		jsonData, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled DeliveryResult
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, result.Channel, unmarshaled.Channel)
		assert.Equal(t, result.Success, unmarshaled.Success)
		assert.Equal(t, result.MessageID, unmarshaled.MessageID)
		assert.Equal(t, result.DeliveryTime, unmarshaled.DeliveryTime)
	})
}

// Test custom type constants validity
func TestEnumValidity(t *testing.T) {
	t.Run("all ChannelType constants are non-empty", func(t *testing.T) {
		channels := []ChannelType{
			ChannelSlack,
			ChannelEmail,
			ChannelWebhook,
			ChannelTeams,
			ChannelDiscord,
		}

		for _, channel := range channels {
			assert.NotEmpty(t, string(channel), "ChannelType should not be empty: %s", channel)
			assert.NotContains(t, string(channel), " ", "ChannelType should not contain spaces: %s", channel)
		}
	})

	t.Run("all SeverityLevel constants are non-empty", func(t *testing.T) {
		severities := []SeverityLevel{
			SeverityInfo,
			SeverityWarning,
			SeverityCritical,
			SeverityEmergency,
		}

		for _, severity := range severities {
			assert.NotEmpty(t, string(severity), "SeverityLevel should not be empty: %s", severity)
			assert.NotContains(t, string(severity), " ", "SeverityLevel should not contain spaces: %s", severity)
		}
	})

	t.Run("all Priority constants are non-empty", func(t *testing.T) {
		priorities := []Priority{
			PriorityLow,
			PriorityNormal,
			PriorityHigh,
			PriorityUrgent,
		}

		for _, priority := range priorities {
			assert.NotEmpty(t, string(priority), "Priority should not be empty: %s", priority)
			assert.NotContains(t, string(priority), " ", "Priority should not contain spaces: %s", priority)
		}
	})

	t.Run("enum constants are unique", func(t *testing.T) {
		// Test ChannelType uniqueness
		channels := []ChannelType{ChannelSlack, ChannelEmail, ChannelWebhook, ChannelTeams, ChannelDiscord}
		channelSet := make(map[ChannelType]bool)
		for _, ch := range channels {
			assert.False(t, channelSet[ch], "ChannelType should be unique: %s", ch)
			channelSet[ch] = true
		}

		// Test SeverityLevel uniqueness
		severities := []SeverityLevel{SeverityInfo, SeverityWarning, SeverityCritical, SeverityEmergency}
		severitySet := make(map[SeverityLevel]bool)
		for _, sev := range severities {
			assert.False(t, severitySet[sev], "SeverityLevel should be unique: %s", sev)
			severitySet[sev] = true
		}

		// Test Priority uniqueness
		priorities := []Priority{PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent}
		prioritySet := make(map[Priority]bool)
		for _, pri := range priorities {
			assert.False(t, prioritySet[pri], "Priority should be unique: %s", pri)
			prioritySet[pri] = true
		}
	})
}

// Mock implementation of NotificationChannel interface for testing
type MockNotificationChannel struct {
	channelType        ChannelType
	validationError    error
	richContentSupport bool
	rateLimit          *RateLimit
	sendError          error
	sendResult         *DeliveryResult
}

func (m *MockNotificationChannel) Send(ctx context.Context, notification *Notification) (*DeliveryResult, error) {
	if m.sendError != nil {
		return nil, m.sendError
	}
	if m.sendResult != nil {
		return m.sendResult, nil
	}
	return &DeliveryResult{
		Channel:      m.channelType,
		Success:      true,
		MessageID:    "mock-message-id",
		Timestamp:    time.Now(),
		DeliveryTime: time.Millisecond * 100,
		Error:        nil,
	}, nil
}

func (m *MockNotificationChannel) ValidateConfig() error {
	return m.validationError
}

func (m *MockNotificationChannel) GetChannelType() ChannelType {
	return m.channelType
}

func (m *MockNotificationChannel) SupportsRichContent() bool {
	return m.richContentSupport
}

func (m *MockNotificationChannel) GetRateLimit() *RateLimit {
	return m.rateLimit
}

// Test NotificationChannel interface
func TestNotificationChannelInterface(t *testing.T) {
	t.Run("mock channel implements interface correctly", func(t *testing.T) {
		mock := &MockNotificationChannel{
			channelType:        ChannelSlack,
			validationError:    nil,
			richContentSupport: true,
			rateLimit: &RateLimit{
				RequestsPerMinute: 60,
				BurstSize:         10,
				Window:            time.Minute,
			},
		}

		// Test interface methods
		assert.Equal(t, ChannelSlack, mock.GetChannelType())
		assert.True(t, mock.SupportsRichContent())
		require.NoError(t, mock.ValidateConfig())
		assert.NotNil(t, mock.GetRateLimit())

		// Test Send method
		ctx := context.Background()
		notification := &Notification{
			ID:        "test-123",
			Subject:   "Test",
			Message:   "Test message",
			Severity:  SeverityInfo,
			Priority:  PriorityNormal,
			Timestamp: time.Now(),
		}

		result, err := mock.Send(ctx, notification)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, ChannelSlack, result.Channel)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.MessageID)
		assert.False(t, result.Timestamp.IsZero())
		assert.Greater(t, result.DeliveryTime, time.Duration(0))
	})

	t.Run("mock channel with validation error", func(t *testing.T) {
		mock := &MockNotificationChannel{
			channelType:     ChannelEmail,
			validationError: errInvalidConfiguration,
		}

		require.Error(t, mock.ValidateConfig())
		assert.Equal(t, "invalid configuration", mock.ValidateConfig().Error())
	})

	t.Run("mock channel with send error", func(t *testing.T) {
		mock := &MockNotificationChannel{
			channelType: ChannelWebhook,
			sendError:   errNetworkTimeout,
		}

		ctx := context.Background()
		notification := &Notification{ID: "test"}

		result, err := mock.Send(ctx, notification)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "network timeout", err.Error())
	})

	t.Run("mock channel with custom result", func(t *testing.T) {
		customResult := &DeliveryResult{
			Channel:      ChannelDiscord,
			Success:      false,
			MessageID:    "failed-msg",
			Timestamp:    time.Now(),
			DeliveryTime: 5 * time.Second,
			Error:        errCustom,
		}

		mock := &MockNotificationChannel{
			channelType: ChannelDiscord,
			sendResult:  customResult,
		}

		ctx := context.Background()
		notification := &Notification{ID: "test"}

		result, err := mock.Send(ctx, notification)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, customResult, result)
	})

	t.Run("interface can be used polymorphically", func(t *testing.T) {
		channels := make([]NotificationChannel, 0, 3)
		channels = append(channels, &MockNotificationChannel{channelType: ChannelSlack})
		channels = append(channels, &MockNotificationChannel{channelType: ChannelEmail})
		channels = append(channels, &MockNotificationChannel{channelType: ChannelWebhook})

		for i, channel := range channels {
			// Test that interface methods work
			channelType := channel.GetChannelType()
			assert.NotEmpty(t, string(channelType))

			// Test ValidateConfig method
			err := channel.ValidateConfig()
			require.NoError(t, err, "Channel %d should validate successfully", i)

			// Test other interface methods
			supportsRich := channel.SupportsRichContent()
			assert.IsType(t, false, supportsRich) // Just ensure it returns a bool

			rateLimit := channel.GetRateLimit()
			// Rate limit can be nil, so just check it doesn't panic
			if rateLimit != nil {
				assert.IsType(t, &RateLimit{}, rateLimit)
			}
		}
	})
}
