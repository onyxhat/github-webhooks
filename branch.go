package main

import (
	"context"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
	"github.com/lestrrat-go/backoff"
)

func getBranch(ctx context.Context, client *github.Client, repoOrg string, repoName string, branchName string) (*github.Branch, *github.Response, error) {
	b, cancel := policy.Start(context.Background())
	defer cancel()
	retries := 0
	//Use backoff for rety logic because Repositories.GetBranch sometimes returns 404s if the repo is slow to create
	for backoff.Continue(b) {
		branch, response, err := client.Repositories.GetBranch(ctx, repoOrg, repoName, branchName)
		if err == nil && response.StatusCode == 200 && branch != nil {
			return branch, response, err
		}
		retries++
		log.Printf("Retrying (%v) getBranch for %s in repo %s", retries, branchName, repoName)
	}

	return nil, nil, errors.New("Failed to find branch")
}

func addBranchProtection(w http.ResponseWriter, repoName string, repoOrg string, branchName string) (*github.Protection, error) {
	//Setup authentication for GitHub API
	ctx := context.Background()
	client := setupAuth(ctx)

	//Find out if branch exists and if it is already protected
	branch, _, err := getBranch(ctx, client, repoOrg, repoName, branchName)

	if err != nil {
		log.Printf("Error getting branch! %v", err)
		return nil, err
	}

	if *branch.Protected {
		log.Printf("Branch is already protected!")
		return nil, err
	}

	//Setup protection request: https://godoc.org/github.com/google/go-github/github#ProtectionRequest
	protectionRequest := defaultProtection

	//Enable branch protection on the main branch
	protection, _, err := client.Repositories.UpdateBranchProtection(ctx, repoOrg, repoName, branchName, protectionRequest)
	if err != nil {
		log.Printf("Repositories.UpdateBranchProtection() returned error: %v", err)
		return nil, err
	}

	return protection, nil
}
