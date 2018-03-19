package trello

import (
	"log"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

func (c *Client) loadBoard(boardName string) error {
	boards, err := c.client.SearchBoards(boardName, trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Unable to get board ID for board %s", boardName)
	}
	if len(boards) > 0 {
		for _, board := range boards {
			if board.Name == boardName {
				c.board = board
				return nil
			}
		}
	}
	return errors.New("Unable to find board ID for board \"" + boardName + "\"")
}

func (c *Client) loadLabelMap() error {
	cards, err := c.client.SearchCards(c.config.LabelCardName, trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Unable to get label map for card id \"%s\"", c.config.LabelCardName)
	}

	for _, card := range cards {
		if card.Name == c.config.LabelCardName {
			for _, label := range card.Labels {
				if len(label.Name) > 0 {
					c.labelIDMap[label.Name] = label.ID
				}
			}
			return nil
		}
	}
	return errors.New("Could not find labels for card \"" + c.config.LabelCardName + "\"")
}

func (c *Client) loadListMap() error {
	lists, err := c.board.GetLists(trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Could not get lists for board \"%s\"", c.board.Name)
	}

	for _, list := range lists {
		c.listIDMap[list.Name] = list.ID
	}
	return nil
}

func (c *Client) loadResources(config ClientConfig) {
	if err := c.loadBoard(config.BoardName); err != nil {
		log.Fatalf("Unable to initalize trello connection: %s", err)
	}

	if err := c.loadListMap(); err != nil {
		log.Fatalf("Unable to initialize trello connection: %s", err)
	}

	if len(config.LabelMap) == 0 {
		if len(config.LabelCardName) == 0 {
			log.Fatal("Must specify either 'trello_label_map' or 'trello_label_card_name'")
		}
		if err := c.loadLabelMap(); err != nil {
			log.Fatalf("Unable to initialize trello connection: %s", err)
		}
	}
}
