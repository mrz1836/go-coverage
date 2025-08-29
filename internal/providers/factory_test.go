package providers

import (
	"context"
	"os"
	"testing"

	"github.com/mrz1836/go-coverage/internal/config"
)

func TestFactory_CreateProvider(t *testing.T) {
	logger := NewDefaultLogger(false, false)
	factory := NewFactory(logger)

	tests := []struct {
		name         string
		config       *Config
		envVars      map[string]string
		expectedType string
		expectError  bool
	}{
		{
			name: "Create internal provider explicitly",
			config: &Config{
				Provider: ProviderTypeInternal,
				GitHubContext: &GitHubContext{
					Repository: "owner/repo",
					Owner:      "owner",
					Repo:       "repo",
					Token:      "fake-token",
				},
				InternalConfig: DefaultInternalProviderConfig(),
			},
			expectedType: "internal",
			expectError:  false,
		},
		{
			name: "Create codecov provider with token",
			config: &Config{
				Provider:      ProviderTypeCodecov,
				CodecovConfig: &CodecovProviderConfig{Token: "fake-codecov-token"},
			},
			envVars:      map[string]string{"CODECOV_TOKEN": "fake-codecov-token"},
			expectedType: "codecov",
			expectError:  false,
		},
		{
			name: "Auto-detect internal provider",
			config: &Config{
				Provider: ProviderTypeAuto,
				GitHubContext: &GitHubContext{
					IsGitHubActions: true,
					Repository:      "owner/repo",
					Owner:           "owner",
					Repo:            "repo",
					Token:           "fake-token",
				},
				InternalConfig: DefaultInternalProviderConfig(),
			},
			expectedType: "internal",
			expectError:  false,
		},
		{
			name: "Auto-detect codecov provider",
			config: &Config{
				Provider:      ProviderTypeAuto,
				CodecovConfig: DefaultCodecovProviderConfig(),
			},
			envVars:      map[string]string{"CODECOV_TOKEN": "fake-codecov-token"},
			expectedType: "codecov",
			expectError:  false,
		},
		{
			name: "Unsupported provider type",
			config: &Config{
				Provider: ProviderType("unsupported"),
			},
			expectError: true,
		},
		{
			name:        "Nil configuration",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				originalValue := os.Getenv(key)
				_ = os.Setenv(key, value)
				defer func(k, v string) {
					if v == "" {
						_ = os.Unsetenv(k)
					} else {
						_ = os.Setenv(k, v)
					}
				}(key, originalValue)
			}

			ctx := context.Background()
			provider, err := factory.CreateProvider(ctx, tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("Expected provider, but got nil")
				return
			}

			if provider.Name() != tt.expectedType {
				t.Errorf("Expected provider type %s, got %s", tt.expectedType, provider.Name())
			}
		})
	}
}

func TestFactory_ValidateProviderType(t *testing.T) {
	logger := NewDefaultLogger(false, false)
	factory := NewFactory(logger)

	tests := []struct {
		name         string
		providerType ProviderType
		expectError  bool
	}{
		{"Valid auto", ProviderTypeAuto, false},
		{"Valid internal", ProviderTypeInternal, false},
		{"Valid codecov", ProviderTypeCodecov, false},
		{"Invalid provider", ProviderType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateProviderType(tt.providerType)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCreateConfigFromEnvironment(t *testing.T) {
	// Create a basic main config
	mainCfg := config.Load()

	tests := []struct {
		name        string
		envVars     map[string]string
		provider    ProviderType
		dryRun      bool
		debug       bool
		force       bool
		expectError bool
	}{
		{
			name: "Create config with GitHub environment",
			envVars: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_SHA":        "abc123",
				"GITHUB_TOKEN":      "token123",
				"GITHUB_REF_NAME":   "main",
			},
			provider:    ProviderTypeInternal,
			dryRun:      false,
			debug:       true,
			force:       false,
			expectError: false,
		},
		{
			name: "Create config with Codecov environment",
			envVars: map[string]string{
				"CODECOV_TOKEN": "codecov-token",
			},
			provider:    ProviderTypeCodecov,
			dryRun:      true,
			debug:       false,
			force:       true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				originalValue := os.Getenv(key)
				_ = os.Setenv(key, value)
				defer func(k, v string) {
					if v == "" {
						_ = os.Unsetenv(k)
					} else {
						_ = os.Setenv(k, v)
					}
				}(key, originalValue)
			}

			cfg, err := CreateConfigFromEnvironment(mainCfg, tt.provider, tt.dryRun, tt.debug, tt.force)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cfg == nil {
				t.Error("Expected config, but got nil")
				return
			}

			if cfg.Provider != tt.provider {
				t.Errorf("Expected provider %s, got %s", tt.provider, cfg.Provider)
			}

			if cfg.DryRun != tt.dryRun {
				t.Errorf("Expected dry run %v, got %v", tt.dryRun, cfg.DryRun)
			}

			if cfg.Debug != tt.debug {
				t.Errorf("Expected debug %v, got %v", tt.debug, cfg.Debug)
			}

			if cfg.Force != tt.force {
				t.Errorf("Expected force %v, got %v", tt.force, cfg.Force)
			}
		})
	}
}
