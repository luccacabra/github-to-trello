/* GraphQL query objects */

package github

import (
	"github.com/shurcooL/githubql"
)

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
		} `graphql:"comments(last:1)"`
		CreatedAt  githubql.DateTime
		ID         githubql.String
		Repository struct {
			Name githubql.String
		}
		Title githubql.String
		URL   githubql.String
	} `graphql:"... on Issue"`
}

type Node struct {
	Node IssueNode
}
