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
)

const GithubDbName = "github"

type Exctractor struct {
	Config config.Config

	sqlDb           *gorm.DB
	mongoDb         *mongo.Client
	mongoDbDatabase *mongo.Database
	logger          *logrus.Entry
}

func (e *Exctractor) Run() error {
	e.logger = logrus.WithField("module", "extractor")

	e.logger.Debug("Opening connection to the SQL DB")
	sdb, err := gorm.Open(e.Config.DbDriver, e.Config.DbUrl)
	if err != nil {
		e.logger.Error("Opening SQL db failed")
		return err
	}
	e.sqlDb = sdb

	e.logger.Debug("Applying migrations to SQL DB")

	// TODO ADD MIGRATIONS
	err = e.sqlDb.AutoMigrate(&SqlPr{}).Error
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

	c, err := e.mongoDbDatabase.Collection("events").Find(context.TODO(), map[string]string{
		"type": "PullRequestEvent",
	})
	if err != nil {
		e.logger.Error("failed to fetch events")
		return err
	}
	processedCount := 0
	tx := e.sqlDb.Begin()

	for c.Next(context.TODO()) {
		elem := c.Current

		var evt PullRequestEvent
		_ = bson.Unmarshal(elem, &evt)

		str := fmt.Sprintf("%s-%s", evt.CreatedAt.String(), evt.Payload.PullRequest.URL)
		h := sha256.New()
		h.Write([]byte(str))
		resultHash := hex.EncodeToString(h.Sum([]byte{}))

		var comp []byte

		if e.Config.IncludeRaw {
			comp, _ = GzipCompress(elem)
		}

		resultEvt := SqlPr{
			EventDbId:                 resultHash,
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

		err = tx.Save(&resultEvt).Error
		if err != nil {
			continue
		}

		processedCount++

		if processedCount%1000 == 0 {
			tx.Commit()
			tx = e.sqlDb.Begin()
			e.logger.Infof("Total Items Processed %d", processedCount)
		}
	}
	tx.Commit()

	return nil
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
