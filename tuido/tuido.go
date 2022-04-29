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

func (i *Item) SetStatus(s status) error {
	// [ ] refactor file manipulation as own fcn (pkg)
	// approx: fInsert(file, lineNumber, expected, updated)

	f, err := os.OpenFile(i.file, os.O_RDWR, os.ModeExclusive)
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

	if lines[i.line] != i.raw {
		fmt.Printf("error finding todo: %s != %s", lines[i.line], i.raw)
		return fmt.Errorf("todo no longer in expected location...")
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		fmt.Printf("seek error: %s", err)
	}

	lines[i.line] = s.toString() + i.Text()

	for _, l := range lines[1:] {
		_, err := f.Write([]byte(l + "\n"))
		if err != nil {
			fmt.Printf("error writing to F: %s", err)
			return err
		}
	}

	i.raw = s.toString() + i.Text()

	return nil
}

func (i Item) String() string {
	return i.raw
}

func (i Item) Text() string {
	return i.raw[3:]
}

func (i Item) Tags() []string {
	return Tags(i.Text())
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
