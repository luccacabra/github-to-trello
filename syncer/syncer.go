package syncer

import (
	"fmt"
	"github.com/luccacabra/github-to-trello/github"
	"strings"
)

type UserRelationship int

const (
	ASSIGNEE UserRelationship = iota
	MENTION
)

type Syncer interface {
	Sync() error
}

type Config struct {
	Open struct {
		Types struct {
			Issue IssueConfig
		}
	}
}
type IssueConfig struct {
	Relationship Relationship
}
type Relationship struct {
	Assignee struct {
		Lists  []string
		Labels []string
	}
	Mention struct {
		Lists  []string
		Labels []string
	}
}

func GenerateCardName(title, repositoryName string) string {
	return fmt.Sprintf("**[github/%s]** %s", repositoryName, title)
}

func GenerateComment(commentNode github.CommentNode) string {
	comment := commentNode.Node

	return fmt.Sprintf("## @%s\n\n> %s \n\n___\n\n[View on GitHub](%s)",
		comment.Author.Login,
		strings.Replace(string(comment.Body), "\n", "\n> ", -1),
		comment.URL,
	)
}
