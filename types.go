package main

import "github.com/google/go-github/github"

type config struct {
	BranchNames        []string
	Organization       string
	ProtectionSettings *github.ProtectionRequest
}
