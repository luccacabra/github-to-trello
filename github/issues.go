package github

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubql"
)

type IssuesService service

func (i *IssuesService) Assigned() ([]IssueNode, error) {
	query := githubql.String(
		fmt.Sprintf(
			"is:open assignee:%s org:%s archived:false",
			i.client.getUserName(),
			i.client.getOrgName(),
		),
	)
	issues, err := i.searchIssue(
		Search{
			Query: query,
			First: 100,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying open assigned issues")
	}
	return issues, nil
}

func (i *IssuesService) Mentioned() ([]IssueNode, error) {
	query := githubql.String(
		fmt.Sprintf(
			"is:open mentions:%s -author:%s org:%s archived:false",
			i.client.getUserName(),
			i.client.getUserName(),
			i.client.getOrgName(),
		),
	)
	issues, err := i.searchIssue(
		Search{
			Query: query,
			First: 100,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying open assigned issues")
	}
	return issues, nil
}

func (i *IssuesService) IsClosed(issueName, issueId string) (bool, error) {
	query := githubql.String(
		fmt.Sprintf(
			"is:closed archived:false org %s \"%s\"",
			i.client.getOrgName(),
			issueName,
		),
	)
	issues, err := i.searchIssue(
		Search{
			Query: query,
			First: 1,
		},
	)
	if err != nil {
		return false, errors.Wrap(err, "Error querying closed assigned issues")
	}
	for _, issue := range issues {
		if string(issue.Issue.ID) == issueId {
			return true, nil
		}
	}
	return false, nil
}

func (i *IssuesService) searchIssue(search Search) ([]IssueNode, error) {
	search.Type = ISSUE
	i.client.prepareSearchQuery(&search)

	variables := map[string]interface{}{
		"searchQuery": search.Query,
		"type":        search.Type,
		"first":       search.First,
	}

	type Node struct {
		Node IssueNode
	}
	var Query struct {
		Search struct {
			Edges []Node
		} `graphql:"search(query: $searchQuery, type: $type, first: $first)"`
	}

	if err := i.client.githubql.Query(
		context.Background(),
		&Query,
		variables,
	); err != nil {
		return nil, errors.Wrap(err, "Error querying open assigned issues")
	}

	nodes := Query.Search.Edges
	issues := make([]IssueNode, len(nodes))

	for idx, node := range nodes {
		issues[idx] = node.Node
	}

	return issues, nil
}
