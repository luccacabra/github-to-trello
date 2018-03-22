package trello

import (
	"fmt"

	"github.com/luccacabra/github-to-trello/storage"
	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type Card struct {
	client      *Client
	storageCard *storage.Card
}

func (c *Client) NewCard(storageCard *storage.Card) *Card {
	return &Card{
		client:      c,
		storageCard: storageCard,
	}
}

/*
*
* CARD
*
 */
func (c *Card) Create() error {
	fmt.Printf("\tCreating new trello card on list %s\n", c.storageCard.ListId)
	path := "cards"
	data := map[string]string{
		"name":     c.storageCard.Title,
		"desc":     c.storageCard.Text,
		"pos":      "0",
		"idLabels": c.storageCard.LabelIds,
		"idList":   c.storageCard.ListId,
	}

	trelloCard := &trello.Card{}
	if err := c.client.Post(path, data, trelloCard); err != nil {
		return errors.Wrapf(err, "Failed to create new trello card %s on list %s", c.storageCard.Title, c.storageCard.ListId)
	}

	c.storageCard.TrelloCardId = trelloCard.ID

	return nil
}

func (c *Card) Update(args map[string]string) error {
	path := fmt.Sprintf("cards/%s", c.storageCard.TrelloCardId)
	return c.client.Put(path, args, c)
}

func (c *Card) GetId() string {
	return c.storageCard.TrelloCardId
}

/*
*
*  CARD - COMMENTS
*
 */
func (c *Card) CreateComment(comment string) error {
	path := fmt.Sprintf("cards/%s/actions/comments", c.storageCard.TrelloCardId)
	if err := c.client.Post(path, map[string]string{"text": comment}, nil); err != nil {
		return errors.Wrapf(err, "Error commenting on card %s", c.storageCard.TrelloCardId)
	}
	return nil
}

func (c *Card) DeleteComment(commentActionID string) error {
	path := fmt.Sprintf("cards/%s/actions/%s/comments", c.storageCard.TrelloCardId, commentActionID)
	if err := c.client.Delete(path, map[string]string{}, nil); err != nil {
		return errors.Wrapf(err, "Error deleting comment '%s' from card '%s'", commentActionID, c.storageCard.TrelloCardId)
	}
	return nil
}

func (c *Card) GetComments() (trello.ActionCollection, error) {
	actionCollection := &trello.ActionCollection{}
	path := fmt.Sprintf("cards/%s/actions", c.storageCard.TrelloCardId)
	if err := c.client.Get(
		path,
		map[string]string{
			"filter": "commentCard",
		},
		actionCollection,
	); err != nil {
		return nil, err
	}
	return *actionCollection, nil
}

func (c *Card) UpdateComment(comment, commentActionID string) error {
	path := fmt.Sprintf("cards/%s/actions/%s/comments", c.storageCard.TrelloCardId, commentActionID)
	if err := c.client.Put(
		path,
		map[string]string{
			"text": comment,
		},
		nil); err != nil {
		return errors.Wrapf(err, "Error updating comment '%s' on card '%s'", commentActionID, c.storageCard.TrelloCardId)
	}
	return nil
}

func (c *Card) SyncComments(comments []*storage.Comment) (bool, error) {
	fmt.Printf("\t\tSyncing comments for trello card %s on list %s\n", c.storageCard.Title, c.storageCard.LabelIds)
	// Find existing comments for this card
	cardCommentActions, err := c.GetComments()
	if err != nil {
		return false, errors.Wrapf(err, "Error getting comments for card \"%s\"", c.storageCard.TrelloCardId)
	}

	newActivity, err := c.syncComments(cardCommentActions, comments)
	if err != nil {
		return false, errors.Wrapf(err, "Error syncing comments for card \"%s\"", c.storageCard.TrelloCardId)
	}
	return newActivity, nil
}

func (c *Card) syncComments(oldComments []*trello.Action, newComments []*storage.Comment) (bool, error) {
	fmt.Printf(
		"\t\tSyncing %d comments for trello card %s on list %s\n",
		len(newComments),
		c.storageCard.Title,
		c.storageCard.ListId,
	)
	// trello posts comments in reverse order, so flip 'em here
	newComments = reverse(newComments)

	idx := 0
	newActivity := false
	//oldComments := reverse(oldCommentsReversed) // trello returns comments in reverse order

	for _, oldComment := range oldComments {
		// check for case: comments deleted from GH Issue
		if idx >= len(newComments) {
			fmt.Printf("\t\t\tDeleting stale comment %s...\n", oldComment.Data.Text[:10])
			if err := c.DeleteComment(oldComment.ID); err != nil {
				return false,
					errors.Wrapf(err,
						"Error deleting stale comment \"%s\" from card \"%s\"", oldComment.ID, c.storageCard.TrelloCardId)
			}
			newActivity = true
		} else { // check for case: GH Issue comment text changed
			if oldComment.Data.Text != newComments[idx].Body {
				fmt.Printf("\t\t\tUpdating stale comment %s...\n", oldComment.Data.Text[:10])
				err := c.UpdateComment(newComments[idx].Body, oldComment.ID)
				if err != nil {
					switch errors.Cause(err).(type) {
					case *trello.ErrorURLLengthExceeded:
						fmt.Printf(
							"[WARNING] Unable to update comment for card \"%s\""+
								"- request URL exceeded maximum length allowed.\n",
							c.storageCard.Title,
						)
						return false, nil
					default:
						return false, errors.Wrapf(err, "Error updating comment \"%s\" to card \"%s\"", oldComment.ID, c.storageCard.TrelloCardId)
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
			fmt.Printf("\t\t\tAdding new comment %s...\n", newComments[i].Body[:10])
			if err := c.CreateComment(newComments[i].Body); err != nil {
				return false, errors.Wrapf(err, "Error creating new comment to card \"%s\"", c.storageCard.TrelloCardId)
			}
			newActivity = true
		}
	}

	// re-apply labels if card was updated
	if newActivity {
		c.markNewAcivity()
	}
	return newActivity, nil
}

func (c *Card) markNewAcivity() error {
	if err := c.Update(map[string]string{
		"idLabels": c.storageCard.LabelIds,
	}); err != nil {
		return err
	}
	return nil
}

func reverse(a []*storage.Comment) []*storage.Comment {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}
