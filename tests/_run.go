package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
	"github.com/onyxhat/github-webhooks"
)

var repoName = "testing"
var repoOrg = "wyl1e"

// https://github.com/google/go-github/blob/e8bc002390592dcb5ffe203acf8593ab7651eeba/github/messages.go#L89
func genMAC(message, key []byte) string {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	sha1Hash := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha1=%s", sha1Hash)
}

func testRepoExists() (bool, error) {
	ctx := context.Background()
	client := setupAuth(ctx)
	log.Printf("Determining if repository %s exists already", repoName)
	_, _, err := client.Repositories.Get(ctx, repoOrg, repoName)

	if err != nil {
		log.Printf("Could not find repo %s", repoName)
		return false, err
	}

	return true, nil
}

func destroyTestRepo(t *testing.T, r string) {
	ctx := context.Background()
	client := setupAuth(ctx)
	log.Printf("Destroying repo %s", r)
	_, err := client.Repositories.Delete(ctx, repoOrg, r)

	if err != nil {
		t.Fatalf("Error deleting repo %s, %v", r, err)
		return
	}
}

func createTestRepo(t *testing.T, r string, autoinit bool) {
	ctx := context.Background()
	client := setupAuth(ctx)
	repo := &github.Repository{
		Name:     github.String(r),
		AutoInit: github.Bool(autoinit),
	}

	log.Printf("Creating repo %s", r)
	_, _, err := client.Repositories.Create(ctx, repoOrg, repo)

	if err != nil {
		t.Fatalf("Error creating repo %s, %v", r, err)
		return
	}
}

func TestGetBranch(t *testing.T) {
	ctx := context.Background()
	client := setupAuth(ctx)
	createTestRepo(t, "branchtest", true)
	branch, response, err := getBranch(ctx, client, repoOrg, "branchtest", "master")

	if err != nil {
		t.Fatalf("There was an error with Repositories.GetBranch(): %v", err)
	}

	if response.StatusCode == 404 {
		t.Fatalf("Failed to find master branch!")
	}

	if *branch.Protected {
		t.Fatalf("Branch is somehow already protected!")
	}

	destroyTestRepo(t, "branchtest")
}

func TestFullIntegration(t *testing.T) {
	//Determine if testing repo exists already and delete it if it does
	repoExists, err := testRepoExists()
	if repoExists {
		destroyTestRepo(t, repoName)
		createTestRepo(t, repoName, true)
	} else {
		createTestRepo(t, repoName, true)
	}

	//Setup httptest with main.handleWebhook and read test repository event payload
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleWebhook)
	file, err := os.Open("repository_created.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	//Read test repository event file and post it to the test http handler
	b, err := ioutil.ReadAll(file)
	req, err := http.NewRequest("POST", "/webhook", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	secret := []byte(os.Getenv("GITHUB_WEBHOOK_SECRET"))

	//Generate the signature of the payload to pass github.ValidatePayload
	signature := genMAC(b, secret)
	req.Header.Set("X-Hub-Signature", signature)
	req.Header.Set("X-Github-Event", "repository")

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	//Check that the status code is 200
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `"Successfully added branch protection for repo testing"`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

}
