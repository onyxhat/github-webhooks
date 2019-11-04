// handlers_test.go
package main

import (
    "crypto/sha1"
    "bytes"
    "crypto/hmac"
    "encoding/hex"
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/google/go-github/github"
)

var repoName = "testing"
var repoOrg = "wyl1e"

func genMAC(message, key []byte) string {
    mac := hmac.New(sha1.New, key)
    mac.Write(message)
    sha1Hash := hex.EncodeToString(mac.Sum(nil))
    return fmt.Sprintf("sha1=%s", sha1Hash)
}

func testRepoExists() (bool, error) {
    ctx := context.Background()
    client := setupAuth(ctx)
    _, _, err := client.Repositories.Get(ctx, repoOrg, repoName)

    if err != nil {
        log.Printf("Could not find repo %s", repoName)
        return false, err
    }

    return true, nil
}

func destroyTestRepo(t *testing.T) {
    ctx := context.Background()
    client := setupAuth(ctx)
    _, err := client.Repositories.Delete(ctx, repoOrg, repoName)

    if err != nil {
        t.Fatalf("Error deleting repo %s, %v", repoName, err)
        return
    }
}

func createTestRepo(t *testing.T) {
     ctx := context.Background()
     client := setupAuth(ctx)
     repo := &github.Repository {
        Name: github.String(repoName),
        AutoInit: github.Bool(true),
     }

     _, _, err := client.Repositories.Create(ctx, repoOrg, repo)

     if err != nil {
        t.Fatalf("Error creating repo %s, %v", repoName, err)
     }
}

func TestFullIntegration(t *testing.T) {
    repoExists, err := testRepoExists()

    if repoExists {
        destroyTestRepo(t)
        createTestRepo(t)
    } else {
        createTestRepo(t)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(handleWebhook)
    file, err := os.Open("repository_created.json")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    b, err := ioutil.ReadAll(file)
    req, err := http.NewRequest("POST", "/webhook", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    secret := []byte(os.Getenv("GITHUB_WEBHOOK_SECRET"))
    signature := genMAC(b, secret)
    req.Header.Set("X-Hub-Signature", signature)
    req.Header.Set("X-Github-Event", "repository")

    if err != nil {
        t.Fatal(err)
    }
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `"wyl1e/testing"`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }

}
