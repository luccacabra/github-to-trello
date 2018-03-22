package github

import (
	"fmt"
	"strings"

	"github.com/luccacabra/github-to-trello/github"
	"github.com/luccacabra/github-to-trello/storage"
	"github.com/luccacabra/github-to-trello/syncer"
	trelloWrapper "github.com/luccacabra/github-to-trello/trello"

	"github.com/pkg/errors"
)

var _ syncer.Syncer = (*issueSyncer)(nil)

type issueSyncer struct {
	trello *trelloWrapper.Client
	github *github.Client

	storage *storage.Storage

	config map[syncer.UserRelationship]trelloWrapper.Actions
}

func NewIssueSyncer(
	githubClient *github.Client,
	trello *trelloWrapper.Client,
	storage *storage.Storage,
	config syncer.IssueConfig,
) (o *issueSyncer) {
	actionConfig := map[syncer.UserRelationship]trelloWrapper.Actions{}
	actionConfig[syncer.ASSIGNEE] = config.Relationship.Assignee.Actions
	actionConfig[syncer.MENTION] = config.Relationship.Assignee.Actions

	return &issueSyncer{
		github:  githubClient,
		trello:  trello,
		storage: storage,
		config:  actionConfig,
	}
}

func (i *issueSyncer) Sync() error {
	if err := i.syncAssigned(); err != nil {
		return errors.Wrap(err, "Error syncing open assigned issues")
	}
	if err := i.syncMentioned(); err != nil {
		return errors.Wrap(err, "Error syncing open mentioned issues")
	}
	//if err := i.syncClosed(); err != nil {
	//	return errors.Wrap(err, "Error syncing closed assigned & mentioned issues")
	//}
	return nil
}

func (i *issueSyncer) sync(issueNodes []github.IssueNode, relationship syncer.UserRelationship) error {
	for _, issueNode := range issueNodes {
		fmt.Printf("Syncing issue \"%s\"\n", issueNode.Issue.Title)

		// graphql API returns empty nodes sometimes
		if len(issueNode.Issue.Title) == 0 {
			continue
		}

		issue, err := i.storage.FindIssue(string(issueNode.Issue.ID))
		if err != nil {
			return errors.Wrapf(err, "Error syncing issue %s", issueNode.Issue.Title)
		}
		// New issue
		if issue == nil {
			if err = i.syncNew(issueNode, i.config[relationship]); err != nil {
				fmt.Println("here 3")
				return errors.Wrapf(err, "Error syncing issue %s", issueNode.Issue.Title)
			}
		}
		// Update Existing Issue

		//card := i.convertIssueToCard(issue)
		//
		//fmt.Printf("Syncing issue \"%s\"\n", issue.Issue.Title)
		//_, err := i.trello.CreateOrUpdateCardInstancesByActions(card, actionConfig)
		//if err != nil {
		//	return errors.Wrapf(err, "Error syncing issue \"%s\"", issue.Issue.Title)
		//}
		//
		//err = i.syncComments(car d, issue, actionConfig)
		//if err != nil {
		//	return errors.Wrapf(err, "Error syncing comment for issue \"%s\"", issue.Issue.Title)
		//}
	}
	return nil
}

func (i *issueSyncer) syncAssigned() error {
	issues, err := i.github.Issues.Assigned()
	if err != nil {
		return err
	}
	fmt.Printf("Syncing %d assigned issues", len(issues))
	if err = i.sync(issues, syncer.ASSIGNEE); err != nil {
		return err
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

func (i *issueSyncer) syncNew(issueNode github.IssueNode, actionConfig trelloWrapper.Actions) error {
	fmt.Printf("Syncing new issue \"%s\"\n", issueNode.Issue.Title)
	issue := i.convertIssueNodeToIssue(issueNode)

	if err := i.storage.SaveNewIssue(issue); err != nil {
		return errors.Wrapf(err, "Error syncing new issue \"%s\"", issue.Title)
	}

	for _, listName := range actionConfig.Create.Lists {
		fmt.Printf("\tSyncing new issue for list %s\n", listName)

		card := i.convertIssueToCard(issue, actionConfig.Create.Labels, listName)

		if err := i.createNewCard(card, issue.Comments); err != nil {
			return err
		}
	}

	return nil
}

func (i *issueSyncer) createNewCard(storageCard *storage.Card, comments []*storage.Comment) error {
	// Create corresponding trello card
	trelloCard, err := i.trello.CreateNewCard(storageCard)
	if err != nil {
		return err
	}

	// Save new card
	if err := i.storage.SaveNewCard(storageCard); err != nil {
		return errors.Wrapf(err, "Error creating new card \"%s\" on list %s", storageCard.Title, storageCard.ListId)
	}

	// Sync Issue comments
	_, err = trelloCard.SyncComments(comments)
	if err != nil {
		return err
	}

	return nil
}

func (i *issueSyncer) convertIssueToCard(issue *storage.Issue, labelNames []string, listName string) *storage.Card {
	return &storage.Card{
		IssueId: issue.Id,
		Title:   issue.Title,
		Text:    issue.Body,

		LabelIds: strings.Join(i.trello.GetLabelIdsForNames(labelNames), ","),
		ListId:   i.trello.GetListIdForName(listName),
	}
}

func (i *issueSyncer) convertIssueNodeToIssue(issueNode github.IssueNode) *storage.Issue {
	comments := make([]*storage.Comment, len(issueNode.Issue.Comments.Edges))
	for idx, commentNode := range issueNode.Issue.Comments.Edges {
		comments[idx] = &storage.Comment{
			IssueId: string(issueNode.Issue.ID),
			Body:    syncer.GenerateComment(commentNode),
		}
	}
	return &storage.Issue{
		Body:       syncer.GenerateCardDesc(string(issueNode.Issue.Body), string(issueNode.Issue.URL)),
		IssueId:    string(issueNode.Issue.ID),
		Number:     int64(issueNode.Issue.Number),
		Repository: string(issueNode.Issue.Repository.Name),
		Title:      string(issueNode.Issue.Title),
		URL:        string(issueNode.Issue.URL),

		Comments: comments,
	}
}

// ==============================================================================

//
//func (i *issueSyncer) syncClosed() error {
//	fmt.Println("Syncing closed issues")
//	cards, err := i.trello.GetCardInstancesByName()
//	if err != nil {
//		return errors.Wrap(err, "unable to sync closed issues")
//	}
//
//	assignedIssues, err := i.github.Issues.Assigned()
//	if err != nil {
//		return errors.Wrap(err, "unable to sync closed issues")
//	}
//	if err = i.close(assignedIssues, cards, i.config.Relationship.Assignee.Actions); err != nil {
//		return errors.Wrap(err, "unable to sync closed issues")
//	}
//
//	mentionedIssues, err := i.github.Issues.Mentioned()
//	if err != nil {
//		return errors.Wrap(err, "unable to sync closed issues")
//	}
//	if err = i.close(mentionedIssues, cards, i.config.Relationship.Mention.Actions); err != nil {
//		return errors.Wrap(err, "unable to sync closed issues")
//	}
//	return nil
//}
//
//func (i *issueSyncer) close(
//	issues []github.IssueNode,
//	cardMap map[string][]*trelloWrapper.Card,
//	actionConfig trelloWrapper.Actions,
//) error {
//	issueMap := i.getIssuesByCardName(issues)
//	for cardName, cards := range cardMap {
//		// card isn't in active issues so close
//		if _, ok := issueMap[cardName]; !ok {
//			if cardName != i.trello.LabelCardName() {
//				fmt.Printf("closing issue \"%s\"\n", cardName)
//				i.trello.CloseCards(cards, actionConfig)
//			}
//		}
//	}
//	return nil
//}
