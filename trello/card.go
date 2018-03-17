package trello

import (
	"github.com/luccacabra/trello"
)

type Card struct {
	trelloCard *trello.Card
}

func (c *Card) SyncComments(comments []string) error {
	// Find existing comments for this card
	cardCommentActions, err := c.trelloCard.GetActions(
		map[string]string{
			"filter": "commentCard",
		},
	)
	if err != nil {
		return err
	}

	// if none exist create them
	if len(cardCommentActions) == 0 {
		c.addComments(comments)
	} else { // sync existing comments

	}
	return nil
}

func (c *Card) addComments(comments []string) error {
	for _, comment := range comments {
		if _, err := c.trelloCard.AddComment(comment, trello.Defaults()); err != nil {
			return err
		}
	}
	return nil
}
