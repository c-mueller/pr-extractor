package main

import (
	"github.com/c-mueller/pr-extractor/config"
	"github.com/c-mueller/pr-extractor/extractor"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var (
	configPath = kingpin.Flag("config-path", "Path to the configuration").Short('c').Default("config.yml").ExistingFile()
)

func init() {
	kingpin.Command("full", "Load all events from mongoDB")
	kingpin.Command("comments", "Load pr comment events from mongoDB")
	kingpin.Command("pull-requests", "Load pr events from mongoDB")
	kingpin.Command("review-comments", "Load pr review comment events from mongoDB")
	kingpin.Command("json", "load data from json (stdin) to import gharchive dumps")

	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	cmd := kingpin.Parse()
	contents, err := ioutil.ReadFile(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
	}
	var cfg config.Config
	err = yaml.Unmarshal(contents, &cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to unmarshal config")
	}
	if cfg.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	switch cmd {
	case "json":
		e := &extractor.Extractor{Config: cfg}
		err = e.LoadFromStdin()
		if err != nil {
			logrus.WithError(err).Fatal("Failed during extractor execution")
		}
		break
	case "full":
		e := &extractor.Extractor{Config: cfg}
		err = e.RunFull()
		if err != nil {
			logrus.WithError(err).Fatal("Failed during extractor execution")
		}
		break
	case "comments":
		e := &extractor.Extractor{Config: cfg}
		err = e.RunIssueComments()
		if err != nil {
			logrus.WithError(err).Fatal("Failed during extractor execution")
		}
		break
	case "pull-requests":
		e := &extractor.Extractor{Config: cfg}
		err = e.RunPullRequests()
		if err != nil {
			logrus.WithError(err).Fatal("Failed during extractor execution")
		}
		break
	case "review-comments":
		e := &extractor.Extractor{Config: cfg}
		err = e.RunReviewComments()
		if err != nil {
			logrus.WithError(err).Fatal("Failed during extractor execution")
		}
		break
	default:
		os.Exit(1)
	}
}
