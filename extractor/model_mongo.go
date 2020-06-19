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

type PRCommentEvent struct {
	ID        string                `bson:"id"`
	Type      string                `bson:"type"`
	Actor     EventActor            `bson:"actor"`
	Repo      RepoInfo              `bson:"repo"`
	Public    bool                  `bson:"public"`
	CreatedAt time.Time             `bson:"created_at"`
	OrgInfo   *OrgInfo              `bson:"org"`
	Payload   PRCommentEventPayload `bson:"payload"`
}

type PRCommentEventPayload struct {
	Action  string     `bson:"action"`
	Comment ApiComment `bson:"comment"`
	Issue   ApiIssue   `bson:"issue"`
}

type ApiIssue struct {
	URL         string          `bson:"url"`
	HtmlURL     string          `bson:"html_url"`
	ID          int             `bson:"id"`
	NodeID      string          `bson:"node_id"`
	Number      int             `bson:"number"`
	Title       string          `bson:"title"`
	User        User            `bson:"user"`
	State       string          `bson:"state"`
	Comments    int             `bson:"comments"`
	CreatedAt   time.Time       `bson:"created_at"`
	UpdatedAt   time.Time       `bson:"updated_at"`
	ClosedAt    *time.Time      `bson:"closed_at"`
	PullRequest *ApiPullRequest `bson:"pull_request"`
}

type ApiComment struct {
	URL               string    `bson:"url"`
	ID                int       `bson:"id"`
	NodeID            string    `bson:"node_id"`
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	User              User      `bson:"user"`
	AuthorAssociation string    `bson:"author_association"`
	Body              string    `bson:"body"`
}

type PRReviewCommentEvent struct {
	ID        string                      `bson:"id"`
	Type      string                      `bson:"type"`
	Actor     EventActor                  `bson:"actor"`
	Repo      RepoInfo                    `bson:"repo"`
	Public    bool                        `bson:"public"`
	CreatedAt time.Time                   `bson:"created_at"`
	OrgInfo   *OrgInfo                    `bson:"org"`
	Payload   PRReviewCommentEventPayload `bson:"payload"`
}

type PRReviewCommentEventPayload struct {
	Action      string                      `bson:"action"`
	PullRequest ApiPullRequest              `bson:"pull_request"`
	Comment     ApiPullRequestReviewComment `bson:"comment"`
}

type ApiPullRequestReviewComment struct {
	ReviewId          int       `bson:"pull_request_review_id"`
	CommentId         int       `bson:"id"`
	Body              string    `bson:"body"`
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	AuthorAssociation string    `bson:"author_association"`
	User              User      `bson:"user"`
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
	Number            int        `bson:"number"`
}

type User struct {
	Login     string `bson:"login"`
	Id        int    `bson:"id"`
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
