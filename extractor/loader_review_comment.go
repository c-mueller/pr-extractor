package extractor

import (
	"go.mongodb.org/mongo-driver/bson"
)

func (e *Extractor) loadPullRequestReviewComments() error {
	filter := map[string]string{
		"type": "PullRequestReviewCommentEvent",
	}
	return e.runDataFetcher(filter, "events", func(data bson.Raw) error {
		var evt PRReviewCommentEvent
		_ = bson.Unmarshal(data, &evt)

		return e.insertPullRequestReviewComment(evt, data)
	}, "pull_request_review_comment_fetcher")
}

func (e *Extractor) insertPullRequestReviewComment(evt PRReviewCommentEvent, elem bson.Raw) error {
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
		EventTimestamp:    evt.CreatedAt,
		EventAction:       evt.Payload.Action,
		RawPayload:        comp,
	}

	return e.sqlDb.Save(&revievCommentEvent).Error
}
