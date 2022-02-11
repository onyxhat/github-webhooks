package main

import (
	"time"

	"github.com/google/go-github/github"
	"github.com/lestrrat-go/backoff"
)

// Settings for backoff libary used in getBranch()
var policy = backoff.NewExponential(
	backoff.WithInterval(100*time.Millisecond), // base interval
	backoff.WithJitterFactor(0.05),             // 5% jitter
	backoff.WithMaxRetries(25),                 // If not specified, default number of retries is 10
)

var defaultProtection = &github.ProtectionRequest{
	RequiredStatusChecks: &github.RequiredStatusChecks{
		Strict:   true,
		Contexts: []string{"continuous-integration"},
	},
	RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
		DismissStaleReviews:          true,
		RequireCodeOwnerReviews:      true,
		RequiredApprovingReviewCount: 2,
	},
	EnforceAdmins: true,
}
