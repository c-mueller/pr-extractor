package extractor

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"sync/atomic"
)

func (e *Extractor) loadIssueComments() error {
	loggerMod := "pull_request_issue_comment_fetcher"
	filter := map[string]string{
		"type": "IssueCommentEvent",
	}
	skippedCount := uint64(0)
	insertedCount := uint64(0)
	return e.runDataFetcher(filter, "events", func(data bson.Raw) error {
		var evt PRCommentEvent
		_ = bson.Unmarshal(data, &evt)

		inserted, err := e.insertIssueComments(evt, data)
		if !inserted {
			atomic.AddUint64(&skippedCount, 1)
			if skippedCount%10000 == 0 {
				total := skippedCount + insertedCount
				skipPerecentage := (float64(skippedCount) / float64(total)) * 100

				e.logger.WithFields(logrus.Fields{
					"module":             loggerMod,
					"skipped":            skippedCount,
					"inserted":           insertedCount,
					"total":              total,
					"skipped_percentage": fmt.Sprintf("%f%%", skipPerecentage),
				}).Infof("Skipped %d Issue Events", skippedCount)
			}
		} else {
			atomic.AddUint64(&insertedCount, 1)
		}
		return err
	}, loggerMod)
}

func (e *Extractor) insertIssueComments(evt PRCommentEvent, elem bson.Raw) (bool, error) {
	tx := e.sqlDb.Begin()
	defer tx.RollbackUnlessCommitted()

	if evt.Payload.Issue.PullRequest == nil {
		return false, nil
	}

	prId := getPullRequestId(evt)

	//var cnt int64
	//err := tx.Where(&PullRequest{}, "pull_request_id = ?", prId).Count(&cnt).Error
	//if err != nil {
	//	return false, err
	//} else if cnt == 0 {
	//	return false, nil
	//}

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
		PRUrl:             evt.Payload.Issue.PullRequest.URL,
		PRNumber:          evt.Payload.Issue.Number,
		CommentCreatedAt:  evt.Payload.Comment.CreatedAt,
		CommentUpdatedAt:  evt.Payload.Comment.UpdatedAt,
		CommentAuthorName: evt.Payload.Comment.User.Login,
		CommentAuthorType: evt.Payload.Comment.User.Type,
		Body:              evt.Payload.Comment.Body,
		EventTimestamp:    evt.CreatedAt,
		EventAction:       evt.Payload.Action,
		RawPayload:        comp,
	}

	err := tx.Save(&revievCommentEvent).Error
	if err != nil {
		return false, err
	}

	return true, tx.Commit().Error
}
