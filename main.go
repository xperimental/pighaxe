package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	logLevelStr string
)

func main() {
	pflag.StringVarP(&logLevelStr, "log-level", "v", "info", "")
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
}
