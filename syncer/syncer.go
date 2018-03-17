package syncer

import "fmt"

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
