package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	token := "test-token"
	client := New(token)

	assert.NotNil(t, client)
	assert.Equal(t, token, client.token)
	assert.Equal(t, "https://api.github.com", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	assert.NotNil(t, client.config)
	assert.Equal(t, "coverage-system/1.0", client.config.UserAgent)
}

func TestNewWithConfig(t *testing.T) {
	config := &Config{
		Token:      "custom-token",
		BaseURL:    "https://custom.api.com",
		Timeout:    60 * time.Second,
		RetryCount: 5,
		UserAgent:  "custom-agent/2.0",
	}

	client := NewWithConfig(config)

	assert.NotNil(t, client)
	assert.Equal(t, config.Token, client.token)
	assert.Equal(t, config.BaseURL, client.baseURL)
	assert.Equal(t, config.Timeout, client.httpClient.Timeout)
	assert.Equal(t, config, client.config)
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name             string
		existingComments []Comment
		body             string
		expectedAction   string
		expectedComment  *Comment
		expectError      bool
	}{
		{
			name:             "create new comment when none exists",
			existingComments: []Comment{},
			body:             "<!-- coverage-comment -->\nNew coverage report",
			expectedAction:   "create",
			expectedComment: &Comment{
				ID:   123,
				Body: "<!-- coverage-comment -->\nNew coverage report",
			},
		},
		{
			name: "update existing coverage comment",
			existingComments: []Comment{
				{ID: 456, Body: "<!-- coverage-comment -->\nOld coverage report"},
				{ID: 789, Body: "Some other comment"},
			},
			body:           "<!-- coverage-comment -->\nUpdated coverage report",
			expectedAction: "update",
			expectedComment: &Comment{
				ID:   456,
				Body: "<!-- coverage-comment -->\nUpdated coverage report",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == "GET" && strings.Contains(r.URL.Path, "/comments"):
					// Return existing comments
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tt.existingComments)
				case r.Method == "POST" && strings.Contains(r.URL.Path, "/comments"):
					// Create new comment
					var req CommentRequest
					_ = json.NewDecoder(r.Body).Decode(&req)
					comment := Comment{
						ID:   123,
						Body: req.Body,
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(comment)
				case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/comments/"):
					// Update existing comment
					var req CommentRequest
					_ = json.NewDecoder(r.Body).Decode(&req)
					comment := Comment{
						ID:   456,
						Body: req.Body,
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(comment)
				default:
					t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			comment, err := client.CreateComment(ctx, "owner", "repo", 123, tt.body)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, comment)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, comment)
				assert.Equal(t, tt.expectedComment.ID, comment.ID)
				assert.Equal(t, tt.expectedComment.Body, comment.Body)
			}
		})
	}
}

func TestCreateCommentError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "unauthorized",
			statusCode:    401,
			responseBody:  `{"message": "Bad credentials"}`,
			expectedError: "GitHub API error: 401",
		},
		{
			name:          "not found",
			statusCode:    404,
			responseBody:  `{"message": "Not Found"}`,
			expectedError: "GitHub API error: 404",
		},
		{
			name:          "rate limited",
			statusCode:    429,
			responseBody:  `{"message": "API rate limit exceeded"}`,
			expectedError: "retryable HTTP error: 429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && strings.Contains(r.URL.Path, "/comments") {
					// Return empty comments array
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]Comment{})
				} else {
					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			comment, err := client.CreateComment(ctx, "owner", "repo", 123, "test body")

			require.Error(t, err)
			assert.Nil(t, comment)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestCreateStatus(t *testing.T) {
	tests := []struct {
		name         string
		status       *StatusRequest
		statusCode   int
		responseBody string
		expectError  bool
	}{
		{
			name: "success status",
			status: &StatusRequest{
				State:       StatusSuccess,
				TargetURL:   "https://example.com/coverage",
				Description: "Coverage: 85.0%",
				Context:     ContextCoverage,
			},
			statusCode:   201,
			responseBody: `{"state": "success"}`,
			expectError:  false,
		},
		{
			name: "failure status",
			status: &StatusRequest{
				State:       StatusFailure,
				TargetURL:   "https://example.com/coverage",
				Description: "Coverage below threshold",
				Context:     ContextCoverage,
			},
			statusCode:   201,
			responseBody: `{"state": "failure"}`,
			expectError:  false,
		},
		{
			name: "error response",
			status: &StatusRequest{
				State:   StatusSuccess,
				Context: ContextCoverage,
			},
			statusCode:   422,
			responseBody: `{"message": "Validation Failed"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/statuses/")
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

				var status StatusRequest
				err := json.NewDecoder(r.Body).Decode(&status)
				assert.NoError(t, err)
				assert.Equal(t, tt.status.State, status.State)
				assert.Equal(t, tt.status.Context, status.Context)

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			err := client.CreateStatus(ctx, "owner", "repo", "abc123", tt.status)

			if tt.expectError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPullRequest(t *testing.T) {
	tests := []struct {
		name         string
		prNumber     int
		statusCode   int
		responseBody string
		expectedPR   *PullRequest
		expectError  bool
	}{
		{
			name:       "successful retrieval",
			prNumber:   123,
			statusCode: 200,
			responseBody: `{
				"number": 123,
				"title": "Test PR",
				"state": "open",
				"head": {
					"sha": "abc123def456"
				}
			}`,
			expectedPR: &PullRequest{
				Number: 123,
				Title:  "Test PR",
				State:  "open",
				Head: struct {
					SHA string `json:"sha"`
				}{SHA: "abc123def456"},
			},
			expectError: false,
		},
		{
			name:         "not found",
			prNumber:     999,
			statusCode:   404,
			responseBody: `{"message": "Not Found"}`,
			expectedPR:   nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, fmt.Sprintf("/pulls/%d", tt.prNumber))
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			pr, err := client.GetPullRequest(ctx, "owner", "repo", tt.prNumber)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, pr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, pr)
				assert.Equal(t, tt.expectedPR.Number, pr.Number)
				assert.Equal(t, tt.expectedPR.Title, pr.Title)
				assert.Equal(t, tt.expectedPR.State, pr.State)
				assert.Equal(t, tt.expectedPR.Head.SHA, pr.Head.SHA)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	client := &Client{
		token:   "test-token",
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: &Config{
			UserAgent: "test-agent",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.CreateComment(ctx, "owner", "repo", 123, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestContainsCoverageMarker(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{
			name:     "contains coverage report marker",
			body:     "Some text\n## Coverage Report\nMore text",
			expected: true,
		},
		{
			name:     "contains HTML comment marker",
			body:     "<!-- coverage-comment -->\nCoverage data here",
			expected: true,
		},
		{
			name:     "contains emoji marker",
			body:     "📊 **Coverage** increased by 2%",
			expected: true,
		},
		{
			name:     "no coverage markers",
			body:     "This is just a regular comment",
			expected: false,
		},
		{
			name:     "empty body",
			body:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsCoverageMarker(tt.body)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "hello", "hello", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
		{"substring in middle", "hello beautiful world", "beautiful", true},
		{"not found", "hello world", "goodbye", false},
		{"empty substring", "hello", "", true},
		{"empty string", "", "hello", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIndexOfHelper(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"found at start", "hello world", "hello", 0},
		{"found at end", "hello world", "world", 6},
		{"found in middle", "hello beautiful world", "beautiful", 6},
		{"not found", "hello world", "goodbye", -1},
		{"empty substring", "hello", "", 0},
		{"empty string", "", "hello", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindCoverageCommentIntegration(t *testing.T) {
	comments := []Comment{
		{ID: 1, Body: "Regular comment"},
		{ID: 2, Body: "<!-- coverage-comment -->\nOld coverage data"},
		{ID: 3, Body: "Another regular comment"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/comments")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := &Client{
		token:   "test-token",
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: &Config{
			UserAgent: "test-agent",
		},
	}

	ctx := context.Background()
	comment, err := client.findCoverageComment(ctx, "owner", "repo", 123)

	require.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, 2, comment.ID)
	assert.Contains(t, comment.Body, "<!-- coverage-comment -->")
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, "success", StatusSuccess)
	assert.Equal(t, "failure", StatusFailure)
	assert.Equal(t, "error", StatusError)
	assert.Equal(t, "pending", StatusPending)

	assert.Equal(t, "coverage/total", ContextCoverage)
	assert.Equal(t, "coverage/trend", ContextTrend)
}

// Test workflow-related functions

func TestGetWorkflowRuns(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		statusCode    int
		responseBody  string
		expectedRuns  int
		expectedTotal int
		expectError   bool
	}{
		{
			name:       "successful retrieval with limit",
			limit:      5,
			statusCode: 200,
			responseBody: `{
				"total_count": 25,
				"workflow_runs": [
					{
						"id": 12345,
						"name": "CI",
						"status": "completed",
						"conclusion": "success",
						"head_sha": "abc123",
						"created_at": "2023-01-01T10:00:00Z",
						"updated_at": "2023-01-01T10:05:00Z",
						"run_started_at": "2023-01-01T10:01:00Z"
					},
					{
						"id": 12346,
						"name": "Tests",
						"status": "in_progress",
						"conclusion": null,
						"head_sha": "def456",
						"created_at": "2023-01-01T11:00:00Z",
						"updated_at": "2023-01-01T11:02:00Z",
						"run_started_at": "2023-01-01T11:01:00Z"
					}
				]
			}`,
			expectedRuns:  2,
			expectedTotal: 25,
			expectError:   false,
		},
		{
			name:       "successful retrieval without limit",
			limit:      0,
			statusCode: 200,
			responseBody: `{
				"total_count": 3,
				"workflow_runs": [
					{
						"id": 98765,
						"name": "Deploy",
						"status": "completed",
						"conclusion": "failure",
						"head_sha": "xyz789",
						"created_at": "2023-01-01T12:00:00Z",
						"updated_at": "2023-01-01T12:10:00Z",
						"run_started_at": "2023-01-01T12:01:00Z"
					}
				]
			}`,
			expectedRuns:  1,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:         "API error",
			limit:        10,
			statusCode:   403,
			responseBody: `{"message": "API rate limit exceeded"}`,
			expectError:  true,
		},
		{
			name:         "invalid JSON response",
			limit:        5,
			statusCode:   200,
			responseBody: `invalid json`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/actions/runs")
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

				// Check for per_page parameter if limit > 0
				if tt.limit > 0 {
					assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("per_page=%d", tt.limit))
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			response, err := client.GetWorkflowRuns(ctx, "owner", "repo", tt.limit)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedTotal, response.TotalCount)
				assert.Len(t, response.WorkflowRuns, tt.expectedRuns)

				if len(response.WorkflowRuns) > 0 {
					run := response.WorkflowRuns[0]
					assert.NotZero(t, run.ID)
					assert.NotEmpty(t, run.Name)
					assert.NotEmpty(t, run.Status)
					assert.NotEmpty(t, run.HeadSHA)
				}
			}
		})
	}
}

func TestGetWorkflowRunsByWorkflow(t *testing.T) {
	tests := []struct {
		name             string
		workflowName     string
		limit            int
		workflowsResp    string
		workflowRunsResp string
		expectedRuns     int
		expectError      bool
	}{
		{
			name:         "successful retrieval",
			workflowName: "CI",
			limit:        3,
			workflowsResp: `{
				"total_count": 2,
				"workflows": [
					{
						"id": 123,
						"name": "CI",
						"path": ".github/workflows/ci.yml",
						"state": "active",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z"
					},
					{
						"id": 456,
						"name": "Deploy",
						"path": ".github/workflows/deploy.yml",
						"state": "active",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z"
					}
				]
			}`,
			workflowRunsResp: `{
				"total_count": 5,
				"workflow_runs": [
					{
						"id": 111,
						"name": "CI",
						"workflow_id": 123,
						"status": "completed",
						"conclusion": "success",
						"head_sha": "abc123",
						"created_at": "2023-01-01T10:00:00Z",
						"updated_at": "2023-01-01T10:05:00Z",
						"run_started_at": "2023-01-01T10:01:00Z"
					}
				]
			}`,
			expectedRuns: 1,
			expectError:  false,
		},
		{
			name:         "workflow not found",
			workflowName: "NonExistent",
			limit:        5,
			workflowsResp: `{
				"total_count": 1,
				"workflows": [
					{
						"id": 123,
						"name": "CI",
						"path": ".github/workflows/ci.yml",
						"state": "active",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z"
					}
				]
			}`,
			expectError: true,
		},
		{
			name:         "empty workflow list",
			workflowName: "CI",
			limit:        5,
			workflowsResp: `{
				"total_count": 0,
				"workflows": []
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

				w.WriteHeader(200)
				if strings.Contains(r.URL.Path, "/workflows") && !strings.Contains(r.URL.Path, "/runs") {
					// First request: get workflows
					_, _ = w.Write([]byte(tt.workflowsResp))
				} else if strings.Contains(r.URL.Path, "/runs") {
					// Second request: get workflow runs (only if first succeeded)
					if !tt.expectError {
						assert.Contains(t, r.URL.Path, "/workflows/123/runs")
						if tt.limit > 0 {
							assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("per_page=%d", tt.limit))
						}
						_, _ = w.Write([]byte(tt.workflowRunsResp))
					}
				} else {
					t.Errorf("Unexpected request: %s", r.URL.Path)
				}
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			response, err := client.GetWorkflowRunsByWorkflow(ctx, "owner", "repo", tt.workflowName, tt.limit)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, response)
				if tt.workflowName == "NonExistent" || tt.workflowName == "CI" && strings.Contains(tt.workflowsResp, "[]") {
					assert.Contains(t, err.Error(), "workflow not found")
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.WorkflowRuns, tt.expectedRuns)
				assert.Equal(t, 2, requestCount) // Should make 2 requests: workflows + workflow runs

				if len(response.WorkflowRuns) > 0 {
					run := response.WorkflowRuns[0]
					assert.Equal(t, int64(111), run.ID)
					assert.Equal(t, "CI", run.Name)
					assert.Equal(t, int64(123), run.WorkflowID)
					assert.Equal(t, "completed", run.Status)
					assert.Equal(t, "success", run.Conclusion)
				}
			}
		})
	}
}

func TestGetWorkflowRun(t *testing.T) {
	tests := []struct {
		name         string
		runID        int64
		statusCode   int
		responseBody string
		expectedRun  *WorkflowRun
		expectError  bool
	}{
		{
			name:       "successful retrieval",
			runID:      12345,
			statusCode: 200,
			responseBody: `{
				"id": 12345,
				"node_id": "MDEyOldlcmVmcmVxdWVzdDI2M",
				"name": "CI",
				"head_branch": "main",
				"head_sha": "abc123def456",
				"path": ".github/workflows/ci.yml",
				"display_title": "Update README",
				"run_number": 42,
				"event": "push",
				"status": "completed",
				"conclusion": "success",
				"workflow_id": 456,
				"check_suite_id": 789,
				"check_suite_node_id": "MDEyOkNoZWNrU3VpdGUxMjM",
				"url": "https://api.github.com/repos/owner/repo/actions/runs/12345",
				"html_url": "https://github.com/owner/repo/actions/runs/12345",
				"created_at": "2023-01-01T10:00:00Z",
				"updated_at": "2023-01-01T10:05:00Z",
				"run_started_at": "2023-01-01T10:01:00Z",
				"jobs_url": "https://api.github.com/repos/owner/repo/actions/runs/12345/jobs",
				"logs_url": "https://api.github.com/repos/owner/repo/actions/runs/12345/logs",
				"check_suite_url": "https://api.github.com/repos/owner/repo/check-suites/789",
				"artifacts_url": "https://api.github.com/repos/owner/repo/actions/runs/12345/artifacts",
				"cancel_url": "https://api.github.com/repos/owner/repo/actions/runs/12345/cancel",
				"rerun_url": "https://api.github.com/repos/owner/repo/actions/runs/12345/rerun",
				"workflow_url": "https://api.github.com/repos/owner/repo/actions/workflows/456"
			}`,
			expectedRun: &WorkflowRun{
				ID:           12345,
				NodeID:       "MDEyOldlcmVmcmVxdWVzdDI2M",
				Name:         "CI",
				HeadBranch:   "main",
				HeadSHA:      "abc123def456",
				Path:         ".github/workflows/ci.yml",
				DisplayTitle: "Update README",
				RunNumber:    42,
				Event:        "push",
				Status:       "completed",
				Conclusion:   "success",
				WorkflowID:   456,
			},
			expectError: false,
		},
		{
			name:         "not found",
			runID:        99999,
			statusCode:   404,
			responseBody: `{"message": "Not Found"}`,
			expectError:  true,
		},
		{
			name:         "unauthorized",
			runID:        12345,
			statusCode:   401,
			responseBody: `{"message": "Bad credentials"}`,
			expectError:  true,
		},
		{
			name:         "invalid JSON response",
			runID:        12345,
			statusCode:   200,
			responseBody: `invalid json`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, fmt.Sprintf("/actions/runs/%d", tt.runID))
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				token:   "test-token",
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			run, err := client.GetWorkflowRun(ctx, "owner", "repo", tt.runID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, run)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, run)
				assert.Equal(t, tt.expectedRun.ID, run.ID)
				assert.Equal(t, tt.expectedRun.NodeID, run.NodeID)
				assert.Equal(t, tt.expectedRun.Name, run.Name)
				assert.Equal(t, tt.expectedRun.HeadBranch, run.HeadBranch)
				assert.Equal(t, tt.expectedRun.HeadSHA, run.HeadSHA)
				assert.Equal(t, tt.expectedRun.Path, run.Path)
				assert.Equal(t, tt.expectedRun.DisplayTitle, run.DisplayTitle)
				assert.Equal(t, tt.expectedRun.RunNumber, run.RunNumber)
				assert.Equal(t, tt.expectedRun.Event, run.Event)
				assert.Equal(t, tt.expectedRun.Status, run.Status)
				assert.Equal(t, tt.expectedRun.Conclusion, run.Conclusion)
				assert.Equal(t, tt.expectedRun.WorkflowID, run.WorkflowID)
			}
		})
	}
}
