package main

import (
	"context"
	"fmt"
	"strings"

	hub "github.com/github/hub/github"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func getClient(ctx context.Context, host string) (*github.Client, transport.AuthMethod, error) {
	hostConfig := hub.CurrentConfig().Find(host)
	if hostConfig == nil {
		return nil, nil, fmt.Errorf(`can not find authentication for %q. Install "hub" and authenticate.`, host)
	}

	token := &oauth2.Token{
		AccessToken: hostConfig.AccessToken,
	}
	tokenSource := oauth2.StaticTokenSource(token)
	httpClient := oauth2.NewClient(ctx, tokenSource)
	httpClient.Timeout = httpTimeout

	auth := &http.BasicAuth{
		Username: hostConfig.User,
		Password: hostConfig.AccessToken,
	}

	if host == "github.com" {
		return github.NewClient(httpClient), auth, nil
	}

	if !strings.HasPrefix(host, hostConfig.Protocol) {
		host = fmt.Sprintf("%s://%s", hostConfig.Protocol, host)
	}

	baseURL := fmt.Sprintf("%s/api/v3", host)
	client, err := github.NewEnterpriseClient(baseURL, "", httpClient)
	if err != nil {
		return nil, nil, fmt.Errorf("can not create enterprise client: %s", err)
	}

	return client, auth, nil
}
