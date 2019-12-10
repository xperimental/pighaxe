package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	logLevelStr  string
	host         string
	organization string
	httpTimeout  time.Duration
)

func main() {
	pflag.StringVarP(&logLevelStr, "log-level", "v", "info", "Logging level.")
	pflag.StringVarP(&host, "host", "h", "github.com", "GitHub host to use.")
	pflag.StringVarP(&organization, "organization", "o", "", "Limit search to certain organization.")
	pflag.DurationVar(&httpTimeout, "http-timeout", 5*time.Second, "Timeout for HTTP Requests.")
	pflag.Parse()

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can not parse log level %q: %s", logLevelStr, err)
		os.Exit(1)
	}

	log := &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
		},
		Level: logLevel,
	}

	args := pflag.Args()
	if len(args) == 0 {
		log.Fatal("no patterns passed")
	}

	patternStr := strings.Join(args, "")
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		log.Fatalf("Can not parse %q: %s", patternStr, err)
	}
	log.Debugf("Pattern: %s", pattern)

	ctx := context.TODO()
	client, err := getClient(ctx, host)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}

	repositories, _, err := client.Repositories.List(ctx, organization, nil)
	if err != nil {
		log.Fatalf("Error listing repositories: %s", err)
	}

	for _, repo := range repositories {
		log.Debugf("Repo: %s", repo.GetCloneURL())
	}
}
