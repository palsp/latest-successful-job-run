package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
)

// https://github.com/actions/toolkit/blob/main/packages/core/src/core.ts
func getInput(inputName string, required bool) string {
	input := os.Getenv(fmt.Sprintf("INPUT_%s", strings.ReplaceAll(strings.ToUpper(inputName), " ", "_")))
	if required && strings.TrimSpace(input) == "" {
		panic(fmt.Sprintf("Input required and not supplied: %s", inputName))
	}
	return input
}

// Return the commit hash of the last workflow run in which the specified job was successful.
// Defaults to the commit hash of the latest commit if the job was never successful or if this was the first run.
func getLastSuccessfulWorkflowRunCommit(ctx context.Context, client *github.Client, jobName string) string {
	owner_repo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	owner := owner_repo[0]
	repo := owner_repo[1]
	previousWorkflowRuns, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, nil)
	if err != nil {
		log.Printf("Error getting workflow runs: %s", err)
		panic(err)
	}

	// iterate the list of workflow from newest to oldest,
	// if the workflow run contains the specified job and it was successful, return the commit hash
	for _, workflowRun := range previousWorkflowRuns.WorkflowRuns {
		if workflowRun.GetStatus() == "completed" {
			workflowRunJobs, _, err := client.Actions.ListWorkflowJobs(ctx, owner, repo, workflowRun.GetID(), nil)
			if err != nil {
				log.Printf("Error getting workflow jobs: %s", err)
				panic(err)
			}

			for _, workflowRunJob := range workflowRunJobs.Jobs {
				if workflowRunJob.GetName() == jobName && workflowRunJob.GetStatus() == "completed" && workflowRunJob.GetConclusion() == "success" {
					jobId := workflowRun.GetHeadCommit().GetID()
					log.Printf("The hash of the latest commit in which the specified job was successful: %s", jobId)
					return jobId
				}
			}
		}
	}

	// default to the commit hash of the latest commit
	log.Printf("Unable to find the specified job in successful state in any of the previous workflow runs, defaulting to the latest commit hash")
	return previousWorkflowRuns.WorkflowRuns[0].GetHeadCommit().GetID()
}

func main() {
	log.Print()
	log.Printf("Starting the action")

	ghClient := github.NewClient(nil)
	ctx := context.Background()

	input := getInput("paths", true)
	job := getInput("job", true)

	sha := getLastSuccessfulWorkflowRunCommit(ctx, ghClient, job)

	// grab its hash
	// get the current commit hash
	// see if the output of git diff contains the files that were changed

	log.Printf("Paths: %s", input)
	log.Printf("The commit hash of the last successful run of the specified job: %s", sha)

	// TODO diff to see the name of the files (or just make this return the sha)?
}
