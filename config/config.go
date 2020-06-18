package config

type Config struct {
	MongoUrl          string `yaml:"mongo_url"`
	IncludeRaw        bool   `yaml:"include_raw"`
	DbDriver          string `yaml:"db_driver"`
	DbUrl             string `yaml:"db_url"`
	Verbose           bool   `yaml:"verbose"`
	WorkerCount       int    `yaml:"worker_count"`
	WorkerQueueLength int    `yaml:"worker_queue_length"`
}
