package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlackConfigJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   SlackConfig
		expected string
	}{
		{
			name: "complete config",
			config: SlackConfig{
				WebhookURL:   "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				Channel:      "#general",
				Username:     "bot",
				IconEmoji:    ":robot_face:",
				IconURL:      "https://example.com/icon.png",
				Enabled:      true,
				Timeout:      30,
				CustomFields: map[string]string{"env": "prod", "service": "coverage"},
			},
			expected: `{"webhook_url":"https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX","channel":"#general","username":"bot","icon_emoji":":robot_face:","icon_url":"https://example.com/icon.png","enabled":true,"timeout":30,"custom_fields":{"env":"prod","service":"coverage"}}`,
		},
		{
			name: "minimal config",
			config: SlackConfig{
				WebhookURL: "https://hooks.slack.com/services/minimal",
				Enabled:    false,
				Timeout:    0,
			},
			expected: `{"webhook_url":"https://hooks.slack.com/services/minimal","enabled":false,"timeout":0}`,
		},
		{
			name:     "empty config",
			config:   SlackConfig{},
			expected: `{"webhook_url":"","enabled":false,"timeout":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled SlackConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

func TestDiscordConfigJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   DiscordConfig
		expected string
	}{
		{
			name: "complete config",
			config: DiscordConfig{
				WebhookURL: "https://discord.com/api/webhooks/123456789/abcdefgh",
				Username:   "CoverageBot",
				AvatarURL:  "https://example.com/avatar.png",
				EmbedColor: 0x00ff00,
				Enabled:    true,
				Timeout:    15,
			},
			expected: `{"webhook_url":"https://discord.com/api/webhooks/123456789/abcdefgh","username":"CoverageBot","avatar_url":"https://example.com/avatar.png","embed_color":65280,"enabled":true,"timeout":15}`,
		},
		{
			name: "minimal config",
			config: DiscordConfig{
				WebhookURL: "https://discord.com/api/webhooks/minimal",
				Enabled:    false,
			},
			expected: `{"webhook_url":"https://discord.com/api/webhooks/minimal","enabled":false,"timeout":0}`,
		},
		{
			name: "zero embed color",
			config: DiscordConfig{
				WebhookURL: "https://discord.com/api/webhooks/test",
				EmbedColor: 0,
				Enabled:    true,
			},
			expected: `{"webhook_url":"https://discord.com/api/webhooks/test","enabled":true,"timeout":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled DiscordConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

func TestEmailConfigJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   EmailConfig
		expected string
	}{
		{
			name: "complete config",
			config: EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    587,
				Username:    "user@example.com",
				Password:    "secretpassword",
				FromEmail:   "noreply@example.com",
				FromName:    "Coverage Bot",
				ToEmails:    []string{"dev@example.com", "qa@example.com"},
				CCEmails:    []string{"manager@example.com"},
				BCCEmails:   []string{"audit@example.com"},
				UseTLS:      true,
				UseStartTLS: false,
				Enabled:     true,
				Timeout:     30,
			},
			expected: `{"smtp_host":"smtp.gmail.com","smtp_port":587,"username":"user@example.com","password":"secretpassword","from_email":"noreply@example.com","from_name":"Coverage Bot","to_emails":["dev@example.com","qa@example.com"],"cc_emails":["manager@example.com"],"bcc_emails":["audit@example.com"],"use_tls":true,"use_starttls":false,"enabled":true,"timeout":30}`,
		},
		{
			name: "minimal config",
			config: EmailConfig{
				SMTPHost:  "localhost",
				SMTPPort:  25,
				FromEmail: "test@localhost",
				ToEmails:  []string{"test@localhost"},
				Enabled:   false,
			},
			expected: `{"smtp_host":"localhost","smtp_port":25,"username":"","password":"","from_email":"test@localhost","to_emails":["test@localhost"],"use_tls":false,"use_starttls":false,"enabled":false,"timeout":0}`,
		},
		{
			name: "nil arrays",
			config: EmailConfig{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  465,
				FromEmail: "from@example.com",
				ToEmails:  nil,
				CCEmails:  nil,
				BCCEmails: nil,
				UseTLS:    true,
				Enabled:   true,
			},
			expected: `{"smtp_host":"smtp.example.com","smtp_port":465,"username":"","password":"","from_email":"from@example.com","to_emails":null,"use_tls":true,"use_starttls":false,"enabled":true,"timeout":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled EmailConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

func TestWebhookConfigJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   WebhookConfig
		expected string
	}{
		{
			name: "complete config with bearer auth",
			config: WebhookConfig{
				URL:         "https://api.example.com/webhook",
				Method:      "POST",
				Headers:     map[string]string{"X-Custom": "value", "User-Agent": "CoverageBot/1.0"},
				ContentType: "application/json",
				AuthType:    "bearer",
				AuthToken:   "eyJhbGciOiJIUzI1NiJ9...",
				Enabled:     true,
				Timeout:     60,
			},
			expected: `{"url":"https://api.example.com/webhook","method":"POST","headers":{"User-Agent":"CoverageBot/1.0","X-Custom":"value"},"content_type":"application/json","auth_type":"bearer","auth_token":"eyJhbGciOiJIUzI1NiJ9...","enabled":true,"timeout":60}`,
		},
		{
			name: "basic auth config",
			config: WebhookConfig{
				URL:          "https://webhook.example.com",
				Method:       "PUT",
				ContentType:  "application/xml",
				AuthType:     "basic",
				AuthUsername: "admin",
				AuthPassword: "password123",
				Enabled:      true,
				Timeout:      45,
			},
			expected: `{"url":"https://webhook.example.com","method":"PUT","content_type":"application/xml","auth_type":"basic","auth_username":"admin","auth_password":"password123","enabled":true,"timeout":45}`,
		},
		{
			name: "minimal config",
			config: WebhookConfig{
				URL:         "https://simple.webhook.com",
				Method:      "GET",
				ContentType: "text/plain",
				Enabled:     false,
			},
			expected: `{"url":"https://simple.webhook.com","method":"GET","content_type":"text/plain","enabled":false,"timeout":0}`,
		},
		{
			name: "nil headers map",
			config: WebhookConfig{
				URL:         "https://no-headers.example.com",
				Method:      "POST",
				Headers:     nil,
				ContentType: "application/json",
				Enabled:     true,
			},
			expected: `{"url":"https://no-headers.example.com","method":"POST","content_type":"application/json","enabled":true,"timeout":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled WebhookConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

func TestTeamsConfigJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   TeamsConfig
		expected string
	}{
		{
			name: "complete config",
			config: TeamsConfig{
				WebhookURL:    "https://outlook.office.com/webhook/abc123",
				ThemeColor:    "0078D4",
				ActivityTitle: "Coverage Report",
				ActivityImage: "https://example.com/coverage-icon.png",
				Enabled:       true,
				Timeout:       20,
			},
			expected: `{"webhook_url":"https://outlook.office.com/webhook/abc123","theme_color":"0078D4","activity_title":"Coverage Report","activity_image":"https://example.com/coverage-icon.png","enabled":true,"timeout":20}`,
		},
		{
			name: "minimal config",
			config: TeamsConfig{
				WebhookURL: "https://teams.webhook.url",
				Enabled:    false,
			},
			expected: `{"webhook_url":"https://teams.webhook.url","enabled":false,"timeout":0}`,
		},
		{
			name: "with hex color",
			config: TeamsConfig{
				WebhookURL: "https://teams.example.com",
				ThemeColor: "#FF5722",
				Enabled:    true,
				Timeout:    10,
			},
			expected: `{"webhook_url":"https://teams.example.com","theme_color":"#FF5722","enabled":true,"timeout":10}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			var unmarshaled TeamsConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

// Test zero value behaviors
func TestConfigZeroValues(t *testing.T) {
	t.Run("SlackConfig zero values", func(t *testing.T) {
		var config SlackConfig
		assert.Empty(t, config.WebhookURL)
		assert.Empty(t, config.Channel)
		assert.Empty(t, config.Username)
		assert.Empty(t, config.IconEmoji)
		assert.Empty(t, config.IconURL)
		assert.False(t, config.Enabled)
		assert.Equal(t, 0, config.Timeout)
		assert.Nil(t, config.CustomFields)
	})

	t.Run("DiscordConfig zero values", func(t *testing.T) {
		var config DiscordConfig
		assert.Empty(t, config.WebhookURL)
		assert.Empty(t, config.Username)
		assert.Empty(t, config.AvatarURL)
		assert.Equal(t, 0, config.EmbedColor)
		assert.False(t, config.Enabled)
		assert.Equal(t, 0, config.Timeout)
	})

	t.Run("EmailConfig zero values", func(t *testing.T) {
		var config EmailConfig
		assert.Empty(t, config.SMTPHost)
		assert.Equal(t, 0, config.SMTPPort)
		assert.Empty(t, config.Username)
		assert.Empty(t, config.Password)
		assert.Empty(t, config.FromEmail)
		assert.Empty(t, config.FromName)
		assert.Nil(t, config.ToEmails)
		assert.Nil(t, config.CCEmails)
		assert.Nil(t, config.BCCEmails)
		assert.False(t, config.UseTLS)
		assert.False(t, config.UseStartTLS)
		assert.False(t, config.Enabled)
		assert.Equal(t, 0, config.Timeout)
	})

	t.Run("WebhookConfig zero values", func(t *testing.T) {
		var config WebhookConfig
		assert.Empty(t, config.URL)
		assert.Empty(t, config.Method)
		assert.Nil(t, config.Headers)
		assert.Empty(t, config.ContentType)
		assert.Empty(t, config.AuthType)
		assert.Empty(t, config.AuthToken)
		assert.Empty(t, config.AuthUsername)
		assert.Empty(t, config.AuthPassword)
		assert.False(t, config.Enabled)
		assert.Equal(t, 0, config.Timeout)
	})

	t.Run("TeamsConfig zero values", func(t *testing.T) {
		var config TeamsConfig
		assert.Empty(t, config.WebhookURL)
		assert.Empty(t, config.ThemeColor)
		assert.Empty(t, config.ActivityTitle)
		assert.Empty(t, config.ActivityImage)
		assert.False(t, config.Enabled)
		assert.Equal(t, 0, config.Timeout)
	})
}

// Test edge cases with special characters and encoding
func TestConfigSpecialCharacters(t *testing.T) {
	t.Run("SlackConfig with special characters", func(t *testing.T) {
		config := SlackConfig{
			WebhookURL:   "https://hooks.slack.com/services/TEST/with%20spaces",
			Channel:      "#test-channel_with-special.chars",
			Username:     "bot name with spaces",
			IconEmoji:    ":custom_emoji:",
			CustomFields: map[string]string{"unicode": "test", "special": "!@#$%^&*()_+-={}|[]\\:\";'<>?,./ "},
			Enabled:      true,
			Timeout:      30,
		}

		// Test JSON serialization handles special characters correctly
		jsonData, err := json.Marshal(config)
		require.NoError(t, err)

		// Test deserialization preserves special characters
		var unmarshaled SlackConfig
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, config, unmarshaled)
	})

	t.Run("WebhookConfig with complex headers", func(t *testing.T) {
		config := WebhookConfig{
			URL:         "https://api.example.com/webhook?param=value&other=test",
			Method:      "POST",
			Headers:     map[string]string{"Content-Type": "application/json; charset=utf-8", "X-Custom-Header": "value with spaces and special_chars"},
			ContentType: "application/json; charset=utf-8",
			AuthToken:   "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			Enabled:     true,
			Timeout:     45,
		}

		// Test JSON serialization/deserialization with complex data
		jsonData, err := json.Marshal(config)
		require.NoError(t, err)

		var unmarshaled WebhookConfig
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, config, unmarshaled)
	})
}
