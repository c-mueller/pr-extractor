package extractor

import "time"

type Event struct {
	ID        string     `bson:"id"`
	Type      string     `bson:"type"`
	Actor     EventActor `bson:"actor"`
	Repo      RepoInfo   `bson:"repo"`
	Public    bool       `bson:"public"`
	CreatedAt time.Time  `bson:"created_at"`
	OrgInfo   *OrgInfo   `bson:"org"`
}

type OrgInfo struct {
	ID    string `bson:"id"`
	Login string `bson:"login"`
	URL   string `bson:"url"`
}

type PullRequestEvent struct {
	ID        string                  `bson:"id"`
	Type      string                  `bson:"type"`
	Actor     EventActor              `bson:"actor"`
	Repo      RepoInfo                `bson:"repo"`
	Public    bool                    `bson:"public"`
	CreatedAt time.Time               `bson:"created_at"`
	OrgInfo   *OrgInfo                `bson:"org"`
	Payload   PullRequestEventPayload `bson:"payload"`
}

type PullRequestEventPayload struct {
	Action      string         `bson:"action"`
	Number      int            `bson:"number"`
	PullRequest ApiPullRequest `bson:"pull_request"`
}

type ApiPullRequest struct {
	URL               string     `bson:"url"`
	NodeID            string     `bson:"node_id"`
	State             string     `bson:"state"`
	Locked            bool       `bson:"locked"`
	Title             string     `bson:"title"`
	User              User       `bson:"user"`
	Body              string     `bson:"body"`
	CreatedAt         time.Time  `bson:"created_at"`
	UpdatedAt         *time.Time `bson:"updated_at"`
	ClosedAt          *time.Time `bson:"closed_at"`
	MergedAt          *time.Time `bson:"merged_at"`
	Merged            bool       `bson:"merged"`
	MergedBy          *User      `bson:"merged_by"`
	MergeCommitSha    string     `bson:"merge_commit_sha"`
	AuthorAssociation string     `bson:"author_association"`
	Comments          int        `bson:"comments"`
	Commits           int        `bson:"commits"`
	Additions         int        `bson:"additions"`
	Deletions         int        `bson:"deletions"`
	FilesChanged      int        `bson:"files_changed"`
}

type User struct {
	Login     string `bson:"login"`
	NodeId    string `bson:"node_id"`
	HTMLUrl   string `bson:"html_url"`
	Type      string `bson:"type"`
	SiteAdmin bool   `bson:"site_admin"`
}

type RepoInfo struct {
	Name string `bson:"name"`
	Url  string `bson:"url"`
}

type EventActor struct {
	Login        string `bson:"login"`
	DisplayLogin string `bson:"display_login"`
	URL          string `bson:"url"`
}
