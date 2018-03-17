package trello

import (
	"fmt"
	"strings"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type BoardService service

func (b *BoardService) CreateOrUpdateCard(card *trello.Card, labelNames, listNames []string) error {
	// See if any cards on the board for this issue already exist
	cards, err := b.SearchCardsByName(card.Name)
	if err != nil {
		return err
	}

	// Index cards by list ID to check for their existence on the board's lists
	cardsByListID := map[string]*trello.Card{}
	for _, card := range cards {
		cardsByListID[card.IDList] = card
	}

	// See if if any cards on the trello board for each wanted list already exist
	for _, listName := range listNames {
		// if card doesn't exist for this list
		if _, ok := cardsByListID[b.listIDMap[listName]]; !ok {
			// create it
			card.IDList = b.listIDMap[listName]
			if err = b.CreateCard(card, labelNames); err != nil {
				return err
			}
		} // else update
	}
	return nil
}

func (b *BoardService) CreateCard(card *trello.Card, labelNames []string) error {
	err := b.trello.client.CreateCard(card, map[string]string{
		"idLabels": strings.Join(b.trello.getLabelIDsForNames(labelNames), ","),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create card \"%s\"", card.Name)
	}
	return nil
}

func (b *BoardService) SearchCardsByName(cardName string) ([]*trello.Card, error) {
	cards, err := b.trello.client.SearchCards(fmt.Sprintf("board:%s \"%s\"", b.trello.board.ID, cardName), trello.Defaults())
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up card \"%s\"", cardName)
	}
	return cards, nil
}
