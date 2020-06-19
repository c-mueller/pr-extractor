package extractor

import "time"

type PREvent interface {
	GetRepoName() string
	GetPullRequestNumber() int
	GetPullRequestURL() string
	GetEventTimestamp() time.Time
}

func (p PullRequestEvent) GetRepoName() string {
	return p.Repo.Name
}

func (p PullRequestEvent) GetPullRequestNumber() int {
	return p.Payload.Number
}

func (p PullRequestEvent) GetPullRequestURL() string {
	return p.Payload.PullRequest.URL
}

func (p PullRequestEvent) GetEventTimestamp() time.Time {
	return p.CreatedAt
}

func (p PRReviewCommentEvent) GetRepoName() string {
	return p.Repo.Name
}

func (p PRReviewCommentEvent) GetPullRequestNumber() int {
	return p.Payload.PullRequest.Number
}

func (p PRReviewCommentEvent) GetPullRequestURL() string {
	return p.Payload.PullRequest.URL
}

func (p PRReviewCommentEvent) GetEventTimestamp() time.Time {
	return p.CreatedAt
}

func (p PRCommentEvent) GetRepoName() string {
	return p.Repo.Name
}

func (p PRCommentEvent) GetPullRequestNumber() int {
	return p.Payload.Issue.Number
}

func (p PRCommentEvent) GetPullRequestURL() string {
	return p.Payload.Issue.URL
}

func (p PRCommentEvent) GetEventTimestamp() time.Time {
	return p.CreatedAt
}
