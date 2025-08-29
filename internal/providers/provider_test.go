package providers

import (
	"testing"
	"time"
)

func TestDefaultInternalProviderConfig(t *testing.T) {
	config := DefaultInternalProviderConfig()

	if config == nil {
		t.Fatal("DefaultInternalProviderConfig returned nil")
	}

	if !config.EnablePages {
		t.Error("Expected EnablePages to be true by default")
	}

	if !config.GenerateNavigation {
		t.Error("Expected GenerateNavigation to be true by default")
	}

	if !config.EnableTrends {
		t.Error("Expected EnableTrends to be true by default")
	}

	if len(config.CleanupPatterns) == 0 {
		t.Error("Expected CleanupPatterns to have default values")
	}

	if config.VerificationTimeout != 30*time.Second {
		t.Errorf("Expected VerificationTimeout to be 30s, got %v", config.VerificationTimeout)
	}
}

func TestDefaultCodecovProviderConfig(t *testing.T) {
	config := DefaultCodecovProviderConfig()

	if config == nil {
		t.Fatal("DefaultCodecovProviderConfig returned nil")
	}

	if config.APIURL != "https://codecov.io" {
		t.Errorf("Expected APIURL to be https://codecov.io, got %s", config.APIURL)
	}

	if !config.EnablePRComments {
		t.Error("Expected EnablePRComments to be true by default")
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout to be 60s, got %v", config.Timeout)
	}

	if config.Flags == nil {
		t.Error("Expected Flags to be initialized")
	}
}

func TestInternalProvider_Name(t *testing.T) {
	provider := NewInternalProvider(DefaultInternalProviderConfig())
	if provider.Name() != "internal" {
		t.Errorf("Expected provider name 'internal', got '%s'", provider.Name())
	}
}

func TestInternalProvider_Capabilities(t *testing.T) {
	provider := NewInternalProvider(DefaultInternalProviderConfig())
	capabilities := provider.Capabilities()

	if !capabilities.SupportsHistory {
		t.Error("Expected internal provider to support history")
	}

	if !capabilities.SupportsBadges {
		t.Error("Expected internal provider to support badges")
	}

	if !capabilities.SupportsReports {
		t.Error("Expected internal provider to support reports")
	}

	if !capabilities.SupportsDeployment {
		t.Error("Expected internal provider to support deployment")
	}

	if !capabilities.RequiresToken {
		t.Error("Expected internal provider to require token")
	}
}

func TestInternalProvider_Validate(t *testing.T) {
	tests := []struct {
		name        string
		provider    *InternalProvider
		config      *Config
		expectError bool
	}{
		{
			name:     "Valid configuration",
			provider: NewInternalProvider(DefaultInternalProviderConfig()),
			config: &Config{
				GitHubContext: &GitHubContext{
					Repository: "owner/repo",
					Owner:      "owner",
					Repo:       "repo",
					Token:      "fake-token",
				},
			},
			expectError: false,
		},
		{
			name:        "No provider configuration",
			provider:    &InternalProvider{},
			expectError: true,
		},
		{
			name:        "No GitHub context",
			provider:    NewInternalProvider(DefaultInternalProviderConfig()),
			config:      &Config{},
			expectError: true,
		},
		{
			name:     "Missing repository",
			provider: NewInternalProvider(DefaultInternalProviderConfig()),
			config: &Config{
				GitHubContext: &GitHubContext{
					Token: "fake-token",
				},
			},
			expectError: true,
		},
		{
			name:     "Missing token",
			provider: NewInternalProvider(DefaultInternalProviderConfig()),
			config: &Config{
				GitHubContext: &GitHubContext{
					Repository: "owner/repo",
					Owner:      "owner",
					Repo:       "repo",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				tt.provider.providerConfig = tt.config
			}

			err := tt.provider.Validate()

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

func TestCodecovProvider_Name(t *testing.T) {
	provider := NewCodecovProvider(DefaultCodecovProviderConfig())
	if provider.Name() != "codecov" {
		t.Errorf("Expected provider name 'codecov', got '%s'", provider.Name())
	}
}

func TestCodecovProvider_Capabilities(t *testing.T) {
	config := DefaultCodecovProviderConfig()
	config.EnablePRComments = true
	provider := NewCodecovProvider(config)
	capabilities := provider.Capabilities()

	if !capabilities.SupportsHistory {
		t.Error("Expected codecov provider to support history")
	}

	if !capabilities.SupportsBadges {
		t.Error("Expected codecov provider to support badges")
	}

	if !capabilities.SupportsReports {
		t.Error("Expected codecov provider to support reports")
	}

	if capabilities.SupportsDeployment {
		t.Error("Expected codecov provider to not support deployment")
	}

	if !capabilities.SupportsPRComments {
		t.Error("Expected codecov provider to support PR comments when enabled")
	}

	if !capabilities.RequiresToken {
		t.Error("Expected codecov provider to require token")
	}
}

func TestCodecovProvider_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *CodecovProviderConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: &CodecovProviderConfig{
				Token:   "fake-token",
				APIURL:  "https://codecov.io",
				Timeout: 60 * time.Second,
			},
			expectError: false,
		},
		{
			name:        "No configuration",
			config:      nil,
			expectError: true,
		},
		{
			name: "Missing token",
			config: &CodecovProviderConfig{
				APIURL:  "https://codecov.io",
				Timeout: 60 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Missing API URL",
			config: &CodecovProviderConfig{
				Token:   "fake-token",
				Timeout: 60 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Invalid timeout",
			config: &CodecovProviderConfig{
				Token:   "fake-token",
				APIURL:  "https://codecov.io",
				Timeout: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCodecovProvider(tt.config)
			err := provider.Validate()

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

func TestCoverageData_Structure(t *testing.T) {
	// Test that CoverageData structure is properly defined
	coverage := &CoverageData{
		Percentage:   85.5,
		TotalLines:   1000,
		CoveredLines: 855,
		Timestamp:    time.Now(),
		CommitSHA:    "abc123",
		Branch:       "main",
		Packages: []PackageCoverage{
			{
				Name:         "main",
				Coverage:     90.0,
				TotalLines:   500,
				CoveredLines: 450,
				Files:        []string{"main.go"},
			},
		},
		Files: []FileCoverage{
			{
				Filename:     "main.go",
				Coverage:     90.0,
				TotalLines:   500,
				CoveredLines: 450,
				MissedLines:  []int{10, 20, 30},
			},
		},
	}

	if coverage.Percentage != 85.5 {
		t.Errorf("Expected percentage 85.5, got %f", coverage.Percentage)
	}

	if len(coverage.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(coverage.Packages))
	}

	if len(coverage.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(coverage.Files))
	}

	if coverage.Files[0].Filename != "main.go" {
		t.Errorf("Expected filename 'main.go', got '%s'", coverage.Files[0].Filename)
	}

	if len(coverage.Files[0].MissedLines) != 3 {
		t.Errorf("Expected 3 missed lines, got %d", len(coverage.Files[0].MissedLines))
	}
}
