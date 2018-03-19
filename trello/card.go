package trello

import (
	"strings"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
	"fmt"
)

type Card struct {
	trelloCard *trello.Card
}

func (c *Card) SyncComments(comments, labelIDs []string) error {
	// Find existing comments for this card
	cardCommentActions, err := c.trelloCard.GetActions(
		map[string]string{
			"filter": "commentCard",
		},
	)
	if err != nil {
		return errors.Wrapf(err, "Error getting comments for card \"%s\"", c.trelloCard.ID)
	}

	if err = c.syncComments(cardCommentActions, comments, labelIDs); err != nil {
		return errors.Wrapf(err, "Error syncing comments for card \"%s\"", c.trelloCard.ID)
	}
	return nil
}

func (c *Card) syncComments(oldComments []*trello.Action, newComments, labelIDs []string) error {
	idx := 0
	newActivity := false

	for _, oldComment := range oldComments {
		// check for case: comments deleted from GH Issue
		if idx > len(newComments) {
			if err := c.trelloCard.DeleteComment(oldComment.ID); err != nil {
				return errors.Wrapf(err,
					"Error deleting stale comment \"%s\" from card \"%s\"", oldComment.ID, c.trelloCard.ID)
			}
			newActivity = true
		} else { // check for case: GH Issue comment text changed
			if oldComment.Data.Text != newComments[idx] {
				err := c.trelloCard.UpdateComment(newComments[idx], oldComment.ID)
				switch errors.Cause(err).(type) {
				case *trello.ErrorURLLengthExceeded:
					fmt.Printf(
						"[WARNING] Unable to update comment for card \"%s\""+
							"- request URL exceeded maximum length allowed.\n",
						c.trelloCard.Name,
					)
					return nil
				default:
					return errors.Wrapf(err, "Error updating comment \"%s\" to card \"%s\"", oldComment.ID, c.trelloCard.ID)
				}
				newActivity = true
			}
		}
		idx += 1
	}

	// check for case: new comments added to GH Issue
	if idx < len(newComments) {
		for i := idx; i < len(newComments); i++ {
			_, err := c.trelloCard.AddComment(newComments[i], trello.Defaults())
			switch errors.Cause(err).(type) {
			case *trello.ErrorURLLengthExceeded:
				if _, err := c.trelloCard.AddComment("<TOO LONG TO AUTOMATICALLY ADD>", trello.Defaults()); err != nil {
					return err
				}
				fmt.Printf(
					"[WARNING] Unable to automatically create comment for card \"%s\""+
						"- request URL exceeded maximum length allowed.\n",
					c.trelloCard.Name,
				)
				return nil
			default:
				return errors.Wrapf(err, "Error creating new comment to card \"%s\"", c.trelloCard.ID)
			}
			newActivity = true
		}
	}

	// re-apply labels if card was updated
	if newActivity {
		c.markNewAcivity(labelIDs)
	}
	return nil
}

func (c *Card) markNewAcivity(idLabels []string) error {
	if err := c.trelloCard.Update(map[string]string{
		"idLabels": strings.Join(idLabels, ","),
	}); err != nil {
		return err
	}
	return nil
}
