package extractor

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/c-mueller/pr-extractor/config"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"sync"
	"time"
)

const GithubDbName = "github"

type Extractor struct {
	Config config.Config

	sqlDb              *gorm.DB
	mongoDb            *mongo.Client
	mongoDbDatabase    *mongo.Database
	logger             *logrus.Entry
	stopFuncsPostSleep []func()
	stopFuncsPreSleep  []func()
}

func (e *Extractor) RunFull() error {
	err := e.init(true)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		err := e.loadPullRequests()
		if err != nil {
			e.logger.WithError(err).Fatalf("Failed to load Pull Requests...")
		}
	}()

	go func() {
		defer wg.Done()
		err := e.loadPullRequestReviewComments()
		if err != nil {
			e.logger.WithError(err).Fatalf("Failed to load Pull Request review comments...")
		}
	}()

	go func() {
		defer wg.Done()
		err := e.loadIssueComments()
		if err != nil {
			e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
		}
	}()

	wg.Wait()

	e.stopFuncsPostSleep = make([]func(), 0)
	e.stopFuncsPreSleep = make([]func(), 0)

	err = e.loadPullRequestReviewComments()
	if err != nil {
		e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
	}

	return nil
}

func (e *Extractor) RunIssueComments() error {
	err := e.init(true)
	if err != nil {
		return err
	}

	err = e.loadIssueComments()
	if err != nil {
		e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
	}

	return nil
}

func (e *Extractor) RunPullRequests() error {
	err := e.init(true)
	if err != nil {
		return err
	}

	err = e.loadPullRequests()
	if err != nil {
		e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
	}

	return nil
}

func (e *Extractor) RunReviewComments() error {
	err := e.init(true)
	if err != nil {
		return err
	}

	err = e.loadPullRequests()
	if err != nil {
		e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
	}

	return nil
}

func (e *Extractor) LoadFromStdin() error {
	err := e.init(false)
	if err != nil {
		return err
	}

	totalEventStats := make(map[string]uint)
	failEventStats := make(map[string]uint)
	scn := bufio.NewScanner(os.Stdin)
	for scn.Scan() {
		payload := scn.Bytes()
		var evt Event
		err = json.Unmarshal(payload, &evt)
		if err != nil {
			e.logger.WithError(err).Warn("Failed to read event. Skipping.")
			continue
		}
		totalEventStats[evt.Type]++

		switch evt.Type {
		case "PullRequestEvent":
			var prEvent PullRequestEvent
			err = json.Unmarshal(payload, &prEvent)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			err = e.insertPullRequest(prEvent, payload)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			break
		case "IssueCommentEvent":
			var commentEvent PRCommentEvent
			err = json.Unmarshal(payload, &commentEvent)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			_, err = e.insertIssueComments(commentEvent, payload)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			break
		case "PullRequestReviewCommentEvent":
			var reviewCommentEvent PRReviewCommentEvent
			err = json.Unmarshal(payload, &reviewCommentEvent)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			err = e.insertPullRequestReviewComment(reviewCommentEvent, payload)
			if err != nil {
				failEventStats[evt.Type]++
				continue
			}
			break
		default:
			break
		}
	}

	e.logger.Infof("Done Loading. Loaded the following events")
	for s, u := range totalEventStats {
		e.logger.Infof("%s: total=%d fail=%d", s, u, failEventStats[s])
	}
	return nil
}

func (e *Extractor) init(useMongo bool) error {
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
	err = e.sqlDb.AutoMigrate(&PullRequestComment{}).Error
	if err != nil {
		return err
	}
	err = e.sqlDb.AutoMigrate(&PullRequestReviewComment{}).Error
	if err != nil {
		return err
	}

	if useMongo {
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

	}

	e.initInterruptHook()

	return nil
}

func (e *Extractor) initInterruptHook() {
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
			if e.mongoDbDatabase != nil {
				e.mongoDbDatabase.Client().Disconnect(context.TODO())
			}
			e.sqlDb.Close()
		}
	}()
}
