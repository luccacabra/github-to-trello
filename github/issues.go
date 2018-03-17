package github

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubql"
)

type IssuesService service

func (i *IssuesService) Assigned() ([]IssueNode, error) {
	searchQuery := githubql.String(
		fmt.Sprintf(
			"\\\"is:open author:%s org:%s archived:false\\\"",
			i.client.getUserName(),
			i.client.getOrgName(),
		),
	)
	variables := map[string]interface{}{
		"searchQuery": searchQuery,
	}

	var Query struct {
		Search struct {
			Edges []Node
		} `graphql:"search(query: $searchQuery, type: ISSUE, first:2)"`
	}

	err := i.client.githubql.Query(
		context.Background(),
		&Query,
		variables,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying open assigned issues")
	}

	nodes := Query.Search.Edges
	issues := make([]IssueNode, len(nodes))

	for idx, node := range nodes {
		issues[idx] = node.Node
	}

	return issues, nil
}

func (i *IssuesService) Mentioned() ([]*IssueNode, error) {
	return nil, nil
}
