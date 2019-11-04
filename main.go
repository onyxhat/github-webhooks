package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
)

func setupAuth(ctx context.Context) *github.Client{
    //Setup GitHub API authentication with OAuth and return github.Client
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
    )
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    return client
}

func createIssueWithProtectionDetails(protection *github.Protection, repoName string, repoOrg string) (*github.Issue, error){
    //Setup authentication for GithubAPI
    ctx := context.Background()
    client := setupAuth(ctx)

    //Add intentation for branch protection details to "pretty print" in issue
    protectionDetails, err := json.MarshalIndent(protection, "", "    ")
    bodyString := fmt.Sprintf("@hobbsh, branch protection was automatically added to this repo with the following details:\n```%s\n```", protectionDetails) 

    //Construct issue request and create issue
    issueRequest := &github.IssueRequest{
        Title: github.String("AUTO: Added branch protection"),
        Body: github.String(bodyString),
    }
    log.Printf("Creating issue announcing branch protection for repo %s", repoName)
    issue, _, err := client.Issues.Create(ctx, repoOrg, repoName, issueRequest)

    if err != nil {
        log.Printf("Error creating issue %v", err)
        return nil, err
    }

    return issue, nil
}

func addBranchProtection(w http.ResponseWriter, repoName string, repoOrg string) (*github.Protection, error){
    //Setup authentication for GitHub API
    ctx := context.Background()
    client := setupAuth(ctx)
    branchName := "master"

    //Added 2 second delay because the branch would sometimes not exist before adding branch protection was attempted
    //Not ideal but it works around the issue
    time.Sleep(2*time.Second)
    branch, _, err := client.Repositories.GetBranch(ctx, repoOrg, repoName, branchName)

    if err != nil {
        log.Printf("Repositories.GetBranch() returned error: %v", err)
        return nil, err
    }

    // Return if branch is already protected
    if *branch.Protected {
        log.Printf("Branch %v of repo %v is already protected", branchName, repoName)
        return nil, err
    }

    //Setup protection request: https://godoc.org/github.com/google/go-github/github#ProtectionRequest
    protectionRequest := &github.ProtectionRequest{
        RequiredStatusChecks: &github.RequiredStatusChecks{
            Strict:   true,
            Contexts: []string{"continuous-integration"},
        },
        RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
            DismissStaleReviews: true,
            RequireCodeOwnerReviews: true,
            RequiredApprovingReviewCount: 2,
        },
        EnforceAdmins: true,
    }

    //Enable branch protection on the master branch
    protection, _, err := client.Repositories.UpdateBranchProtection(ctx, repoOrg, repoName, branchName, protectionRequest)
    if err != nil {
        log.Printf("Repositories.UpdateBranchProtection() returned error: %v", err)
        return nil, err
    }

    return protection, nil
}

// Response functions from: https://github.com/krishbhanushali/go-rest-unit-testing/blob/master/api.go
// RespondWithError is called on an error to return info regarding error
func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

// Called for responses to encode and send json data
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    //encode payload to json
    response, _ := json.Marshal(payload)

    // set headers and write response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

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
            log.Printf("Adding branch protection for repository: %s with owner: %s", repoName, repoOrg)
            protection, err := addBranchProtection(w, repoName, repoOrg)

            // If there was a problem adding branch protection, return 400.
            // Otherwise, if protection is added, create the issue in the repo with the protection details
            if err != nil {
                respondWithError(w, http.StatusBadRequest, fmt.Sprintf("There was a problem adding branch protection: %v",err))
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

func main() {
    log.Println("server started")
    http.HandleFunc("/webhook", handleWebhook)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
