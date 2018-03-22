package trello

import (
	"fmt"

	"github.com/luccacabra/github-to-trello/storage"
	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

func (c *Client) CreateNewCard(storageCard *storage.Card) (*Card, error) {
	fmt.Printf("\tCreating new trello card on list %s\n", storageCard.ListId)

	card, err := c.getCard(storageCard)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create new trello card \"%s\" on list %s", storageCard.Title, storageCard.ListId)
	}

	// If card already existed (from a failed run) update it with the correct info
	if len(card.storageCard.TrelloCardId) > 0 {
		if err = card.Update(map[string]string{
			"desc":     storageCard.Text,
			"idLabels": storageCard.LabelIds,
		}); err != nil {
			return nil, errors.Wrapf(err, "Failed to create new trello card \"%s\" on list %s", storageCard.Title, storageCard.ListId)
		}

	} else {
		// Otherwise, create the card
		if err := card.Create(); err != nil {
			return nil, errors.Wrapf(err, "Failed to create new trello card \"%s\" on list %s", storageCard.Title, storageCard.ListId)
		}
	}

	return card, nil
}
func (c *Client) SyncCommentsForNewCard(storageCard *storage.Card, comments []*storage.Comment) error {
	fmt.Printf("\t\tSyncing %d comments for card %s on list %s\n", len(comments), storageCard.Title, storageCard.ListId)
	card := c.NewCard(storageCard)
	_, err := card.SyncComments(comments)
	if err != nil {
		return errors.Wrap(err, "Error syncing comments for new card")
	}
	return nil
}

func (c *Client) GetLabelIdsForNames(labelNames []string) []string {
	labelIds := make([]string, len(labelNames))
	for idx, labelName := range labelNames {
		labelIds[idx] = c.labelIDMap[labelName]
	}

	return labelIds
}

func (c *Client) GetListIdForName(listName string) string {
	return c.listIDMap[listName]
}

func (c *Client) getCard(storageCard *storage.Card) (*Card, error) {
	trelloCards, err := c.client.SearchCards(fmt.Sprintf("board:%s \"%s\"", c.board.ID, storageCard.Title), trello.Defaults())
	fmt.Printf("Found %d trello cards with name %s\n", len(trelloCards), storageCard.Title)
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up card \"%s\"", storageCard.Title)
	}
	for _, trelloCard := range trelloCards {
		if trelloCard.IDList == storageCard.ListId {
			storageCard.TrelloCardId = trelloCard.ID
			return &Card{
				client:      c,
				storageCard: storageCard,
			}, nil
		}
	}
	return &Card{
		client:      c,
		storageCard: storageCard,
	}, nil
}
