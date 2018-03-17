/* Thin wrapper around github.com/shurcooL/githubql */

package github

import (
	"context"

	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

type Config struct {
	OrgName  string
	UserName string
}

type Client struct {
	githubql *githubql.Client

	orgName  string
	userName string

	common service

	Issues       *IssuesService
	PullRequests *PullRequestService
}

type service struct {
	client *Client
}

func NewClient(token string, config Config) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	c := &Client{
		githubql: githubql.NewClient(httpClient),
		orgName:  config.OrgName,
		userName: config.UserName,
	}

	c.common.client = c
	c.Issues = (*IssuesService)(&c.common)
	c.PullRequests = (*PullRequestService)(&c.common)

	return c
}

func (c *Client) getOrgName() string {
	return c.orgName
}

func (c *Client) getUserName() string {
	return c.userName
}
