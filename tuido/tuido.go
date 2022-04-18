package tuido

import (
	"time"
)

type status string

const (
	Open     status = "open"
	Checked  status = "checked"
	Ongoing  status = "ongoing"
	Obsolete status = "obsolete"
)

type Item struct {
	// metadata

	file string
	line int

	// item data

	raw    string
	status status
	due    time.Time
	tags   []string
}

func (i Item) String() string {
	return i.raw
}

func New(
	file string,
	line int,
	raw string,
) Item {
	return Item{
		file:   file,
		line:   line,
		raw:    raw,
		due:    time.Now(), // todo
		status: Open,       // todo
	}
}
