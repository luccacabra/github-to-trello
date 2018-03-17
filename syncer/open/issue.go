package open

import (
	"github.com/luccacabra/aws-github-to-trello/github"
	"github.com/luccacabra/aws-github-to-trello/syncer"
	trelloWrapper "github.com/luccacabra/aws-github-to-trello/trello"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

var _ syncer.Syncer = (*openIssueSyncer)(nil)

type openIssueSyncer struct {
	trelloBoard *trelloWrapper.BoardService
	github      *github.Client
	labelMap    map[syncer.UserRelationship][]string
	listMap     map[syncer.UserRelationship][]string
}

func NewIssueSyncer(
	githubClient *github.Client,
	trelloBoard *trelloWrapper.BoardService,
	config syncer.IssueConfig,
) (o *openIssueSyncer) {

	return &openIssueSyncer{
		github:      githubClient,
		trelloBoard: trelloBoard,

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
		if err = o.trelloBoard.CreateOrUpdateCard(
			o.convertIssueToCard(issue),
			o.labelMap[syncer.ASSIGNEE],
			o.listMap[syncer.ASSIGNEE],
		); err != nil {
			return err
		}
	}
	return nil
}

func (o *openIssueSyncer) convertIssueToCard(issue *github.IssueNode) *trello.Card {
	card := &trello.Card{
		Name: syncer.GenerateCardName(string(issue.Issue.Title), string(issue.Issue.Repository.Name)),
		Desc: string(issue.Issue.Body),
	}
	return card
}
