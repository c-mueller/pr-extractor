package extractor

import (
	"go.mongodb.org/mongo-driver/bson"
	"sync/atomic"
)

func (e *Extractor) loadIssueComments() error {
	loggerMod := "pull_request_issue_comment_fetcher"
	filter := map[string]string{
		"type": "IssueCommentEvent",
	}
	skippedCount := uint64(0)
	return e.runDataFetcher(filter, "events", func(data bson.Raw) error {
		var evt PRCommentEvent
		_ = bson.Unmarshal(data, &evt)

		inserted, err := e.insertIssueComments(evt, data)
		if !inserted {
			atomic.AddUint64(&skippedCount, 1)
			if skippedCount%1000 == 0 {
				e.logger.WithField("module", loggerMod).Infof("Skipped %d Issue Events", skippedCount)
			}
		}
		return err
	}, loggerMod)
}

func (e *Extractor) insertIssueComments(evt PRCommentEvent, elem bson.Raw) (bool, error) {
	tx := e.sqlDb.Begin()
	defer tx.RollbackUnlessCommitted()

	prId := getPullRequestId(evt)

	var count int64

	err := tx.Model(&PullRequestComment{}).Where(&PullRequestComment{
		PullRequestId: prId,
	}).Count(&count).Error
	if err != nil {
		return false, err
	} else if count == 0 {
		return false, nil
	}

	eventId := getEventId(evt)

	var comp []byte

	if e.Config.IncludeRaw {
		comp, _ = GzipCompress(elem)
	}

	revievCommentEvent := PullRequestComment{
		EventDbId:         eventId,
		PullRequestId:     prId,
		RepoName:          evt.Repo.Name,
		RepoUrl:           evt.Repo.Url,
		PRUrl:             evt.Payload.Issue.HtmlURL,
		PRNumber:          evt.Payload.Issue.Number,
		CommentCreatedAt:  evt.Payload.Comment.CreatedAt,
		CommentUpdatedAt:  evt.Payload.Comment.UpdatedAt,
		CommentAuthorName: evt.Payload.Comment.User.Login,
		CommentAuthorType: evt.Payload.Comment.User.Type,
		Body:              evt.Payload.Comment.Body,
		RawPayload:        comp,
	}

	err = tx.Save(&revievCommentEvent).Error
	if err != nil {
		return false, err
	}

	return true, tx.Commit().Error
}
