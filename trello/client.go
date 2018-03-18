/* Thin wrapper around github.com/luccacabra/trello */

package trello

import (
	"fmt"
	"strings"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type Config struct {
	BoardName     string
	LabelCardName string
	LabelMap      map[string]string
}

type Client struct {
	config Config

	client *trello.Client
	board  *trello.Board

	labelIDMap map[string]string // label Name -> label ID
	listIDMap  map[string]string // list Name  -> *trello.List

}

func NewClient(key, token string, config Config) *Client {
	c := &Client{
		client: trello.NewClient(key, token),
		config: config,
	}

	c.labelIDMap = map[string]string{}
	c.listIDMap = map[string]string{}

	c.loadResources(config)

	return c
}

func (c *Client) CreateCard(card *trello.Card, labelNames []string) error {
	fmt.Printf("Creating new card \"%s\"\n", card.Name)

	err := c.client.CreateCard(card, map[string]string{
		"idLabels": strings.Join(c.getLabelIDsForNames(labelNames), ","),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create card \"%s\"", card.Name)
	}
	return nil
}

func (c *Client) CreateOrUpdateCard(card *trello.Card, labelNames, listNames []string) (*Card, error) {
	// See if any cards on the board for this issue already exist
	cards, err := c.SearchCardsByName(card.Name)
	if err != nil {
		return nil, err
	}

	// Index cards by list ID to check for their existence on the board's lists
	cardsByListID := map[string]*trello.Card{}
	for _, card := range cards {
		cardsByListID[card.IDList] = card
	}

	// See if if any cards on the trello board for each wanted list already exist
	for _, listName := range listNames {
		// if card doesn't exist for this list
		_, ok := cardsByListID[c.listIDMap[listName]]
		if !ok {
			// create it
			card.IDList = c.listIDMap[listName]
			if err = c.CreateCard(card, labelNames); err != nil {
				return nil, errors.Wrapf(err, "Error creating card for issue \"%s\"", card.Name)
			}
		} else { // else update it
			oldCard := cardsByListID[c.listIDMap[listName]]
			if card.Desc != oldCard.Desc {
				fmt.Printf("Updating card for issue \"%s\"\n", card.Name)
				if err = oldCard.Update(map[string]string{
					"desc":     card.Desc,
					"idLabels": strings.Join(c.getLabelIDsForNames(labelNames), ","),
				}); err != nil {
					return nil, errors.Wrapf(err, "Error updating card \"%s\" for issue \"%s\"", card.ID, card.Name)
				}
			}
			// transfer old card to new card here to grab the client obj from the old card obj
			oldCard.Desc = card.Desc
			card = oldCard
		}
	}
	return &Card{trelloCard: card}, nil
}

func (c *Client) SearchCardsByName(cardName string) ([]*trello.Card, error) {
	cards, err := c.client.SearchCards(fmt.Sprintf("board:%s \"%s\"", c.board.ID, cardName), trello.Defaults())
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up card \"%s\"", cardName)
	}
	return cards, nil
}

func (c *Client) getLabelIDsForNames(labelNames []string) []string {
	idLabels := make([]string, len(labelNames))
	for idx, labelName := range labelNames {
		idLabels[idx] = c.labelIDMap[labelName]
	}
	return idLabels
}
