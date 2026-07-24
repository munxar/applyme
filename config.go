package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

//go:embed config.schema.json
var configSchemaJSON []byte

// Config holds applyme's settings. Its zero value has empty defaults;
// loadConfig fills it in from config.json and command line flags.
type Config struct {
	JobAdAPIBaseURL       string `json:"jobAdApiBaseUrl"`
	ApplicationsDir       string `json:"applicationsDir"`
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds"`
	CVFileName            string `json:"cvFileName"`
	CoverFileName         string `json:"coverFileName"`
}

// defaultCVFileName and defaultCoverFileName are used whenever
// CVFileName/CoverFileName are left empty, e.g. because config.json
// predates these settings.
const (
	defaultCVFileName    = "cv.pdf"
	defaultCoverFileName = "cover.pdf"
)

// configTemplate holds the sane defaults `applyme init` writes into a
// fresh config.json.
var configTemplate = Config{
	JobAdAPIBaseURL:       "https://www.job-room.ch/jobadservice/api/jobAdvertisements",
	ApplicationsDir:       "applications",
	RequestTimeoutSeconds: 10,
	CVFileName:            defaultCVFileName,
	CoverFileName:         defaultCoverFileName,
}

// loadConfig builds the effective Config for a command: start from the zero
// value, layer in config.json if present, then bind command line flags on
// fs so they take precedence, and parse args. fs must not be parsed yet -
// callers register their own flags on it before calling loadConfig.
func loadConfig(fs *flag.FlagSet, args []string) (Config, error) {
	cfg, err := readConfigFile()
	if err != nil {
		return Config{}, err
	}

	bindConfigFlags(fs, &cfg)
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func readConfigFile() (Config, error) {
	var cfg Config
	raw, err := os.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading config.json: %w", err)
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config.json: %w", err)
	}
	return cfg, nil
}

func bindConfigFlags(fs *flag.FlagSet, cfg *Config) {
	fs.StringVar(&cfg.JobAdAPIBaseURL, "api-base-url", cfg.JobAdAPIBaseURL, "job-room.ch job advertisement API base url")
	fs.StringVar(&cfg.ApplicationsDir, "applications-dir", cfg.ApplicationsDir, "folder applications are stored under")
	fs.IntVar(&cfg.RequestTimeoutSeconds, "request-timeout", cfg.RequestTimeoutSeconds, "http request timeout in seconds")
	fs.StringVar(&cfg.CVFileName, "cv-filename", cfg.CVFileName, "go template for the generated cv pdf's file name, executed with the Application as data")
	fs.StringVar(&cfg.CoverFileName, "cover-filename", cfg.CoverFileName, "go template for the generated cover letter pdf's file name, executed with the Application as data")
}
