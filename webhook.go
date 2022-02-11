package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
)

// Parts from https://groob.io/tutorial/go-github-webhook/
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Validate the payload
	// https://github.com/google/go-github/blob/e8bc002390592dcb5ffe203acf8593ab7651eeba/github/messages.go#L147
	payload, err := github.ValidatePayload(r, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Printf("Error validating request body: err=%s\n", err)
		respondWithError(w, http.StatusBadRequest, "Error validating request body")
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("Could not parse webhook: err=%s\n", err)
		respondWithError(w, http.StatusBadRequest, "Could not parse webhook")
		return
	}

	//If it's a repository create event, add branch protection. Else return 400.
	switch e := event.(type) {
	case *github.RepositoryEvent:
		if *e.Action == "created" {
			repoName := *e.Repo.Name
			repoOrg := *e.Repo.Owner.Login
			branchName := "main" //Temporary until picked up from config file and defaults defined.

			log.Printf("Adding branch protection for repository: %s with owner: %s", repoName, repoOrg)
			protection, err := addBranchProtection(w, repoName, repoOrg, branchName)

			// If there was a problem adding branch protection, return 400.
			// Otherwise, if protection is added, create the issue in the repo with the protection details
			if err != nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("There was a problem adding branch protection: %v", err))
			} else {
				if protection != nil {
					// Create the issue
					_, err := createIssueWithProtectionDetails(protection, repoName, repoOrg)
					if err != nil {
						respondWithError(w, http.StatusInternalServerError, "Could not create issue in repo")
					} else {
						respondWithJSON(w, http.StatusOK, fmt.Sprintf("Successfully added branch protection for repo %s", repoName))
					}
				} else {
					//Protection payload is nil, meaning that branch protection is already added
					respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Branch protection already added for repo '%s'", repoName))
				}
			}
		} else {
			// If the repository event is not a "create" event, return 204
			respondWithJSON(w, http.StatusNoContent, fmt.Sprintf("Repository event is %s, not a create event. Ignoring", *e.Action))
		}
		return
	default:
		// Default case - should not reach it if only the Repositories events type is selected in the webhook
		log.Printf("Unknown event type %s\n", github.WebHookType(r))
		respondWithError(w, http.StatusBadRequest, "Unknown event type: "+github.WebHookType(r))
		return
	}
}
