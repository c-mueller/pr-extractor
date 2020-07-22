package extractor

import "time"

type Event struct {
	ID        string     `bson:"id" json:"id"`
	Type      string     `bson:"type" json:"type"`
	Actor     EventActor `bson:"actor" json:"actor"`
	Repo      RepoInfo   `bson:"repo" json:"repo"`
	Public    bool       `bson:"public" json:"public"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	OrgInfo   *OrgInfo   `bson:"org" json:"org_info"`
}

type PRCommentEvent struct {
	ID        string                `bson:"id" json:"id"`
	Type      string                `bson:"type" json:"type"`
	Actor     EventActor            `bson:"actor" json:"actor"`
	Repo      RepoInfo              `bson:"repo" json:"repo"`
	Public    bool                  `bson:"public" json:"public"`
	CreatedAt time.Time             `bson:"created_at" json:"created_at"`
	OrgInfo   *OrgInfo              `bson:"org" json:"org_info"`
	Payload   PRCommentEventPayload `bson:"payload" json:"payload"`
}

type PRCommentEventPayload struct {
	Action  string     `bson:"action" json:"action"`
	Comment ApiComment `bson:"comment" json:"comment"`
	Issue   ApiIssue   `bson:"issue" json:"issue"`
}

type ApiIssue struct {
	URL         string          `bson:"url" json:"url"`
	HtmlURL     string          `bson:"html_url" json:"html_url"`
	ID          int             `bson:"id" json:"id"`
	NodeID      string          `bson:"node_id" json:"node_id"`
	Number      int             `bson:"number" json:"number"`
	Title       string          `bson:"title" json:"title"`
	User        User            `bson:"user" json:"user"`
	State       string          `bson:"state" json:"state"`
	Comments    int             `bson:"comments" json:"comments"`
	CreatedAt   time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `bson:"updated_at" json:"updated_at"`
	ClosedAt    *time.Time      `bson:"closed_at" json:"closed_at"`
	PullRequest *ApiPullRequest `bson:"pull_request" json:"pull_request"`
}

type ApiComment struct {
	URL               string    `bson:"url" json:"url"`
	ID                int       `bson:"id" json:"id"`
	NodeID            string    `bson:"node_id" json:"node_id"`
	CreatedAt         time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updated_at"`
	User              User      `bson:"user" json:"user"`
	AuthorAssociation string    `bson:"author_association" json:"author_association"`
	Body              string    `bson:"body" json:"body"`
}

type PRReviewCommentEvent struct {
	ID        string                      `bson:"id" json:"id"`
	Type      string                      `bson:"type" json:"type"`
	Actor     EventActor                  `bson:"actor" json:"actor"`
	Repo      RepoInfo                    `bson:"repo" json:"repo"`
	Public    bool                        `bson:"public" json:"public"`
	CreatedAt time.Time                   `bson:"created_at" json:"created_at"`
	OrgInfo   *OrgInfo                    `bson:"org" json:"org_info"`
	Payload   PRReviewCommentEventPayload `bson:"payload" json:"payload"`
}

type PRReviewCommentEventPayload struct {
	Action      string                      `bson:"action" json:"action"`
	PullRequest ApiPullRequest              `bson:"pull_request" json:"pull_request"`
	Comment     ApiPullRequestReviewComment `bson:"comment" json:"comment"`
}

type ApiPullRequestReviewComment struct {
	ReviewId          int       `bson:"pull_request_review_id" json:"review_id"`
	CommentId         int       `bson:"id" json:"comment_id"`
	Body              string    `bson:"body" json:"body"`
	CreatedAt         time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updated_at"`
	AuthorAssociation string    `bson:"author_association" json:"author_association"`
	User              User      `bson:"user" json:"user"`
}

type OrgInfo struct {
	ID    string `bson:"id" json:"id"`
	Login string `bson:"login" json:"login"`
	URL   string `bson:"url" json:"url"`
}

type PullRequestEvent struct {
	ID        string                  `bson:"id" json:"id"`
	Type      string                  `bson:"type" json:"type"`
	Actor     EventActor              `bson:"actor" json:"actor"`
	Repo      RepoInfo                `bson:"repo" json:"repo"`
	Public    bool                    `bson:"public" json:"public"`
	CreatedAt time.Time               `bson:"created_at" json:"created_at"`
	OrgInfo   *OrgInfo                `bson:"org" json:"org_info"`
	Payload   PullRequestEventPayload `bson:"payload" json:"payload"`
}

type PullRequestEventPayload struct {
	Action      string         `bson:"action" json:"action"`
	Number      int            `bson:"number" json:"number"`
	PullRequest ApiPullRequest `bson:"pull_request" json:"pull_request"`
}

type ApiPullRequest struct {
	URL               string     `bson:"url" json:"url"`
	NodeID            string     `bson:"node_id" json:"node_id"`
	State             string     `bson:"state" json:"state"`
	Locked            bool       `bson:"locked" json:"locked"`
	Title             string     `bson:"title" json:"title"`
	User              User       `bson:"user" json:"user"`
	Body              string     `bson:"body" json:"body"`
	CreatedAt         time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt         *time.Time `bson:"updated_at" json:"updated_at"`
	ClosedAt          *time.Time `bson:"closed_at" json:"closed_at"`
	MergedAt          *time.Time `bson:"merged_at" json:"merged_at"`
	Merged            bool       `bson:"merged" json:"merged"`
	MergedBy          *User      `bson:"merged_by" json:"merged_by"`
	MergeCommitSha    string     `bson:"merge_commit_sha" json:"merge_commit_sha"`
	AuthorAssociation string     `bson:"author_association" json:"author_association"`
	Comments          int        `bson:"comments" json:"comments"`
	Commits           int        `bson:"commits" json:"commits"`
	Additions         int        `bson:"additions" json:"additions"`
	Deletions         int        `bson:"deletions" json:"deletions"`
	FilesChanged      int        `bson:"files_changed" json:"files_changed"`
	Number            int        `bson:"number" json:"number"`
}

type User struct {
	Login     string `bson:"login" json:"login"`
	Id        int    `bson:"id" json:"id"`
	NodeId    string `bson:"node_id" json:"node_id"`
	HTMLUrl   string `bson:"html_url" json:"html_url"`
	Type      string `bson:"type" json:"type"`
	SiteAdmin bool   `bson:"site_admin" json:"site_admin"`
}

type RepoInfo struct {
	Name string `bson:"name" json:"name"`
	Url  string `bson:"url" json:"url"`
}

type EventActor struct {
	Login        string `bson:"login" json:"login"`
	DisplayLogin string `bson:"display_login" json:"display_login"`
	URL          string `bson:"url" json:"url"`
}
