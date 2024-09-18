package main

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-github/v63/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

// TestListRunners tests the listing of self-hosted runners.
func TestListRunners(t *testing.T) {
	// Set up mock environment for GitHub API
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock the ListRunners API response
	runnerResponse := `{
		"runners": [
			{
				"id": 12345,
				"name": "self-hosted-runner"
			}
		]
	}`
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/owner/repo/actions/runners",
		httpmock.NewStringResponder(200, runnerResponse))

	// Mock token and OAuth2 setup
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "dummy-token"})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List runners and validate the response
	runners, _, err := client.Actions.ListRunners(ctx, "owner", "repo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, runners)

	// Ensure the correct runner is returned
	assert.Equal(t, int64(12345), *runners.Runners[0].ID)
	assert.Equal(t, "self-hosted-runner", *runners.Runners[0].Name)
}

// TestListWorkflowRuns tests the retrieval of workflow runs.
func TestListWorkflowRuns(t *testing.T) {
	// Set up mock environment for GitHub API
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock the ListRepositoryWorkflowRuns API response
	workflowRunsResponse := `{
		"workflow_runs": [
			{
				"id": 56789,
				"run_attempt": 1,
				"status": "completed",
				"conclusion": "success"
			}
		]
	}`
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/owner/repo/actions/runs",
		httpmock.NewStringResponder(200, workflowRunsResponse))

	// Mock token and OAuth2 setup
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "dummy-token"})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List workflow runs and validate the response
	runs, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, "owner", "repo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, runs)

	// Ensure the correct workflow run is returned
	assert.Equal(t, int64(56789), *runs.WorkflowRuns[0].ID)
	assert.Equal(t, 1, *runs.WorkflowRuns[0].RunAttempt)
	assert.Equal(t, "completed", *runs.WorkflowRuns[0].Status)
	assert.Equal(t, "success", *runs.WorkflowRuns[0].Conclusion)
}

// TestGetRunLogs tests fetching and saving of run logs.
func TestGetRunLogs(t *testing.T) {
	// Set up mock environment for GitHub API
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock the GetWorkflowRunLogs API response
	logsURLResponse := `https://github.com/owner/repo/actions/runs/logs.zip`
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/owner/repo/actions/runs/56789/logs",
		httpmock.NewStringResponder(200, logsURLResponse))

	// Mock token and OAuth2 setup
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "dummy-token"})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create a temporary directory for saving logs
	tempDir := t.TempDir()
	err := os.MkdirAll(tempDir, os.ModePerm)
	assert.NoError(t, err)

	// Set up dummy runner and workflow run
	runner := &github.Runner{ID: github.Int64(12345)}
	run := &github.WorkflowRun{ID: github.Int64(56789)}

	// Mock the log download response
	httpmock.RegisterResponder("GET", logsURLResponse,
		httpmock.NewStringResponder(200, "dummy-log-data"))

	// Call getRunLogs and ensure the log is saved correctly
	getRunLogs(ctx, "owner", "repo", run, runner, client)

	// Validate that the log file is saved correctly
	logFile := tempDir + "/runner-12345-workflowRun-56789.zip"
	_, err = os.Stat(logFile)
	assert.NoError(t, err)
}
