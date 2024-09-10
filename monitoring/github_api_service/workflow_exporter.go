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

	owner := "ayomided"
	repo := "ci-cd-performance-benchmark"

	runners, _, err := client.Actions.ListRunners(ctx, owner, repo, nil)
	if err != nil {
		fmt.Printf("Error listing runners: %v\n", err)
		return
	}

	// Print details of each self-hosted runner
	for _, runner := range runners.Runners {
		fmt.Printf("Runner ID: %d\n", *runner.ID)
		// List workflow runs for the repository
		workflowRuns, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, nil)
		if err != nil {
			fmt.Printf("Error listing workflow runs: %v\n", err)
			continue
		}

		// Find the latest workflow run and get its ID
		// var latestRunID int64
		//
		//
		if _, err = os.ReadDir("runLogs"); err != nil {
			err = os.Mkdir("runLogs", 0750)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, run := range workflowRuns.WorkflowRuns {
			fmt.Printf("Run ID: %d\n", *run.ID)
			// jobs, _, err := client.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, nil)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// for _, job := range jobs.Jobs {
			// 	logUrl, _, err := client.Actions.GetWorkflowJobLogs(ctx, owner, repo, *job.ID, 0)
			// 	if err != nil {
			// 		fmt.Printf("Error getting workflow job logs: %v\n", err)
			// 		continue
			// 	}

			// 	resp, err := http.Get(logUrl.String())
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	defer resp.Body.Close()

			// 	body, err := io.ReadAll(resp.Body)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	filename := fmt.Sprintf("logs/%d.txt", *job.ID)
			// 	err = os.WriteFile(filename, body, 0644)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// }
			logs, _, err := client.Actions.GetWorkflowRunLogs(ctx, owner, repo, *run.ID, 0)
			if err != nil {
				fmt.Printf("Error getting workflow run logs: %v\n", err)
				continue
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

			filename := fmt.Sprintf("runLogs/runner-%d-jobRun-%d.zip", *runner.ID, *run.ID)
			err = os.WriteFile(filename, body, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

	}
}
