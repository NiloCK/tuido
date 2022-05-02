package tuido

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

var statuses []status = []status{Open, Ongoing, Checked, Obsolete}

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
	return strToStatus(i.trimmed())
}

// SetStatus writes the updated status to the item's file
// on disk and updates the status of the in-memory item.
//
// If the disk write fails, the in-memory update is abandoned.
func (i *Item) SetStatus(s status) error {
	newRaw := i.scrap() + s.toString() + i.Text()

	err := fileInsert(i.file, i.line, i.raw, newRaw)
	if err != nil {
		return err
	}

	i.raw = newRaw
	return nil
}

// fileInsert replaces the lineNumberth line of file with updated, as long
// it finds that the current contents of that line are as expected.
func fileInsert(file string, lineNumber int, expected string, updated string) error {
	f, err := os.OpenFile(file, os.O_RDWR, os.ModeExclusive)
	defer f.Close()

	if err != nil {
		fmt.Printf("error opening file for setStatus: %s", err)
		return err
	}
	scanner := bufio.NewScanner(f)

	lines := []string{""} // blank line to offset

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if lines[lineNumber] != expected {
		fmt.Printf("error finding todo: %s != %s", lines[lineNumber], expected)
		return fmt.Errorf("todo no longer in expected location, or changed on disk...")
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		fmt.Printf("seek error: %s", err)
	}

	lines[lineNumber] = updated

	for _, l := range lines[1:] {
		_, err := f.Write([]byte(l + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

// String returns the item status box plus body text. EG, for the item
//  - [x] this one
//
// String() returns "[x] this one"
func (i Item) String() string {
	return i.trimmed()
}

// Text returns the item's body text. EG, for item
//  - [x] this one is done
//
// the Text() is "this one is done"
func (i Item) Text() string {
	return i.trimmed()[3:]
}

func (i Item) Tags() []string {
	return Tags(i.Text())
}

// IsTuido inspects a raw string for parsibility into a tuido item.
// It relaxes the [x]it spec in the following ways:
//  - leading whitespace is allowed
//  - markdown style bulleted items are allowed
//  - golang inline "//" comments are parsed for items
//
// [ ] unit #test this w/ a bunch of expected passes & failures
// [ ] #maybe allow numbered md lists (1. [ ] ...)
// [ ] #maybe include a language map for code-comment parsing. ie, {".rb": "#", ".go": "//"}
// [ ] #maybe require a file extension for this fcn. Allows for PL specific rules, as well as md
func IsTuido(raw string) bool {
	trimmed := trim(raw)

	for _, status := range statuses {
		if strings.HasPrefix(trimmed, status.toString()) {
			return true
		}
	}

	return false
}

// trim left-prepares a string for tuido item parsing
//
// [ ] #test w/ expected in-outs
func trim(raw string) string {
	// remove leading whitespace & markdown bullet list identifiers.
	trimmed := strings.TrimLeft(raw, " \t")
	if strings.HasPrefix(trimmed, "- ") {
		trimmed = trimmed[2:]
	}

	// remove non-comment content from go (c, java, js, ts, etc)
	// style inlne commented lines
	if strings.Contains(trimmed, "// ") {
		split := strings.Split(trimmed, "// ")
		trimmed = strings.Join(split[1:], "// ") // only the leading instance begins a comment
	}

	return trimmed
}

func (i Item) Line() int { return i.line }

func (i Item) File() string { return i.file }

func (i Item) trimmed() string {
	return trim(i.raw)
}

func (i Item) scrap() string {
	return strings.Replace(i.raw, i.trimmed(), "", 1)
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

func Tags(s string) []string {
	tags := []string{}
	split := strings.Split(s, " ")

	for _, token := range split {
		if strings.HasPrefix(token, "#") && len(token) > 1 {
			tags = append(tags, token[1:])
		}
	}

	return tags
}
