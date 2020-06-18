package extractor

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"sync/atomic"
)

func (e *Exctractor) runDataFetcher(filter map[string]string, collectionName string, inserterFunc func(data bson.Raw) error, loggerName string) error {
	logger := e.logger.WithField("module", loggerName)

	logger.Info("Fetching count from mongoDB")
	count, err := e.mongoDbDatabase.Collection(collectionName).CountDocuments(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to count documents")
		return err
	}
	logger.Infof("Going to load %d documents from mongoDB", count)

	c, err := e.mongoDbDatabase.Collection(collectionName).Find(context.TODO(), filter)
	if err != nil {
		logger.Error("failed to fetch documents")
		return err
	}

	stop := false

	totalCount := int32(0)
	processedSucessfulCount := int32(0)
	processedFailCount := int32(0)

	wg := sync.WaitGroup{}

	logger.Infof("Initializing %d Workers", e.Config.WorkerCount)
	workers := make([]chan bson.Raw, e.Config.WorkerCount)
	for i := 0; i < len(workers); i++ {
		chn := make(chan bson.Raw, e.Config.WorkerQueueLength)
		workers[i] = chn
		wg.Add(1)
		go func() {
			for elem := range chn {
				err := inserterFunc(elem)
				if err != nil {
					atomic.AddInt32(&processedFailCount, 1)
				} else {
					atomic.AddInt32(&processedSucessfulCount, 1)
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

	logger.Info("Processing Data...")
	for c.Next(context.TODO()) {
		if stop {
			break
		}
		elem := c.Current

		workers[int(totalCount)%len(workers)] <- elem

		totalCount++
		if totalCount%1000 == 0 {
			totalFetched := processedSucessfulCount + processedFailCount
			failPercentage := (float64(processedFailCount) / float64(totalFetched)) * 100

			insertedpercentage := (float64(totalFetched) / float64(count)) * 100
			percentage := (float64(totalCount) / float64(count)) * 100
			logger.WithFields(logrus.Fields{
				"total_insertions":     totalFetched,
				"failed_insertions":    processedFailCount,
				"sucessful_insertions": processedSucessfulCount,
				"total_fetched":        totalFetched,
				"total_items":          count,
				"fail_percentage":      fmt.Sprintf("%f%%", failPercentage),
				"inserted_percentage":  fmt.Sprintf("%f%%", insertedpercentage),
				"fetched_percentage":   fmt.Sprintf("%f%%", percentage),
			}).Infof("Inserted %d/%d (%f%%) documents.", totalFetched, count, insertedpercentage)
		}
	}

	logger.Info("Done. Closing worker channels")

	for _, worker := range workers {
		close(worker)
	}

	wg.Wait()
	return nil
}
