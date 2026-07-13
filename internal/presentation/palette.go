package presentation

import (
	"io"
	"os"

	"golang.org/x/term"
)

type Role int

const (
	ID Role = iota
	Important
	Completed
	Selected
	Success
	Warning
	Error
	Muted
	Key
	Logo
)

type Palette struct {
	Enabled bool
}

func ForWriter(w io.Writer) Palette {
	file, ok := w.(*os.File)
	return Palette{Enabled: ok && Allowed() && term.IsTerminal(int(file.Fd()))}
}

func Allowed() bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}
	return os.Getenv("TERM") != "dumb"
}

func (p Palette) Paint(role Role, text string) string {
	if !p.Enabled || text == "" {
		return text
	}
	return code(role) + text + "\x1b[0m"
}

func code(role Role) string {
	switch role {
	case ID:
		return "\x1b[36m"
	case Important:
		return "\x1b[1;33m"
	case Completed, Success:
		return "\x1b[32m"
	case Selected:
		return "\x1b[1;30;46m"
	case Warning:
		return "\x1b[33m"
	case Error:
		return "\x1b[31m"
	case Muted:
		return "\x1b[2m"
	case Key:
		return "\x1b[35m"
	case Logo:
		return "\x1b[1;35m"
	default:
		return ""
	}
}
