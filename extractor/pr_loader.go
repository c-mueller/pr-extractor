package extractor

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"sync/atomic"
)

func (e *Exctractor) loadPullRequests() error {
	logger := e.logger.WithField("module","pr_loader")
	filter := map[string]string{
		"type": "PullRequestEvent",
	}

	count, err := e.mongoDbDatabase.Collection("events").CountDocuments(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to fetch PR events")
		return err
	}

	c, err := e.mongoDbDatabase.Collection("events").Find(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to fetch PR events")
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
			log := e.logger.WithField("worker", fmt.Sprintf("pulls_%d", workerId))

			for elem := range chn {
				var evt PullRequestEvent
				_ = bson.Unmarshal(elem, &evt)

				err = insertPullRequest(evt, e, elem, e.sqlDb)
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
		e.logger.Info("Stopping!")
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

func insertPullRequest(evt PullRequestEvent, e *Exctractor, elem bson.Raw, tx *gorm.DB) error {
	eventId := getEventId(evt)

	prId := getPullRequestId(evt)

	var comp []byte

	if e.Config.IncludeRaw {
		comp, _ = GzipCompress(elem)
	}

	resultEvt := PullRequest{
		EventDbId:                 eventId,
		PullRequestId:             prId,
		RepoName:                  evt.Repo.Name,
		RepoUrl:                   evt.Repo.Url,
		PRUrl:                     evt.Payload.PullRequest.URL,
		PRNumber:                  evt.Payload.Number,
		State:                     evt.Payload.PullRequest.State,
		PRAuthorLogin:             evt.Payload.PullRequest.User.Login,
		PRAuthorType:              evt.Payload.PullRequest.User.Type,
		PullRequestCreatedAt:      evt.Payload.PullRequest.CreatedAt,
		PullRequestUpdatedAt:      evt.Payload.PullRequest.UpdatedAt,
		PullRequestClosedAt:       evt.Payload.PullRequest.ClosedAt,
		PullRequestMergedAt:       evt.Payload.PullRequest.MergedAt,
		EventInitiatorLogin:       evt.Actor.Login,
		EventInitiatorDisplayName: evt.Actor.DisplayLogin,
		Comments:                  evt.Payload.PullRequest.Comments,
		Commits:                   evt.Payload.PullRequest.Commits,
		Additions:                 evt.Payload.PullRequest.Additions,
		Deletions:                 evt.Payload.PullRequest.Deletions,
		FilesChanged:              evt.Payload.PullRequest.FilesChanged,
		EventTimestamp:            evt.CreatedAt,
		EventAction:               evt.Payload.Action,
		RawPayload:                comp,
	}
	return tx.Save(&resultEvt).Error
}
