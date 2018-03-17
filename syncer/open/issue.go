package open

import (
	"fmt"

	"github.com/luccacabra/github-to-trello/github"
	"github.com/luccacabra/github-to-trello/syncer"
	trelloWrapper "github.com/luccacabra/github-to-trello/trello"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

var _ syncer.Syncer = (*openIssueSyncer)(nil)

type openIssueSyncer struct {
	trello   *trelloWrapper.Client
	github   *github.Client
	labelMap map[syncer.UserRelationship][]string
	listMap  map[syncer.UserRelationship][]string
}

func NewIssueSyncer(
	githubClient *github.Client,
	trello *trelloWrapper.Client,
	config syncer.IssueConfig,
) (o *openIssueSyncer) {

	return &openIssueSyncer{
		github: githubClient,
		trello: trello,
		labelMap: map[syncer.UserRelationship][]string{
			syncer.ASSIGNEE: config.Relationship.Assignee.Labels,
			syncer.MENTION:  config.Relationship.Mention.Labels,
		},
		listMap: map[syncer.UserRelationship][]string{
			syncer.ASSIGNEE: config.Relationship.Assignee.Lists,
			syncer.MENTION:  config.Relationship.Mention.Lists,
		},
	}
}

func (o *openIssueSyncer) Sync() error {
	if err := o.syncAssigned(); err != nil {
		return err
	}
	return nil
}

func (o *openIssueSyncer) syncAssigned() error {
	issues, err := o.github.Issues.Assigned()
	if err != nil {
		return errors.Wrap(err, "unable to sync open assigned issues")
	}

	for _, issue := range issues {
		fmt.Printf("Syncing issue \"%s\"\n", issue.Issue.Title)
		card, err := o.trello.CreateOrUpdateCard(
			o.convertIssueToCard(issue),
			o.labelMap[syncer.ASSIGNEE],
			o.listMap[syncer.ASSIGNEE],
		)
		if err != nil {
			return err
		}

		fmt.Printf("Syncing comments for issue \"%s\"\n", issue.Issue.Title)
		comments := make([]string, len(issue.Issue.Comments.Edges))
		for idx, comment := range issue.Issue.Comments.Edges {
			comments[idx] = syncer.GenerateComment(comment)
		}
		if err = card.SyncComments(comments); err != nil {
			return err
		}
	}
	return nil
}

func (o *openIssueSyncer) convertIssueToCard(issue github.IssueNode) *trello.Card {
	card := &trello.Card{
		Name: syncer.GenerateCardName(string(issue.Issue.Title), string(issue.Issue.Repository.Name)),
		Desc: string(issue.Issue.Body),
	}
	return card
}
