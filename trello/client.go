/* Thin wrapper around github.com/luccacabra/trello */

package trello

import (
	"fmt"
	"strings"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type ClientConfig struct {
	BoardName     string
	LabelCardName string
	LabelMap      map[string]string
}

type Client struct {
	config ClientConfig

	client *trello.Client
	board  *trello.Board

	labelIDMap map[string]string // label Name -> label ID
	listIDMap  map[string]string // list Name  -> *trello.List

}

func NewClient(key, token string, config ClientConfig) *Client {
	c := &Client{
		client: trello.NewClient(key, token),
		config: config,
	}

	c.labelIDMap = map[string]string{}
	c.listIDMap = map[string]string{}

	c.loadResources(config)

	return c
}

func (c *Client) createCard(card *trello.Card, labelNames []string) error {
	fmt.Printf("Creating new card \"%s\" on list \"%s\"\n", card.Name, c.listIDMap[card.IDList])
	labelIDs := map[string]string{
		"idLabels": strings.Join(c.GetLabelIDsForNames(labelNames), ","),
	}

	err := c.client.CreateCard(card, labelIDs)
	if err != nil {
		switch errors.Cause(err).(type) {
		case *trello.ErrorURLLengthExceeded:
			card.Desc = ""
			if err = c.client.CreateCard(card, labelIDs); err != nil {
				return err
			}
			fmt.Printf(
				"[WARNING] Unable to automatically create card for issue \"%s\""+
					"- request URL exceeded maximum length allowed.\n"+
					"Issue created with no desc.\n",
				card.Name,
			)
			return nil
		default:
			return errors.Wrapf(err, "unable to create card \"%s\"", card.Name)
		}
	}
	return nil
}

func (c *Client) updateCard(card *trello.Card, desc string, labelNames []string) error {
	fmt.Printf("Updating card for issue \"%s\" on list \"%s\"\n", card.Name, c.listIDMap[card.IDList])
	labelIDs := strings.Join(c.GetLabelIDsForNames(labelNames), ",")
	err := card.Update(map[string]string{
		"desc":     desc,
		"idLabels": labelIDs,
	})
	switch errors.Cause(err).(type) {
	case *trello.ErrorURLLengthExceeded:
		if err = card.Update(map[string]string{
			"idLabels": labelIDs,
		}); err != nil {
			return err
		}
		fmt.Printf(
			"[WARNING] Unable to automatically update card for issue \"%s\""+
				"- request URL exceeded maximum length allowed.\n"+
				"Issue updated with no desc.\n",
			card.Name,
		)
		return nil
	default:
		return errors.Wrapf(err, "unable to update card \"%s\"", card.Name)
	}
	return nil
}

func (c *Client) updateCardInstances(card *trello.Card, actionConfig Actions, cardMap map[string]*Card) error {
	for _, listName := range actionConfig.Update.Lists {
		_, ok := cardMap[c.listIDMap[listName]]
		// check if card already exists in list
		if !ok {
			// create it
			card.IDList = c.listIDMap[listName]
			if err := c.createCard(card, actionConfig.Update.Labels); err != nil {
				return errors.Wrapf(err, "Error creating card for issue \"%s\"", card.Name)
			}
		} else {
			oldCard := cardMap[c.listIDMap[listName]]
			if card.Desc != oldCard.trelloCard.Desc {
				if err := c.updateCard(oldCard.trelloCard, card.Desc, actionConfig.Update.Labels); err != nil {
					return errors.Wrapf(err, "Error updating card \"%s\" for issue \"%s\"", card.ID, card.Name)
				}
			}
			// transfer old card to new card here to grab the client obj from the old card obj
			oldCard.trelloCard.Desc = card.Desc
			card = oldCard.trelloCard
		}
	}
	return nil
}

func (c *Client) CreateOrUpdateCardInstancesByActions(card *trello.Card, actionConfig Actions) error {
	// See if any cards on the board for this issue already exist
	cards, err := c.SearchCardsByName(card.Name)
	if err != nil {
		return err
	}

	// if no cards for this issue exist yet
	if len(cards) == 0 {
		// create one card for each configured list
		for _, listName := range actionConfig.Create.Lists {
			// create it
			card.IDList = c.listIDMap[listName]
			if err = c.createCard(card, actionConfig.Create.Labels); err != nil {
				return errors.Wrapf(err, "Error creating card for issue \"%s\"", card.Name)
			}
		}
	} else { // if cards for this issue exist
		// update them all to mirror issue text
		err := c.updateCardInstances(card, actionConfig, c.getCardsByIDs(cards))
		if err != nil {
			return errors.Wrapf(err, "Error updating cards for issue \"%s\"", card.Name)
		}
	}
	return nil
}

func (c *Client) SearchCardsByName(cardName string) ([]*Card, error) {
	trelloCards, err := c.client.SearchCards(fmt.Sprintf("board:%s \"%s\"", c.board.ID, cardName), trello.Defaults())
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up card \"%s\"", cardName)
	}
	cards := make([]*Card, len(trelloCards))
	for idx, trelloCard := range trelloCards {
		cards[idx] = &Card{
			trelloCard: trelloCard,
		}
	}
	return cards, nil
}

func (c *Client) GetLabelIDsForNames(labelNames []string) []string {
	idLabels := make([]string, len(labelNames))
	for idx, labelName := range labelNames {
		idLabels[idx] = c.labelIDMap[labelName]
	}
	return idLabels
}

func (c *Client) getCardsByIDs(cards []*Card) map[string]*Card {
	// Index cards by list ID to check for their existence on the board's lists
	cardsByListID := map[string]*Card{}
	for _, card := range cards {
		cardsByListID[card.trelloCard.IDList] = card
	}
	return cardsByListID
}
