package extractor

import (
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
)

func (e *Exctractor) loadPullRequestReviewComments() error {
	filter := map[string]string{
		"type": "PullRequestReviewCommentEvent",
	}
	return e.runDataFetcher(filter, "events", func(data bson.Raw) error {
		var evt PRReviewCommentEvent
		_ = bson.Unmarshal(data, &evt)

		return insertPullRequestReviewComment(evt, e, data, e.sqlDb)
	}, "pull_request_review_comment_fetcher")
}

func insertPullRequestReviewComment(evt PRReviewCommentEvent, e *Exctractor, elem bson.Raw, tx *gorm.DB) error {
	eventId := getEventId(evt)

	prId := getPullRequestId(evt)

	var comp []byte

	if e.Config.IncludeRaw {
		comp, _ = GzipCompress(elem)
	}

	revievCommentEvent := PullRequestReviewComment{
		EventDbId:         eventId,
		PullRequestId:     prId,
		RepoName:          evt.Repo.Name,
		RepoUrl:           evt.Repo.Url,
		PRUrl:             evt.Payload.PullRequest.URL,
		PRNumber:          evt.Payload.PullRequest.Number,
		ReviewId:          evt.Payload.Comment.ReviewId,
		CommentCreatedAt:  evt.Payload.Comment.CreatedAt,
		CommentUpdatedAt:  evt.Payload.Comment.UpdatedAt,
		CommentAuthorName: evt.Payload.Comment.User.Login,
		CommentAuthorType: evt.Payload.Comment.User.Type,
		Body:              evt.Payload.Comment.Body,
		RawPayload:        comp,
	}

	return tx.Save(&revievCommentEvent).Error
}