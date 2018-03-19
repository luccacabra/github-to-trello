package syncer

import (
	"fmt"
	"strings"

	"github.com/luccacabra/github-to-trello/github"
	"github.com/luccacabra/github-to-trello/trello"
)

type UserRelationship int

const (
	ASSIGNEE UserRelationship = iota
	MENTION
)

type Config struct {
	Issue IssueConfig
}
type IssueConfig struct {
	Relationship Relationship
}
type Relationship struct {
	Assignee struct {
		Actions trello.Actions
	}
	Mention struct {
		Actions trello.Actions
	}
}

type Syncer interface {
	Sync() error
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
