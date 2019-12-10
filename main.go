package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/memory"
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

	patternStr := strings.Join(args, " ")
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		log.Fatalf("Can not parse %q: %s", patternStr, err)
	}
	log.Infof("Pattern: %s", pattern)

	ctx := context.TODO()
	client, repoAuth, err := getClient(ctx, host)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}

	repositories, _, err := client.Repositories.List(ctx, organization, nil)
	if err != nil {
		log.Fatalf("Error listing repositories: %s", err)
	}

	header := []string{"repo", "file"}
	if pattern.NumSubexp() == 0 {
		header = append(header, "line")
	} else {
		for i := 0; i < pattern.NumSubexp(); i++ {
			header = append(header, fmt.Sprintf("group%d", i))
		}
	}

	output := csv.NewWriter(os.Stdout)
	output.Write(header)
	output.Flush()

	for _, repo := range repositories {
		repoURL := repo.GetCloneURL()
		log.Infof("Searching: %s", repoURL)
		if err := findInRepo(log.WithField("repo", repoURL), writeOutput(output, repoURL), repoURL, repoAuth, pattern); err != nil {
			log.Warnf("Error in %q: %s", repoURL, err)
		}
	}
}

type outputFunc func(fileName, line string, match []string)

func writeOutput(output *csv.Writer, repoURL string) outputFunc {
	return func(fileName, line string, match []string) {
		record := []string{repoURL, fileName}
		if len(match) == 0 {
			record = append(record, line)
		} else {
			record = append(record, match[1:]...)
		}
		output.Write(record)
		output.Flush()
	}
}

func findInRepo(log logrus.FieldLogger, output outputFunc, repoURL string, auth transport.AuthMethod, pattern *regexp.Regexp) error {
	fs := memfs.New()
	storer := memory.NewStorage()

	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:   repoURL,
		Auth:  auth,
		Depth: 1,
	})
	if err != nil {
		return fmt.Errorf("can not clone: %s", err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("can not get worktree: %s", err)
	}

	dir, err := fs.ReadDir("")
	if err != nil {
		return fmt.Errorf("can not read root: %s", err)
	}

	return walkInfo(log, output, tree.Filesystem, "", dir, pattern)
}

func walkInfo(log logrus.FieldLogger, output outputFunc, fs billy.Filesystem, base string, dir []os.FileInfo, pattern *regexp.Regexp) error {
	for _, info := range dir {
		name := filepath.Join(base, info.Name())
		if info.IsDir() {
			if err := findInDir(log.WithField("dir", name), output, fs, name, pattern); err != nil {
				log.Warnf("error finding in %q: %s", name, err)
			}
			continue
		}

		if err := findInFile(log.WithField("file", name), output, fs, name, pattern); err != nil {
			log.Warnf("error finding in %q: %s", name, err)
		}
	}

	return nil
}

func findInDir(log logrus.FieldLogger, output outputFunc, fs billy.Filesystem, dir string, pattern *regexp.Regexp) error {
	log.Debugf("dir: %s", dir)

	subDir, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("can not list %q: %s", dir, err)
	}

	return walkInfo(log, output, fs, dir, subDir, pattern)
}

func findInFile(log logrus.FieldLogger, output outputFunc, fs billy.Filesystem, name string, pattern *regexp.Regexp) error {
	log.Debugf("file: %s", name)
	file, err := fs.Open(name)
	if err != nil {
		return fmt.Errorf("can not read %q: %s", name, err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if match := pattern.FindStringSubmatch(line); match != nil {
			output(name, line, match)
		}
	}
	return nil
}
