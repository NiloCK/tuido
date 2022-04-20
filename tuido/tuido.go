package tuido

import (
	"time"
)

type status string

const (
	Open     status = "open"
	Ongoing  status = "ongoing"
	Checked  status = "checked"
	Obsolete status = "obsolete"
	unknown  status = "unknown"
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

// Status returns the status of the item. One of:
//  - open (ie, noted but not begun)
//  - ongoing (ie, in progress)
//  - checked (ie, completed)
//  - obsolete (ie, no longer necessary)
func (i Item) Satus() status {

	if i.raw[1] == ' ' {
		return Open
	}
	if i.raw[1] == '@' {
		return Ongoing
	}
	if i.raw[1] == '~' {
		return Obsolete
	}
	if i.raw[1] == 'x' || i.raw[1] == 'X' {
		return Checked
	}

	return unknown
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
