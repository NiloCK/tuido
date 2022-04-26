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

func (s status) toString() string {
	switch s {

	case Open:
		return "[ ]"
	case Ongoing:
		return "[@]"
	case Checked:
		return "[x]"
	case Obsolete:
		return "[~]"
	case unknown:
		return "[?]"
	default:
		return ""
	}
}
func strToStatus(s string) status {
	s = s[:3]

	if s == "[ ]" {
		return Open
	}
	if s == "[@]" {
		return Ongoing
	}
	if s == "[x]" || s == "[X]" {
		return Checked
	}
	if s == "[~]" {
		return Obsolete
	}
	return unknown
}

type Item struct {
	// metadata

	file string
	line int

	// item data

	raw  string
	due  time.Time
	tags []string
}

// Status returns the status of the item. One of:
//  - open (ie, noted but not begun)
//  - ongoing (ie, in progress)
//  - checked (ie, completed)
//  - obsolete (ie, no longer necessary)
func (i Item) Satus() status {
	return strToStatus(i.raw)
}


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
		file: file,
		line: line,
		raw:  raw,
		due:  time.Now(), // todo
	}
}
