// This Go program is part of the "action-monitoring-template" repository,
// designed to benchmark the performance of CI/CD pipelines, focusing on
// self-hosted GitHub Action runners. It retrieves a list of runners, fetches
// workflow run logs, and saves them for analysis.
// The program is designed to be reusable for any repository by taking inputs
// for the GitHub repository owner and name.

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/go-github/v63/github"
	"github.com/subosito/gotenv"
	"golang.org/x/oauth2"
)

func main() {
	// Load environment variables from the .env file
	if err := gotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve GitHub token from environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GitHub token not provided. Set GITHUB_TOKEN in your environment")
	}

	// Define command-line flags for repository owner and name
	var owner string
	var repo string
	flag.StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	flag.StringVar(&repo, "repo", "", "GitHub repository name (required)")
	flag.Parse()

	// Ensure both flags are provided
	if owner == "" || repo == "" {
		log.Fatal("Both 'owner' and 'repo' flags are required")
	}

	// Set up OAuth2 authentication using the GitHub token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create CSV file for saving step data
	csvFile, err := os.Create("all_steps.csv")
	if err != nil {
		log.Fatalf("Error creating CSV file: %s", err)
	}
	defer csvFile.Close()

	// Set up CSV writer and write the header row
	csvWriter := csv.NewWriter(csvFile)
	if err := csvWriter.Write([]string{"JobID", "RunID", "Attempt", "Status", "Conclusion", "JobName", "StartedAt", "CompletedAt", "StepName"}); err == nil {
		log.Print("CSV header written successfully")
	}
	defer csvWriter.Flush()

	// List all self-hosted runners for the specified repository
	runners, _, err := client.Actions.ListRunners(ctx, owner, repo, nil)
	if err != nil {
		fmt.Printf("Error listing runners: %v\n", err)
		return
	}

	// Ensure directory for saving log files exists
	if _, err = os.ReadDir("runLogs"); err != nil {
		err = os.Mkdir("runLogs", 0750)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Iterate over each runner and retrieve workflow run details
	for _, runner := range runners.Runners {
		fmt.Printf("Runner ID: %d\n", *runner.ID)

		// Retrieve the workflow runs for the repository
		workflowRuns, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, nil)
		if err != nil {
			fmt.Printf("Error listing workflow runs: %v\n", err)
			continue
		}

		// For each workflow run, list the associated jobs and steps
		for _, run := range workflowRuns.WorkflowRuns {
			jobs, _, err := client.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, nil)
			if err != nil {
				fmt.Printf("Error listing workflow jobs: %v\n", err)
				continue
			}

			// Process each job and its details
			for _, job := range jobs.Jobs {
				// Extract job and run details
				jobName := *job.WorkflowName
				jobId := *job.ID
				runId := *run.ID
				runAttempt := *run.RunAttempt
				status := *run.Status
				conclusion := *run.Conclusion
				name := *job.Name
				startAt := *job.StartedAt
				completedAt := *job.CompletedAt

				// Prepare a CSV row with job and step information
				row := []string{
					strconv.Itoa(int(jobId)),
					strconv.FormatInt(runId, 10),
					strconv.Itoa(runAttempt),
					status,
					conclusion,
					jobName,
					startAt.String(),
					completedAt.String(),
					name,
				}

				// Write the row to the CSV file
				if err := csvWriter.Write(row); err != nil {
					log.Fatalf("Error writing row to CSV: %s", err)
				}
			}
		}
	}
}

// getRunLogs fetches and saves logs for a specific workflow run associated with a runner.
// The logs are saved in ZIP format, with filenames structured by runner ID and workflow run ID.
func getRunLogs(ctx context.Context, owner, repo string, run *github.WorkflowRun, runner *github.Runner, client *github.Client) {
	// Get the workflow run logs URL
	logs, _, err := client.Actions.GetWorkflowRunLogs(ctx, owner, repo, *run.ID, 0)
	if err != nil {
		fmt.Printf("Error getting workflow run logs: %v\n", err)
	}

	// Download the logs from the returned URL
	resp, err := http.Get(logs.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read the response body containing the log data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Save the logs as a ZIP file in the runLogs directory
	filename := fmt.Sprintf("runLogs/runner-%d-workflowRun-%d.zip", *runner.ID, *run.ID)
	err = os.WriteFile(filename, body, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
