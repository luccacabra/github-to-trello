package trello

type CardAction int

const (
	CREATE CardAction = iota
	UPDATE
	CLOSE
)

type Actions struct {
	Create struct {
		Lists  []string
		Labels []string
	}
	Update struct {
		Lists  []string
		Labels []string
	}
	Close struct {
		Lists  []string
		Labels []string
	}
}