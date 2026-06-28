package output

import (
	"os"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

type Mode int

const (
	ModeHuman Mode = iota
	ModePlain
	ModeJSON
)

type Formatter interface {
	FormatTicket(t ticket.Ticket) string
	FormatTicketList(tickets []ticket.Ticket) string
	FormatNext(t *ticket.Ticket) string
	FormatError(err error) string
}

func New(mode Mode) Formatter {
	switch mode {
	case ModeJSON:
		return &jsonFormatter{}
	case ModePlain:
		return &plainFormatter{}
	default:
		return &humanFormatter{}
	}
}

func DetectMode(plain, json bool) Mode {
	if json {
		return ModeJSON
	}
	if plain {
		return ModePlain
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return ModePlain
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return ModePlain
	}
	return ModeHuman
}
