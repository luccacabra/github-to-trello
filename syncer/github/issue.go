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
	labelMap map[syncer.UserRelationship][]string
	listMap  map[syncer.UserRelationship][]string
}

func NewIssueSyncer(
	githubClient *github.Client,
	trello *trelloWrapper.Client,
	config syncer.IssueConfig,
) (o *issueSyncer) {

	return &issueSyncer{
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

func (i *issueSyncer) sync(issues []github.IssueNode, issueType syncer.UserRelationship) error {
	for _, issue := range issues {
		// graphql API returns empty nodes sometimes
		if len(issue.Issue.Title) == 0 {
			continue
		}
		fmt.Printf("Syncing issue \"%s\"\n", issue.Issue.Title)
		card, err := i.trello.CreateOrUpdateCard(
			i.convertIssueToCard(issue),
			i.labelMap[issueType],
			i.listMap[issueType],
		)
		if err != nil {
			switch errors.Cause(err).(type) {
			case *trello.ErrorURLLengthExceeded:
				fmt.Printf(
					"[WARNING] Unable to automatically create or update  card for issue \"%s\""+
						"- request URL exceeded maximum length allowed.\n"+
						"Issue created or updated with no desc.\n",
					issue.Issue.Title,
				)
				continue
			default:
				return errors.Wrapf(err, "Error syncing issue \"%s\"", issue.Issue.Title)
			}
		}

		fmt.Printf("Syncing comments for issue \"%s\"\n", issue.Issue.Title)
		err = i.syncComments(card, issue)
		if err != nil {
			switch errors.Cause(err).(type) {
			case *trello.ErrorURLLengthExceeded:
				fmt.Printf(
					"[WARNING] Unable to automatically create or update  comment for issue \"%s\""+
						"- request URL exceeded maximum length allowed.\n"+
						"Issue Comment created or updated with no text.\n",
					issue.Issue.Title,
				)
				continue
			default:
				return errors.Wrapf(err, "Error syncing comment for issue \"%s\"", issue.Issue.Title)
			}
		}
	}
	return nil
}

func (i *issueSyncer) syncAssigned() error {
	issues, err := i.github.Issues.Assigned()
	if err != nil {
		return errors.Wrap(err, "unable to sync open assigned issues")
	}
	if err = i.sync(issues, syncer.ASSIGNEE); err != nil {
		return errors.Wrap(err, "unable to sync open assigned issues")
	}
	return nil
}

func (i *issueSyncer) syncMentioned() error {
	issues, err := i.github.Issues.Mentioned()
	if err != nil {
		return errors.Wrap(err, "unable to sync open mentioned issues")
	}
	if err = i.sync(issues, syncer.MENTION); err != nil {
		return errors.Wrap(err, "unable to sync open mentioned issues")
	}
	return nil
}

func (i *issueSyncer) syncClosed() error {
	return nil
}

func (i *issueSyncer) syncComments(card *trelloWrapper.Card, node github.IssueNode) error {
	issue := node.Issue
	fmt.Printf("Syncing comments for issue \"%s\"\n", issue.Title)
	comments := make([]string, len(issue.Comments.Edges))
	for idx, comment := range issue.Comments.Edges {
		comments[idx] = syncer.GenerateComment(comment)
	}
	if err := card.SyncComments(comments, i.labelMap[syncer.ASSIGNEE]); err != nil {
		return errors.Wrapf(err, "Error syncing comments for issue \"%s\"", issue.Title)
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
