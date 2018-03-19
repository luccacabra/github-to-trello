package github

import (
	"fmt"

	"github.com/luccacabra/github-to-trello/github"
	"github.com/luccacabra/github-to-trello/syncer"
	trelloWrapper "github.com/luccacabra/github-to-trello/trello"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

var _ syncer.Syncer = (*issueSyncer)(nil)

type issueSyncer struct {
	trello   *trelloWrapper.Client
	github   *github.Client

	config syncer.IssueConfig
}

func NewIssueSyncer(
	githubClient *github.Client,
	trello *trelloWrapper.Client,
	config syncer.IssueConfig,
) (o *issueSyncer) {

	return &issueSyncer{
		github: githubClient,
		trello: trello,
		config: config,
	}
}

func (i *issueSyncer) Sync() error {
	if err := i.syncAssigned(); err != nil {
		return errors.Wrap(err, "Error syncing open assigned issues")
	}
	if err := i.syncMentioned(); err != nil {
		return errors.Wrap(err, "Error syncing open mentioned issues")
	}
	if err := i.syncClosed(); err != nil {
		return errors.Wrap(err, "Error syncing closed assigned & mentioned issues")
	}
	return nil
}

func (i *issueSyncer) sync(issues []github.IssueNode, actionConfig trelloWrapper.Actions) error {
	for _, issue := range issues {
		// graphql API returns empty nodes sometimes
		if len(issue.Issue.Title) == 0 {
			continue
		}

		card := i.convertIssueToCard(issue)

		fmt.Printf("Syncing issue \"%s\"\n", issue.Issue.Title)
		err := i.trello.CreateOrUpdateCardInstancesByActions(card, actionConfig)
		if err != nil {
				return errors.Wrapf(err, "Error syncing issue \"%s\"", issue.Issue.Title)
		}

		err = i.syncComments(card, issue, actionConfig)
		if err != nil {
			return errors.Wrapf(err, "Error syncing comment for issue \"%s\"", issue.Issue.Title)
		}
	}
	return nil
}

func (i *issueSyncer) syncAssigned() error {
	issues, err := i.github.Issues.Assigned()
	if err != nil {
		return errors.Wrap(err, "unable to sync open assigned issues")
	}
	if err = i.sync(issues, i.config.Relationship.Assignee.Actions); err != nil {
		return errors.Wrap(err, "unable to sync open assigned issues")
	}
	return nil
}

func (i *issueSyncer) syncMentioned() error {
	issues, err := i.github.Issues.Mentioned()
	if err != nil {
		return errors.Wrap(err, "unable to sync open mentioned issues")
	}
	if err = i.sync(issues, i.config.Relationship.Mention.Actions); err != nil {
		return errors.Wrap(err, "unable to sync open mentioned issues")
	}
	return nil
}

func (i *issueSyncer) syncClosed() error {
	return nil
}

func (i *issueSyncer) syncComments(card *trello.Card, node github.IssueNode, actionConfig trelloWrapper.Actions) error {
	issue := node.Issue
	fmt.Printf("Syncing comments for issue \"%s\"\n", issue.Title)
	comments := make([]string, len(issue.Comments.Edges))
	for idx, comment := range issue.Comments.Edges {
		comments[idx] = syncer.GenerateComment(comment)
	}


	cards, err := i.trello.SearchCardsByName(card.Name)
	if err != nil {
		return err
	}
	for _, card := range cards {
		if err := card.SyncComments(comments, i.trello.GetLabelIDsForNames(actionConfig.Update.Labels)); err != nil {
			return errors.Wrapf(err, "Error syncing comments for issue \"%s\"", issue.Title)
		}
	}
	return nil
}

func (i *issueSyncer) convertIssueToCard(issue github.IssueNode) *trello.Card {
	card := &trello.Card{
		Name: syncer.GenerateCardName(string(issue.Issue.Title), string(issue.Issue.Repository.Name)),
		Desc: string(issue.Issue.Body),
	}
	return card
}
