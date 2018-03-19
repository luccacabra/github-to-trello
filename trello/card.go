package trello

import (
	"strings"
	"fmt"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type Card struct {
	trelloCard *trello.Card
}

func (c *Card) SyncComments(comments, labelIDs []string) (bool, error) {
	// Find existing comments for this card
	cardCommentActions, err := c.trelloCard.GetActions(
		map[string]string{
			"filter": "commentCard",
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "Error getting comments for card \"%s\"", c.trelloCard.ID)
	}

	newActivity, err := c.syncComments(cardCommentActions, comments, labelIDs)
	if err != nil {
		return false, errors.Wrapf(err, "Error syncing comments for card \"%s\"", c.trelloCard.ID)
	}
	return newActivity, nil
}

func (c *Card) syncComments(oldComments []*trello.Action, newComments, labelIDs []string) (bool, error) {
	idx := 0
	newActivity := false
	//oldComments := reverse(oldCommentsReversed) // trello returns comments in reverse order

	for _, oldComment := range oldComments {
		// check for case: comments deleted from GH Issue
		if idx >= len(newComments) {
			if err := c.trelloCard.DeleteComment(oldComment.ID); err != nil {
				return false,
				errors.Wrapf(err,
					"Error deleting stale comment \"%s\" from card \"%s\"", oldComment.ID, c.trelloCard.ID)
			}
			newActivity = true
		} else { // check for case: GH Issue comment text changed
			if oldComment.Data.Text != newComments[idx] {
				err := c.trelloCard.UpdateComment(newComments[idx], oldComment.ID)
				if err!= nil {
					switch errors.Cause(err).(type) {
					case *trello.ErrorURLLengthExceeded:
						fmt.Printf(
							"[WARNING] Unable to update comment for card \"%s\""+
								"- request URL exceeded maximum length allowed.\n",
							c.trelloCard.Name,
						)
						return false, nil
					default:
						return false, errors.Wrapf(err, "Error updating comment \"%s\" to card \"%s\"", oldComment.ID, c.trelloCard.ID)
					}
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
			if err != nil {
				switch errors.Cause(err).(type) {
				case *trello.ErrorURLLengthExceeded:
					if _, err := c.trelloCard.AddComment("<TOO LONG TO AUTOMATICALLY ADD>", trello.Defaults()); err != nil {
						return false, err
					}
					fmt.Printf(
						"[WARNING] Unable to automatically create comment for card \"%s\""+
							"- request URL exceeded maximum length allowed.\n",
						c.trelloCard.Name,
					)
					return true, nil
				default:
					return false, errors.Wrapf(err, "Error creating new comment to card \"%s\"", c.trelloCard.ID)
				}
			}
			newActivity = true
		}
	}

	// re-apply labels if card was updated
	if newActivity {
		c.markNewAcivity(labelIDs)
	}
	return newActivity, nil
}

func (c *Card) markNewAcivity(idLabels []string) error {
	if err := c.trelloCard.Update(map[string]string{
		"idLabels": strings.Join(idLabels, ","),
	}); err != nil {
		return err
	}
	return nil
}

func reverse(a []*trello.Action) []*trello.Action{
	for i := len(a)/2-1; i >= 0; i-- {
		opp := len(a)-1-i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}