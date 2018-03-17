package syncer

import (
	"fmt"
	"github.com/luccacabra/github-to-trello/github"
	"regexp"
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
	var re = regexp.MustCompile(`\n(.)`)
	comment := commentNode.Node

	body := strings.Replace(string(comment.Body), "\n\n", "\n>\n", -1)
	body = re.ReplaceAllString(body, `\n>$1`)
	return fmt.Sprintf("## @%s \n\n > %s\n\n___\n\n[View on GitHub](%s)",
		comment.Author.Login,
		body,
		comment.URL,
	)
}
