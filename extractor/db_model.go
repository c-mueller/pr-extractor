package extractor

import (
	"github.com/jinzhu/gorm"
	"time"
)

type PullRequestReviewComment struct {
	gorm.Model

	EventDbId     string `gorm:"unique_index"`
	PullRequestId string `sql:"index"`
	RepoName      string `sql:"index"`
	RepoUrl       string
	PRUrl         string
	PRNumber      int

	ReviewId int `sql:"index"`

	CommentCreatedAt  time.Time
	CommentUpdatedAt  time.Time
	CommentAuthorName string
	CommentAuthorType string

	Body string

	RawPayload []byte
}

type PullRequest struct {
	gorm.Model

	EventDbId string `gorm:"unique_index"`

	PullRequestId string `sql:"index"`

	RepoName string `sql:"index"`
	RepoUrl  string
	PRUrl    string
	PRNumber int

	State         string
	PRAuthorLogin string
	PRAuthorType  string `sql:"index"`

	PullRequestCreatedAt time.Time
	PullRequestUpdatedAt *time.Time
	PullRequestClosedAt  *time.Time
	PullRequestMergedAt  *time.Time

	EventInitiatorLogin       string
	EventInitiatorDisplayName string

	Comments     int
	Commits      int
	Additions    int
	Deletions    int
	FilesChanged int

	EventTimestamp time.Time
	EventAction    string

	RawPayload []byte
}
