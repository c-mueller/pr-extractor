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
)

var (
	configPath = kingpin.Arg("config-path", "Path to the configuration").ExistingFile()
)

func init() {
	kingpin.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
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

	e := &extractor.Exctractor{Config: cfg}
	err = e.Run()
	if err != nil {
		logrus.WithError(err).Fatal("Failed during extractor execution")
	}

}
