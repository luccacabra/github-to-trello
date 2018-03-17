/* Thin wrapper around github.com/luccacabra/trello */

package trello

import (
	"log"

	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type service struct {
	trello     *Trello
	labelIDMap map[string]string // label Name -> label ID
	listIDMap  map[string]string // list Name  -> list ID
}

type Config struct {
	BoardName     string
	LabelCardName string
	LabelMap      map[string]string
}

type Trello struct {
	config Config

	client *trello.Client
	board  *trello.Board

	common service

	Board *BoardService
}

func NewClient(key, token string, config Config) *Trello {
	t := &Trello{
		client: trello.NewClient(key, token),
		config: config,
	}

	t.loadResources(config)

	t.common.labelIDMap = map[string]string{}
	t.common.listIDMap = map[string]string{}

	t.common.trello = t
	t.Board = (*BoardService)(&t.common)

	return t
}

func (t *Trello) loadResources(config Config) {
	if err := t.loadBoard(config.BoardName); err != nil {
		log.Fatalf("Unable to initalize trello connection: %s", err)
	}

	if err := t.loadLabelMap(); err != nil {
		log.Fatalf("Unable to initialize trello connection: %s", err)
	}

	if len(config.LabelMap) == 0 {
		if len(config.LabelCardName) == 0 {
			log.Fatal("Must specify either 'trello_label_map' or 'trello_label_card_name'")
		}
		if err := t.loadListMap(); err != nil {
			log.Fatalf("Unable to initialize trello connection: %s", err)
		}
	}
}

func (t *Trello) loadBoard(boardName string) error {
	boards, err := t.client.SearchBoards(boardName, trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Unable to get board ID for board %s", boardName)
	}
	if len(boards) > 0 {
		for _, board := range boards {
			if board.Name == boardName {
				t.board = board
				return nil
			}
		}
	}
	return errors.New("Unable to find board ID for board \"" + boardName + "\"")
}

func (t *Trello) loadLabelMap() error {
	cards, err := t.client.SearchCards(t.config.LabelCardName, trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Unable to get label map for card id \"%s\"", t.config.LabelCardName)
	}

	for _, card := range cards {
		if card.Name == t.config.LabelCardName {
			for _, label := range card.Labels {
				if len(label.Name) > 0 {
					t.common.labelIDMap[label.Name] = label.ID
				}
			}
			return nil
		}
	}
	return errors.New("Could not find labels for card \"" + t.config.LabelCardName + "\"")
}

func (t *Trello) loadListMap() error {
	lists, err := t.board.GetLists(trello.Defaults())
	if err != nil {
		return errors.Wrapf(err, "Could not get lists for board \"%s\"", t.board.Name)
	}

	for _, list := range lists {
		t.common.listIDMap[list.Name] = list.ID
	}
	return nil
}

func (t *Trello) getLabelIDsForNames(labelNames []string) []string {
	idLabels := make([]string, len(labelNames))
	for idx, labelName := range labelNames {
		idLabels[idx] = t.common.labelIDMap[labelName]
	}
	return idLabels
}
