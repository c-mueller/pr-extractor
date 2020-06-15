package extractor

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/c-mueller/pr-extractor/config"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

const GithubDbName = "github"

type Exctractor struct {
	Config config.Config

	sqlDb              *gorm.DB
	mongoDb            *mongo.Client
	mongoDbDatabase    *mongo.Database
	logger             *logrus.Entry
	stopFuncsPostSleep []func()
	stopFuncsPreSleep []func()
}

func (e *Exctractor) Run() error {
	err := e.init()
	if err != nil {
		return err
	}

	err = e.loadPullRequests()
	if err != nil {
		return err
	}

	return nil
}

func (e *Exctractor) loadPullRequests() error {
	filter := map[string]string{
		"type": "PullRequestEvent",
	}

	count, err := e.mongoDbDatabase.Collection("events").CountDocuments(context.TODO(), filter)
	if err != nil {
		e.logger.Error("failed to fetch events")
		return err
	}

	c, err := e.mongoDbDatabase.Collection("events").Find(context.TODO(), filter)
	if err != nil {
		e.logger.Error("failed to fetch events")
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

				err = insertPullRequest(evt, e, elem, err, e.sqlDb)
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
			e.logger.Infof("Fetched %d/%d (%f%%) Items from MongoDB", totalCount, count, percentage)
		}
	}

	wg.Wait()
	return nil
}

func (e *Exctractor) init() error {
	e.logger = logrus.WithField("module", "extractor")

	e.logger.Debug("Opening connection to the SQL DB")
	sdb, err := gorm.Open(e.Config.DbDriver, e.Config.DbUrl)
	if err != nil {
		e.logger.Error("Opening SQL db failed")
		return err
	}
	sdb.LogMode(false)
	e.sqlDb = sdb

	e.logger.Debug("Applying migrations to SQL DB")

	// TODO ADD MIGRATIONS
	err = e.sqlDb.AutoMigrate(&PullRequest{}).Error
	if err != nil {
		return err
	}

	e.logger.Debug("Connnecting to MongoDB")

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(e.Config.MongoUrl))
	if err != nil {
		e.logger.Error("Opening MongoDB Connection failed")
		return err
	}
	e.mongoDb = mongoClient
	if err = e.mongoDb.Connect(context.TODO()); err != nil {
		e.logger.Error("connection to mongodb failed")
		return err
	}

	e.mongoDbDatabase = e.mongoDb.Database(GithubDbName)

	e.initInterruptHook()

	return nil
}

func (e *Exctractor) initInterruptHook() {
	e.stopFuncsPostSleep = make([]func(), 0)
	e.stopFuncsPreSleep = make([]func(), 0)
	chn := make(chan os.Signal, 1)
	signal.Notify(chn, os.Interrupt)
	go func() {
		for range chn {
			for _, stopFunc := range e.stopFuncsPreSleep {
				stopFunc()
			}
			time.Sleep(5 * time.Second)
			for _, stopFunc := range e.stopFuncsPostSleep {
				stopFunc()
			}
			e.mongoDbDatabase.Client().Disconnect(context.TODO())
			e.sqlDb.Close()
		}
	}()
}

func insertPullRequest(evt PullRequestEvent, e *Exctractor, elem bson.Raw, err error, tx *gorm.DB) error {
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

func getPullRequestId(evt PullRequestEvent) string {
	prIdInput := fmt.Sprintf("%s#%d", evt.Repo.Name, evt.Payload.Number)
	h := sha256.New()
	h.Write([]byte(prIdInput))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func getEventId(evt PullRequestEvent) string {
	evtIdInput := fmt.Sprintf("%s-%s", evt.CreatedAt.String(), evt.Payload.PullRequest.URL)
	h := sha256.New()
	h.Write([]byte(evtIdInput))
	resultHash := hex.EncodeToString(h.Sum([]byte{}))
	return resultHash
}

func GzipCompress(data []byte) ([]byte, error) {
	var outputBuffer bytes.Buffer
	compressionWriter := gzip.NewWriter(&outputBuffer)
	_, err := compressionWriter.Write(data)
	if err != nil {
		return nil, err
	}
	compressionWriter.Close()

	return outputBuffer.Bytes(), nil
}

func GzipExtract(data []byte) ([]byte, error) {
	inputBuffer := bytes.NewReader(data)
	compressionReader, err := gzip.NewReader(inputBuffer)
	if err != nil {
		return nil, err
	}

	defer compressionReader.Close()

	return ioutil.ReadAll(compressionReader)
}
