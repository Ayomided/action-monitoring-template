// This Go program is part of the "action-monitoring-template" repository, designed to benchmark the performance
// of CI/CD pipelines, specifically focusing on self-hosted GitHub Action runners. It uses the GitHub API to:
// 1. Retrieve a list of self-hosted runners for a specified GitHub repository.
// 2. Fetch workflow run logs for each runner in the repository.
// 3. Save the workflow logs in ZIP format to the local file system, organized by runner and job run ID.
// The collected data can then be used for performance analysis, troubleshooting, and optimizing CI/CD workflows.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v63/github"
	"github.com/subosito/gotenv"
	"golang.org/x/oauth2"
)

func main() {
	if err := gotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Repo Owner Name
	owner := "ayomided"

	// Repository
	repo := "ci-cd-performance-benchmark"

	runners, _, err := client.Actions.ListRunners(ctx, owner, repo, nil)
	if err != nil {
		fmt.Printf("Error listing runners: %v\n", err)
		return
	}

	// Steps -> StepNumber(Order), JobID/StepID?, Name, Status(Completed || ), Conclusion(Success || Skipped || Failure)

	// Print details of each self-hosted runner
	// Get Log Files
	for _, runner := range runners.Runners {
		fmt.Printf("Runner ID: %d\n", *runner.ID)

		workflowRuns, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, nil)
		if err != nil {
			fmt.Printf("Error listing workflow runs: %v\n", err)
			continue
		}

		if _, err = os.ReadDir("runLogs"); err != nil {
			err = os.Mkdir("runLogs", 0750)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, run := range workflowRuns.WorkflowRuns {
			jobs, _, err := client.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, nil)
			if err != nil {
				fmt.Printf("Error listing workflow runs: %v\n", err)
				continue
			}
			for _, job := range jobs.Jobs {
				jobId := *job.ID
				runId := *run.ID
				runAttempt := *run.RunAttempt
				status := *run.Status
				conclusion := *run.Conclusion
				name := *job.Name
				startAt := *job.StartedAt
				completedAt := *job.CompletedAt
				fmt.Printf("%-10s %-10s %-10s %-12s %-12s %-20s %-20s %-20s\n",
					"JobID", "RunID", "Attempt", "Status", "Conclusion", "JobName", "StartedAt", "CompletedAt")

				fmt.Printf("%-10d %-10d %-10d %-12s %-12s %-20s %-20v %-20v\n",
					jobId, runId, runAttempt, status, conclusion, name, startAt, completedAt)
			}
		}
	}
}

func getRunLogs(ctx context.Context, owner, repo string, run *github.WorkflowRun, runner *github.Runner, client *github.Client) {
	logs, _, err := client.Actions.GetWorkflowRunLogs(ctx, owner, repo, *run.ID, 0)
	if err != nil {
		fmt.Printf("Error getting workflow run logs: %v\n", err)
	}

	resp, err := http.Get(logs.String())
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("runLogs/runner-%d-jworkflowRun-%d.zip", *runner.ID, *run.ID)
	err = os.WriteFile(filename, body, 0644)

	if err != nil {
		log.Fatal(err)
	}
}
