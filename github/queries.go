/* GraphQL query objects */

package github

import (
	"github.com/shurcooL/githubql"
)

type SearchType githubql.String

const (
	ISSUE      SearchType = "ISSUE"
	REPOSITORY            = "REPOSITORY"
	USER                  = "USER"
)

type Search struct {
	After  githubql.String
	Before githubql.String

	First githubql.Int
	Last  githubql.Int

	Query githubql.String
	Type  SearchType
}

type CommentNode struct {
	Node struct {
		Author struct {
			Login githubql.String
		}
		Body githubql.String
		URL  githubql.String
	}
}

type IssueNode struct {
	Issue struct {
		Body     githubql.String
		Comments struct {
			Edges []CommentNode
		} `graphql:"comments(last:100)"`
		CreatedAt  githubql.DateTime
		ID         githubql.String
		Repository struct {
			Name githubql.String
		}
		Title githubql.String
		URL   githubql.String
	} `graphql:"... on Issue"`
}
