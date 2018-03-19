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

func (c *Client) LabelCardName() string {
	return c.config.LabelCardName
}

func (c *Client) CreateOrUpdateCardInstancesByActions(card *trello.Card, actionConfig Actions) (bool, error) {
	// See if any cards on the board for this issue already exist
	cards, err := c.GetCardsByName(card.Name)
	if err != nil {
		return false, err
	}

	// if no cards for this issue exist yet
	if len(cards) == 0 {
		// create one card for each configured list
		for _, listName := range actionConfig.Create.Lists {
			// create it
			card.IDList = c.listIDMap[listName]
			if err = c.CreateCard(card, actionConfig.Create.Labels); err != nil {
				return false, errors.Wrapf(err, "Error creating card for issue \"%s\"", card.Name)
			}
		}
	} else { // if cards for this issue exist
		// update them all to mirror issue text
		newActivity, err := c.updateCardInstances(card, cards, actionConfig, c.getCardsByIDs(cards))
		if err != nil {
			return false, errors.Wrapf(err, "Error updating cards for issue \"%s\"", card.Name)
		}
		return newActivity, nil
	}
	return true, nil
}

func (c *Client) GetCardsByName(cardName string) ([]*Card, error) {
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

func (c *Client) getCardInList(card *trello.Card, listName string) (*Card, error) {
	trelloCards, err := c.client.SearchCards(fmt.Sprintf("board:%s \"%s\"", c.board.ID, card.Name), trello.Defaults())
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up card \"%s\"", card.Name)
	}
	for _, trelloCard := range trelloCards {
		if trelloCard.IDList == c.listIDMap[listName] {
			return &Card{trelloCard:trelloCard}, nil
		}
	}
	return nil, nil
}

func (c *Client) CloseCards(cardInstances []*Card, actionConfig Actions) error {
	tempCard := cardInstances[0]
	// create closed cards in closed lists
	for _, listName := range actionConfig.Close.Lists {
		tempCard.trelloCard.IDList = c.listIDMap[listName]
		if err := c.CreateCard(tempCard.trelloCard, actionConfig.Close.Labels); err != nil {
			return err
		}
	}

	// delete all other card instances
	for _, cardInstance := range cardInstances {
		if err := c.client.DeleteCard(cardInstance.trelloCard); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetCardInstancesByName() (map[string][]*Card, error) {
	cards, err := c.board.GetCards(trello.Defaults())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get cards for board %s", c.board.ID)
	}
	cardMap := map[string][]*Card{}
	for _, card := range cards {
		_, ok := cardMap[card.Name]
		if !ok {
			cardMap[card.Name] = []*Card{{trelloCard: card}}
		} else {
			cardMap[card.Name] = append(cardMap[card.Name], &Card{trelloCard: card})
		}
	}
	return cardMap, nil
}

func (c *Client) getLabelIDsForNames(labelNames []string) []string {
	idLabels := make([]string, len(labelNames))
	for idx, labelName := range labelNames {
		idLabels[idx] = c.labelIDMap[labelName]
	}
	return idLabels
}

func (c *Client) CreateCard(card *trello.Card, labelNames []string) error {
	fmt.Printf("Creating new card \"%s\" on list \"%s\"\n", card.Name, c.listIDMap[card.IDList])
	labelIDs := map[string]string{
		"idLabels": strings.Join(c.getLabelIDsForNames(labelNames), ","),
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

func (c *Client) SyncComments(card *trello.Card, comments []string, actionConfig Actions) error {
	cards, err := c.GetCardsByName(card.Name)
	if err != nil {
		return err
	}

	newActivity := false

	for _, card := range cards {
		newCommentActivity, err := card.SyncComments(comments, c.getLabelIDsForNames(actionConfig.Update.Labels))
		if err != nil {
			return errors.Wrapf(err, "\tError syncing comments for issue \"%s\" - \n", card.trelloCard.Name)
		}
		if newCommentActivity{
			newActivity = true
		}
	}

	// check to see if card was removed by user previous to update
	// if new activity is found, cards need to be recreated in designated update lists, with comments
	if newActivity {
		for _, listName := range actionConfig.Update.Lists {
			commentCard, err := c.getCardInList(card, listName)
			if err != nil {
				return err
			}
			if commentCard != nil{
			} else {
			}
			if commentCard == nil {
				card.IDList = c.listIDMap[listName]
				if err = c.CreateCard(card, actionConfig.Update.Labels); err != nil {
					return err
				}
				commentCard = &Card{trelloCard:card}
				commentCard.SyncComments(comments, c.getLabelIDsForNames(actionConfig.Update.Labels))
			}
		}
	}


	return nil
}

func (c *Client) updateCard(card *trello.Card, desc string, labelNames []string) error {
	fmt.Printf("Updating card for issue \"%s\" on list \"%s\"\n", card.Name, c.listIDMap[card.IDList])
	labelIDs := strings.Join(c.getLabelIDsForNames(labelNames), ",")
	err := card.Update(map[string]string{
		"desc":     desc,
		"idLabels": labelIDs,
	})
	if err != nil {
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
	}
	return nil
}

// don't write out to lists on an update unless content has changed
func (c *Client) updateCardInstances(newCard *trello.Card, OldCards []*Card, actionConfig Actions, cardMap map[string]*Card) (bool, error) {
	newActivity := false
	for _, oldCard := range OldCards {
		if newCard.Desc != oldCard.trelloCard.Desc {
			if err := c.updateCard(oldCard.trelloCard, newCard.Desc, actionConfig.Update.Labels); err != nil {
				return false, errors.Wrapf(err, "Error updating card \"%s\" for issue \"%s\"", newCard.ID, newCard.Name)
			}
		}
		newActivity = true
		// transfer old card to new card here to grab the client obj from the old card obj
		oldCard.trelloCard.Desc = newCard.Desc
		newCard = oldCard.trelloCard
	}
	return newActivity, nil
}

func (c *Client) getCardsByIDs(cards []*Card) map[string]*Card {
	// Index cards by list ID to check for their existence on the board's lists
	cardsByListID := map[string]*Card{}
	for _, card := range cards {
		cardsByListID[card.trelloCard.IDList] = card
	}
	return cardsByListID
}