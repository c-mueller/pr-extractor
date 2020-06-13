package extractor

import (
	"github.com/jinzhu/gorm"
	"time"
)

type SqlPr struct {
	gorm.Model

	EventDbId string `gorm:"unique_index"`

	RepoName string
	RepoUrl  string
	PRUrl    string
	PRNumber int

	State         string
	PRAuthorLogin string
	PRAuthorType  string

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
