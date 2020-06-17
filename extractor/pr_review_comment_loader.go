package extractor

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"sync/atomic"
)

func (e *Exctractor) loadPullRequestReviewComments() error {
	logger := e.logger.WithField("module", "pr_comment_fetcher")
	filter := map[string]string{
		"type": "PullRequestReviewCommentEvent",
	}

	count, err := e.mongoDbDatabase.Collection("events").CountDocuments(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to fetch pr comment events")
		return err
	}

	c, err := e.mongoDbDatabase.Collection("events").Find(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to fetch pr comment events")
		return err
	}

	stop := false

	totalCount := int32(0)
	processedSucessfulCount := int32(0)
	processedFailCount := int32(0)

	wg := sync.WaitGroup{}

	workers := make([]chan bson.Raw, 10)
	for i := 0; i < len(workers); i++ {
		chn := make(chan bson.Raw, 100)
		workers[i] = chn
		workerId := i
		wg.Add(1)
		go func() {
			log := e.logger.WithField("worker", fmt.Sprintf("pull_review_comments_%d", workerId))

			for elem := range chn {
				var evt PRReviewCommentEvent
				_ = bson.Unmarshal(elem, &evt)

				err = insertPullRequestReviewComment(evt, e, elem, e.sqlDb)
				if err != nil {
					atomic.AddInt32(&processedFailCount, 1)
					continue
				}

				atomic.AddInt32(&processedSucessfulCount, 1)

				if processedSucessfulCount%1000 == 0 {
					totalProcessed := processedSucessfulCount + processedFailCount
					succPercentage := (float64(processedFailCount) / float64(totalProcessed)) * 100

					percentage := (float64(totalProcessed) / float64(count)) * 100
					log.Infof("Processed %d/%d (%f%%) Items. %f%% Insertions have failed (%d Failed, %d Succeeded)", totalProcessed, count, percentage, succPercentage, processedFailCount, processedSucessfulCount)
				}
			}
			wg.Done()
		}()
	}

	e.stopFuncsPreSleep = append(e.stopFuncsPreSleep, func() {
		stop = true
	})

	e.stopFuncsPostSleep = append(e.stopFuncsPostSleep, func() {
		logger.Info("Stopping!")
		for _, worker := range workers {
			close(worker)
		}
	})

	for c.Next(context.TODO()) {
		if stop {
			break
		}
		elem := c.Current

		workers[int(totalCount)%len(workers)] <- elem

		totalCount++
		if totalCount%100000 == 0 {
			percentage := (float64(totalCount) / float64(count)) * 100
			logger.Infof("Fetched %d/%d (%f%%) Items from MongoDB", totalCount, count, percentage)
		}
	}

	wg.Wait()
	return nil
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
