package storage

type ActionType string

const (
	CREATE ActionType = "CREATE"
	UPDATE            = "UPDATE"
	CLOSE             = "CLOSE"
)

type Issue struct {
	Id               int64  `db:"primarykey, autoincrement"`
	Body             string `db:"body"`
	IssueId          string `db:"issue_id"`
	Number           int64  `db:"number"`
	Repository       string `db:"repository"`
	Title            string `db:"title"`
	URL              string `db:"url"`
	UserRelationship string `db:"user_relationship"`

	Comments []*Comment `db:"-"`
}

type Comment struct {
	Id      int64  `db:"primarykey, autoincrement"`
	IssueId string `db:"issue_id"`
	Body    string `db:"body"`
}

type Card struct {
	Id           int64  `db:"primarykey, autoincrement"`
	IssueId      int64  `db:"issue_id"`
	Title        string `db:"title"`
	Text         string `db:"text"`
	TrelloCardId string `db:"trello_card_id"`
	ListId       string `db:"list_id"`
	LabelIds     string `db:"label_ids"`
}
