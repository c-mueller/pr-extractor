package extractor

import (
	"github.com/jinzhu/gorm"
	"time"
)

type PullRequestComment struct {
	gorm.Model

	EventDbId     string `gorm:"unique_index"`
	PullRequestId string `sql:"index"`
	RepoName      string `sql:"index"`
	RepoUrl       string
	PRUrl         string
	PRNumber      int `sql:"index"`

	CommentCreatedAt  time.Time
	CommentUpdatedAt  time.Time
	CommentAuthorName string `sql:"index"`
	CommentAuthorType string `sql:"index"`

	Body string

	EventTimestamp time.Time `sql:"index"`
	EventAction    string    `sql:"index"`

	RawPayload []byte
}

type PullRequestReviewComment struct {
	gorm.Model

	EventDbId     string `gorm:"unique_index"`
	PullRequestId string `sql:"index"`
	RepoName      string `sql:"index"`
	RepoUrl       string
	PRUrl         string
	PRNumber      int `sql:"index"`

	ReviewId int `sql:"index"`

	CommentCreatedAt  time.Time
	CommentUpdatedAt  time.Time
	CommentAuthorName string `sql:"index"`
	CommentAuthorType string `sql:"index"`

	Body           string
	EventTimestamp time.Time `sql:"index"`
	EventAction    string    `sql:"index"`

	RawPayload []byte
}

type PullRequest struct {
	gorm.Model

	EventDbId string `gorm:"unique_index"`

	PullRequestId string `sql:"index"`

	RepoName string `sql:"index"`
	RepoUrl  string
	PRUrl    string
	PRNumber int `sql:"index"`

	State         string `sql:"index"`
	PRAuthorLogin string `sql:"index"`
	PRAuthorType  string `sql:"index"`

	PullRequestCreatedAt time.Time `sql:"index"`
	PullRequestUpdatedAt *time.Time
	PullRequestClosedAt  *time.Time
	PullRequestMergedAt  *time.Time `sql:"index"`

	EventInitiatorLogin       string `sql:"index"`
	EventInitiatorDisplayName string `sql:"index"`

	Comments     int
	Commits      int
	Additions    int
	Deletions    int
	FilesChanged int

	EventTimestamp time.Time `sql:"index"`
	EventAction    string    `sql:"index"`

	RawPayload []byte
}
