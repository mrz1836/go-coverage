package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetWorkflowRuns_ErrorHandling tests error scenarios for GetWorkflowRuns
func TestGetWorkflowRuns_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		statusCode    int
		responseBody  string
		expectedError string
		networkError  bool
	}{
		{
			name:          "unauthorized access",
			limit:         10,
			statusCode:    401,
			responseBody:  `{"message": "Bad credentials", "documentation_url": "https://docs.github.com/rest"}`,
			expectedError: "GitHub API error: 401",
		},
		{
			name:          "repository not found",
			limit:         5,
			statusCode:    404,
			responseBody:  `{"message": "Not Found"}`,
			expectedError: "GitHub API error: 404",
		},
		{
			name:          "rate limit exceeded",
			limit:         20,
			statusCode:    403,
			responseBody:  `{"message": "API rate limit exceeded"}`,
			expectedError: "GitHub API error: 403",
		},
		{
			name:          "malformed JSON response",
			limit:         5,
			statusCode:    200,
			responseBody:  `{"total_count": "invalid", "workflow_runs":`,
			expectedError: "failed to decode workflow runs response",
		},
		{
			name:          "network error",
			networkError:  true,
			expectedError: "failed to get workflow runs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if !tt.networkError {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "GET", r.Method)
					assert.Contains(t, r.URL.Path, "/actions/runs")
					assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
					assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
					assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

					if tt.limit > 0 {
						assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("per_page=%d", tt.limit))
					}

					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()
			}

			client := &Client{
				token: "test-token",
				baseURL: func() string {
					if tt.networkError {
						return "http://nonexistent.localhost:99999"
					}
					return server.URL
				}(),
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			response, err := client.GetWorkflowRuns(ctx, "owner", "repo", tt.limit)

			require.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestGetWorkflowRuns_EdgeCases tests edge cases for GetWorkflowRuns
func TestGetWorkflowRuns_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		responseBody string
		expectedRuns int
	}{
		{
			name:  "empty workflow runs list",
			limit: 10,
			responseBody: `{
				"total_count": 0,
				"workflow_runs": []
			}`,
			expectedRuns: 0,
		},
		{
			name:  "zero limit should not add per_page parameter",
			limit: 0,
			responseBody: `{
				"total_count": 5,
				"workflow_runs": [
					{
						"id": 111,
						"name": "CI",
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
		},
		{
			name:  "negative limit should not add per_page parameter",
			limit: -5,
			responseBody: `{
				"total_count": 2,
				"workflow_runs": [
					{
						"id": 222,
						"name": "Test",
						"status": "in_progress",
						"conclusion": null,
						"head_sha": "def456",
						"created_at": "2023-01-01T11:00:00Z",
						"updated_at": "2023-01-01T11:02:00Z",
						"run_started_at": "2023-01-01T11:01:00Z"
					}
				]
			}`,
			expectedRuns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/actions/runs")

				// Check that per_page is not added for zero or negative limits
				if tt.limit <= 0 {
					assert.NotContains(t, r.URL.RawQuery, "per_page")
				} else {
					assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("per_page=%d", tt.limit))
				}

				w.WriteHeader(200)
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

			require.NoError(t, err)
			assert.NotNil(t, response)
			assert.Len(t, response.WorkflowRuns, tt.expectedRuns)
		})
	}
}

// TestGetWorkflowRunsByWorkflow_ErrorHandling tests error scenarios
func TestGetWorkflowRunsByWorkflow_ErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		workflowName     string
		limit            int
		workflowsResp    string
		workflowsStatus  int
		workflowRunsResp string
		runsStatus       int
		expectedError    string
		networkError     bool
	}{
		{
			name:            "failed to get workflows - API error",
			workflowName:    "CI",
			limit:           5,
			workflowsStatus: 403,
			workflowsResp:   `{"message": "Forbidden"}`,
			expectedError:   "failed to get workflow ID",
		},
		{
			name:            "workflow not found in list",
			workflowName:    "NonExistent",
			limit:           5,
			workflowsStatus: 200,
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
			expectedError: "workflow not found: NonExistent",
		},
		{
			name:            "empty workflow list",
			workflowName:    "CI",
			limit:           5,
			workflowsStatus: 200,
			workflowsResp: `{
				"total_count": 0,
				"workflows": []
			}`,
			expectedError: "workflow not found: CI",
		},
		{
			name:            "malformed workflows JSON",
			workflowName:    "CI",
			limit:           5,
			workflowsStatus: 200,
			workflowsResp:   `{"total_count": 1, "workflows":`,
			expectedError:   "failed to decode workflows response",
		},
		{
			name:            "failed to get workflow runs after finding workflow",
			workflowName:    "CI",
			limit:           5,
			workflowsStatus: 200,
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
			runsStatus:       500,
			workflowRunsResp: `{"message": "Internal Server Error"}`,
			expectedError:    "failed to get workflow runs",
		},
		{
			name:          "network error",
			workflowName:  "CI",
			networkError:  true,
			expectedError: "failed to get workflow ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if !tt.networkError {
				var requestCount int
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
					assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

					if strings.Contains(r.URL.Path, "/workflows") && !strings.Contains(r.URL.Path, "/runs") {
						// First request: get workflows
						w.WriteHeader(tt.workflowsStatus)
						_, _ = w.Write([]byte(tt.workflowsResp))
					} else if strings.Contains(r.URL.Path, "/runs") {
						// Second request: get workflow runs
						w.WriteHeader(tt.runsStatus)
						_, _ = w.Write([]byte(tt.workflowRunsResp))
					}
				}))
				defer server.Close()
			}

			client := &Client{
				token: "test-token",
				baseURL: func() string {
					if tt.networkError {
						return "http://nonexistent.localhost:99999"
					}
					return server.URL
				}(),
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			response, err := client.GetWorkflowRunsByWorkflow(ctx, "owner", "repo", tt.workflowName, tt.limit)

			require.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestGetWorkflowRun_ErrorHandling tests error scenarios for GetWorkflowRun
func TestGetWorkflowRun_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		runID         int64
		statusCode    int
		responseBody  string
		expectedError string
		networkError  bool
	}{
		{
			name:          "run not found",
			runID:         99999,
			statusCode:    404,
			responseBody:  `{"message": "Not Found"}`,
			expectedError: "GitHub API error: 404",
		},
		{
			name:          "unauthorized access to run",
			runID:         12345,
			statusCode:    401,
			responseBody:  `{"message": "Bad credentials"}`,
			expectedError: "GitHub API error: 401",
		},
		{
			name:          "forbidden access to run",
			runID:         12345,
			statusCode:    403,
			responseBody:  `{"message": "Forbidden"}`,
			expectedError: "GitHub API error: 403",
		},
		{
			name:          "malformed JSON response",
			runID:         12345,
			statusCode:    200,
			responseBody:  `{"id": "invalid", "name":`,
			expectedError: "failed to decode workflow run response",
		},
		{
			name:          "network error",
			runID:         12345,
			networkError:  true,
			expectedError: "failed to get workflow run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if !tt.networkError {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "GET", r.Method)
					assert.Contains(t, r.URL.Path, fmt.Sprintf("/actions/runs/%d", tt.runID))
					assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
					assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
					assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()
			}

			client := &Client{
				token: "test-token",
				baseURL: func() string {
					if tt.networkError {
						return "http://nonexistent.localhost:99999"
					}
					return server.URL
				}(),
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				config: &Config{
					UserAgent: "test-agent",
				},
			}

			ctx := context.Background()
			run, err := client.GetWorkflowRun(ctx, "owner", "repo", tt.runID)

			require.Error(t, err)
			assert.Nil(t, run)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestContextCancellationWorkflowFunctions tests context cancellation for all workflow functions
func TestContextCancellationWorkflowFunctions(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*Client, context.Context) error
	}{
		{
			name: "GetWorkflowRuns with canceled context",
			testFunc: func(client *Client, ctx context.Context) error {
				_, err := client.GetWorkflowRuns(ctx, "owner", "repo", 10)
				return err
			},
		},
		{
			name: "GetWorkflowRunsByWorkflow with canceled context",
			testFunc: func(client *Client, ctx context.Context) error {
				_, err := client.GetWorkflowRunsByWorkflow(ctx, "owner", "repo", "CI", 5)
				return err
			},
		},
		{
			name: "GetWorkflowRun with canceled context",
			testFunc: func(client *Client, ctx context.Context) error {
				_, err := client.GetWorkflowRun(ctx, "owner", "repo", 12345)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate slow response
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"total_count": 0, "workflow_runs": []}`))
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

			// Create context with short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := tt.testFunc(client, ctx)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		})
	}
}

// TestGetWorkflowIDByName_EdgeCases tests edge cases for the helper function
func TestGetWorkflowIDByName_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		workflowName  string
		responseBody  string
		statusCode    int
		expectedID    int64
		expectedError string
	}{
		{
			name:         "workflow found with exact match",
			workflowName: "CI",
			responseBody: `{
				"total_count": 3,
				"workflows": [
					{
						"id": 123,
						"name": "Build",
						"path": ".github/workflows/build.yml",
						"state": "active"
					},
					{
						"id": 456,
						"name": "CI",
						"path": ".github/workflows/ci.yml",
						"state": "active"
					},
					{
						"id": 789,
						"name": "Deploy",
						"path": ".github/workflows/deploy.yml",
						"state": "disabled"
					}
				]
			}`,
			statusCode: 200,
			expectedID: 456,
		},
		{
			name:         "workflow found - case sensitive",
			workflowName: "ci", // lowercase
			responseBody: `{
				"total_count": 1,
				"workflows": [
					{
						"id": 123,
						"name": "CI",
						"path": ".github/workflows/ci.yml",
						"state": "active"
					}
				]
			}`,
			statusCode:    200,
			expectedError: "workflow not found: ci",
		},
		{
			name:         "empty workflow name search",
			workflowName: "",
			responseBody: `{
				"total_count": 1,
				"workflows": [
					{
						"id": 123,
						"name": "CI",
						"path": ".github/workflows/ci.yml",
						"state": "active"
					}
				]
			}`,
			statusCode:    200,
			expectedError: "workflow not found: ",
		},
		{
			name:         "workflow with special characters",
			workflowName: "Test & Build (Linux)",
			responseBody: `{
				"total_count": 1,
				"workflows": [
					{
						"id": 999,
						"name": "Test & Build (Linux)",
						"path": ".github/workflows/test-build-linux.yml",
						"state": "active"
					}
				]
			}`,
			statusCode: 200,
			expectedID: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/actions/workflows")
				assert.NotContains(t, r.URL.Path, "/runs")
				assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

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
			workflowID, err := client.getWorkflowIDByName(ctx, "owner", "repo", tt.workflowName)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Zero(t, workflowID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, workflowID)
			}
		})
	}
}

// TestWorkflowRun_JSONUnmarshalEdgeCases tests JSON unmarshaling edge cases
func TestWorkflowRun_JSONUnmarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectedRun  *WorkflowRun
		expectError  bool
	}{
		{
			name: "run with null conclusion",
			responseBody: `{
				"id": 12345,
				"name": "CI",
				"status": "in_progress",
				"conclusion": null,
				"head_sha": "abc123",
				"workflow_id": 456,
				"created_at": "2023-01-01T10:00:00Z",
				"updated_at": "2023-01-01T10:05:00Z",
				"run_started_at": "2023-01-01T10:01:00Z"
			}`,
			expectedRun: &WorkflowRun{
				ID:         12345,
				Name:       "CI",
				Status:     "in_progress",
				Conclusion: "", // null becomes empty string
				HeadSHA:    "abc123",
				WorkflowID: 456,
			},
			expectError: false,
		},
		{
			name: "run with minimal required fields",
			responseBody: `{
				"id": 999,
				"name": "Minimal",
				"status": "completed",
				"conclusion": "success",
				"head_sha": "def456",
				"workflow_id": 111,
				"created_at": "2023-01-01T12:00:00Z",
				"updated_at": "2023-01-01T12:05:00Z",
				"run_started_at": "2023-01-01T12:01:00Z"
			}`,
			expectedRun: &WorkflowRun{
				ID:         999,
				Name:       "Minimal",
				Status:     "completed",
				Conclusion: "success",
				HeadSHA:    "def456",
				WorkflowID: 111,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
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
			run, err := client.GetWorkflowRun(ctx, "owner", "repo", tt.expectedRun.ID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, run)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, run)
				assert.Equal(t, tt.expectedRun.ID, run.ID)
				assert.Equal(t, tt.expectedRun.Name, run.Name)
				assert.Equal(t, tt.expectedRun.Status, run.Status)
				assert.Equal(t, tt.expectedRun.Conclusion, run.Conclusion)
				assert.Equal(t, tt.expectedRun.HeadSHA, run.HeadSHA)
				assert.Equal(t, tt.expectedRun.WorkflowID, run.WorkflowID)
			}
		})
	}
}
