package main

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
)

func createIssueWithProtectionDetails(protection *github.Protection, repoName string, repoOrg string) (*github.Issue, error) {
	//Setup authentication for GithubAPI
	ctx := context.Background()
	client := setupAuth(ctx)

	//Add intentation for branch protection details to "pretty print" in issue
	protectionDetails, err := json.MarshalIndent(protection, "", "    ")
	bodyString := fmt.Sprintf("Branch protection was automatically added to this repo with the following details:\n```%s\n```", protectionDetails)

	//Construct issue request and create issue
	issueRequest := &github.IssueRequest{
		Title: github.String("AUTO: Added branch protection"),
		Body:  github.String(bodyString),
	}
	log.Printf("Creating issue announcing branch protection for repo %s", repoName)
	issue, _, err := client.Issues.Create(ctx, repoOrg, repoName, issueRequest)

	if err != nil {
		log.Printf("Error creating issue %v", err)
		return nil, err
	}

	return issue, nil
}
