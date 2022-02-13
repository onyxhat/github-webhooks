package main

import (
	"context"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
	"github.com/lestrrat-go/backoff"
)

func branchFilter(diff []*github.Branch, ref []string) ([]*github.Branch, error) {
	var inBoth []*github.Branch

	for _, d := range diff {
		for _, r := range ref {
			if *d.Name == r {
				inBoth = append(inBoth, d)
			}
		}
	}

	if len(inBoth) == 0 {
		return nil, errors.New("no matching branches")
	}

	return inBoth, nil
}

func getAllBranches(ctx context.Context, client *github.Client, cfg config, repoName string) ([]*github.Branch, *github.Response, error) {
	b, cancel := policy.Start(context.Background())
	defer cancel()
	var retries int

	for backoff.Continue(b) {
		branches, response, err := client.Repositories.ListBranches(ctx, cfg.Organization, repoName, nil)
		if err == nil && response.StatusCode == 200 && len(branches) > 0 {
			return branches, response, err
		}
		retries++
		log.Infof("Retrying (%v) getAllBranches in repo %s", retries, repoName)
	}

	return nil, nil, errors.New("failed to retrieve list of branches")
}

func addBranchProtection(w http.ResponseWriter, repoName string, repoOrg string) ([]*github.Protection, error) {
	//Setup authentication for GitHub API
	ctx := context.Background()
	client := setupAuth(ctx)

	var protections []*github.Protection

	//Get a list of all branches on repo
	allBranches, _, err := getAllBranches(ctx, client, defaultConfig, repoName)
	if err != nil {
		log.Warnf("%s: %v", repoName, err)
		return nil, err
	}

	branches, err := branchFilter(allBranches, defaultConfig.BranchNames)
	if err != nil {
		log.Warnf("%s - %v", repoName, err)
		return nil, err
	}

	for _, b := range branches {
		if *b.Protected {
			log.Warnf("%s is already protected", *b.Name)
		}

		//Setup protection request: https://godoc.org/github.com/google/go-github/github#ProtectionRequest
		protectionRequest, _, err := client.Repositories.UpdateBranchProtection(ctx, repoOrg, repoName, *b.Name, defaultConfig.ProtectionSettings)
		if err != nil {
			log.Errorf("Repositories.UpdateBranchProtection() returned error: %v", err)
		} else {
			protections = append(protections, protectionRequest)
		}
	}

	if len(protections) == 0 {
		return nil, errors.New("no protection requests were successful")
	}

	return protections, nil
}
