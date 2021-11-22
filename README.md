# github-webhooks
Application that accepts github webhooks on repo create and sets branch protection  

## Overview
This project was forked from [hobbsh's](https://github.com/hobbsh/github-webhooks) with some minor changes to allow for using the new `main` branch default, decoupling the Go build from the Docker build and add some basic CICD and release patterns. In the future I would like to add handling for more than one branch, as well as configurable settings for each via a config file or configmap.

Binaries are provided in the releases - it's no longer necessary to do a local build or a bunch of setup to facilitate using `go run`.

## Prerequisistes
- Docker (optional)
- NGrok (or ability to expose/forward a port on your external IP)
- GitHub Account
- GitHub Organization
- Organization Webhook
- Webhook Secret
- Personal Access Token

## Setup your GitHub Access Token
- [Generate a new Personal Access Token](https://github.com/settings/tokens)
- Set the following permissions
  - `repo:*`
  - `delete_repo`
- Copy the displayed value - it will only be shown once. This will be `GITHUB_ACCESS_TOKEN`

## Generate a Webhook Secret
This can be any sufficiently random and long string of characters - this value needs to be stored into the env variable `GITHUB_WEBHOOK_SECRET` for the webservice and when creating the [Github Organization Webhook](#adding-your-organization-webhook)

## Running with docker
```bash
docker run -it --rm -e GITHUB_WEBHOOK_SECRET=$GITHUB_WEBHOOK_SECRET -e GITHUB_ACCESS_TOKEN=$GITHUB_ACCESS_TOKEN -p 8080:8080 onyxhat/github-webhooks:latest
```

## Running from the Command-Line
Mac/Linux
```bash
export GITHUB_WEBHOOK_SECRET={{ YOUR_SECRET_HERE }}
export GITHUB_ACCESS_TOKEN={{ YOUR_TOKEN_HERE }}
./github-webhooks-linux-amd64
```

Windows/PowerShell
```powershell
$env:GITHUB_WEBHOOK_SECRET = "{{ YOUR_SECRET_HERE }}"
$env:GITHUB_ACCESS_TOKEN = "{{ YOUR_TOKEN_HERE }}"
.\github-webhooks-windows-amd64.exe
```

## Exposing your Webservice
The application listens on port 8080 (unless you've changed the port mapping via Docker). For GitHub's webhook to reach your service this will need to be exposed to the outside world. The easiest method is to use ngrok for testing locally - setting up port forwarding will vary upon your network setup/hardware/configuration.

#### Using ngrok
- Install ngrok
  - `brew cask install ngrok` (if you don't have brew, click [here](https://brew.sh/).)
- Start ngrok on port 8080 `ngrok http 8080`
- An HTTPS forwarding URL will be displayed - this will be needing when creating your GitHub webhook. Append a path of `/webhook` to the end (e.g. `https://456ertg.ngrok.io/webhook`)

## Adding your Organization Webhook
- Go into your Organization Settings and Click on Webhooks `https://github.com/organizations/{{YOUR-ORG}}/settings/hooks`
- Click Add Webhook
- Using your webservice URL (e.g. if using ngrok [as above](#using-ngrok) `https://456ertg.ngrok.io/webhook`) as your `Payload URL`
- Choose `application/json` as your `Content type`
- Use the [webhook secret](#generate-a-webhook-secret) you created above for `Secret`
- Select `Let me select individual events` and mark the following events
  - `Repositories`

## Testing
There is a single end-to-end integration test included with this application that will test the entire process from start to finish, including creating/deleting a `testing` repo to perform the operations on. To run the test, be in the root of this repo and run:

`go test -v .`

You can navigate to the repo to validate that the branch protection has been enabled and that an issue explaining that has been created. The organization name is hardcoded into the test. The test will leave the testing repo in place for manual validation and destroys it at the beginning of each run.

### References
- https://github.com/hobbsh/github-webhooks
- https://groob.io/tutorial/go-github-webhook/
- https://github.com/krishbhanushali/go-rest-unit-testing
- A lot of the code in `go-github`, including tests: https://github.com/google/go-github/tree/master/test/integration
- StackOverflow