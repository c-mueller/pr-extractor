package extractor

import (
	"context"
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
	err := e.init()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

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
	err := e.init()
	if err != nil {
		return err
	}

	err = e.loadPullRequestReviewComments()
	if err != nil {
		e.logger.WithError(err).Fatalf("Failed to load Pull Request Issue comments...")
	}

	return nil
}

func (e *Extractor) init() error {
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
			e.mongoDbDatabase.Client().Disconnect(context.TODO())
			e.sqlDb.Close()
		}
	}()
}
