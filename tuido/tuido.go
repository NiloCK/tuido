package tuido

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/nilock/tuido/utils"
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

func (s status) String() string {
	switch s {

	case Open:
		return "[ ]"
	case Ongoing:
		return "[@]"
	case Checked:
		return "[x]" // [✔] [✓] ?
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

	raw string
}

func (i *Item) Location() string {
	if i == nil {
		return ""
	}

	return fmt.Sprintf("%s:%d", i.file, i.line)
}

// File returns the path of the source file the item was read from.
func (i Item) File() string { return i.file }

// Line returns the 1-indexed line number of the item in its source file.
func (i Item) Line() int { return i.line }

// Status returns the status of the item. One of:
//   - open (ie, noted but not begun)
//   - ongoing (ie, in progress)
//   - checked (ie, completed)
//   - obsolete (ie, no longer necessary)
func (i Item) Satus() status {
	return strToStatus(i.trimmed())
}

// SetStatus writes the updated status to the item's file
// on disk and updates the status of the in-memory item.
//
// If the disk write fails, the in-memory update is abandoned.
func (i *Item) SetStatus(s status) error {
	if i == nil {
		return fmt.Errorf("item is nil - cannot update status")
	}
	if s == Checked {
		repeat := i.Repeat()
		if repeat != nil {
			i.setTag(Tag{
				name:  "active",
				value: time.Now().Add(*repeat).Format("2006-01-02"),
			})
			i.setTag(Tag{
				name:  "lastDone",
				value: time.Now().Format("2006-01-02"),
			})

			// prevent fall-through - we no longer want this to be
			// marked "done". It's only been pushed into the future
			return nil
		}
		// [ ] add #completed=[currentDate] if s == Checked?
	}

	newRaw := i.scrap() + s.String() + " " + i.Text()

	err := fileInsert(i.file, i.line, i.raw, newRaw)
	if err != nil {
		return err
	}

	i.raw = newRaw
	return nil
}

func (i *Item) IncrementTimeSpent(seconds int) {
	if i == nil {
		return
	}
	previouslySpent := 0.0

	for _, t := range i.Tags() {
		if t.name == "spent" {
			var err error

			previouslySpent, err = strconv.ParseFloat(t.value, 64)
			if err != nil {
				return
			}
			break
		}
	}
	asMinutes := float64(seconds) / 60

	asStr := fmt.Sprintf("%.2f", previouslySpent+asMinutes)

	i.setTag(Tag{
		name:  "spent",
		value: asStr,
	})
}

// SetText writes the updated text to the item's file
// on disk and updates the text of the in-memory item.
//
// If the disk write fails, the in-memory update is abandoned.
func (i *Item) SetText(t string) error {
	t = expandDateShorthands(t)

	if i == nil {
		return fmt.Errorf("item is nil - cannot update text")
	}

	newRaw := i.scrap() + i.Satus().String() + " " + t
	err := fileInsert(i.file, i.line, i.raw, newRaw)
	if err != nil {
		return err
	}
	i.raw = newRaw
	return nil
}

// GetContext reads and returns some surrounding text from the item's source file.
//
// The returned integer is the line number of the item's text inside the returned context.
func (i *Item) GetContext(height int) (string, int) {
	fileBytes, err := exec.Command("cat", i.file).CombinedOutput()

	if err != nil {
		return err.Error(), 0
	}

	// colorize the file contents
	buf := new(bytes.Buffer)
	quick.Highlight(buf, string(fileBytes), lexers.Match(i.file).Config().Name, utils.GetTerminalColorSupport(), "monokai")

	lines := strings.Split(string(buf.Bytes()), "\n")
	lines = append([]string{""}, lines...)

	first := i.line - (height / 2) + 1
	if first < 0 {
		first = 0
	}
	last := i.line + (height / 2) - 1
	if last > len(lines)-1 {
		last = len(lines) - 1
	}

	preItemLines := strings.Join(lines[first:i.line], "\n")

	item := lines[i.line]

	postItemLines := ""
	if (i.line + 1) < last {
		postItemLines = strings.Join(lines[i.line+1:last], "\n")
	}

	item = lipgloss.NewStyle().Bold(true).Italic(true).Render(item) // not working - clobbered by styles from chroma

	return preItemLines + "\n" + item + "\n" + postItemLines, i.line - first
}

func (i *Item) Snooze() error {
	if i == nil {
		return fmt.Errorf("item is nil - cannot snooze")
	}

	count := i.snoozeCount()
	count++

	// i.set("active", time.Now() + fib(count) days)
	i.setTag(Tag{
		"active",
		time.Now().Add(time.Hour * time.Duration(24*fib(count))).Format("2006-01-02"),
	})
	// i.set("zzz", count)
	return i.setTag(Tag{"zzz", fmt.Sprint(count)})
}

// Escalate increases the "importance" of an item by prefixing it
// with an exclamation point.
func (i *Item) Escalate() error {
	txt := i.Text()
	if len(txt) == 0 {
		return i.SetText("!")
	}

	if txt[0] == '!' {
		return i.SetText("!" + txt)
	}

	return i.SetText("! " + txt)
}

// Deescalate decreases the "importance" of an item by removing
// a leading exclamation mark.
//
// NB: deescalate is not up to [x]it spec wrt items prefixed with
//
//	both periods and exclamations, and will fail to deescalate,
//	eg, "..!!! do this"
//
// [ ] make [x]it spec compliant
// [ ] wants unit tests
func (i *Item) Deescalate() error {
	txt := i.Text()

	if strings.HasPrefix(txt, "!!") {
		return i.SetText(txt[1:])
	}

	if strings.HasPrefix(txt, "! ") {
		return i.SetText(txt[2:])
	}

	return fmt.Errorf("item already has priority 0")
}

func fib(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	if n == 2 {
		return 2
	}

	return fib(n-1) + fib(n-2)
}

func (i *Item) snoozeCount() int {
	for _, tag := range i.Tags() {
		if tag.name == "zzz" {
			count, _ := strconv.Atoi(tag.value)
			return count
		}
	}

	return 0
}

// setTag replaces the value of an existing tag, or appends a new tag.
func (i *Item) setTag(t Tag) error {
	// replace existing value, if exists
	for _, tag := range i.Tags() {
		if tag.name == t.name {
			txt := strings.Replace(i.Text(), tag.String(), t.String(), 1)

			return i.SetText(txt)
		}
	}

	// else, append new tag
	txt := i.Text() + " #" + t.String()
	return i.SetText(txt)
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
//   - [x] this one
//
// String() returns "[x] this one"
func (i Item) String() string {
	trimmed := i.trimmed()

	// provide vizual for items just snoozed via a keypress.
	if i.Satus() == Open && !i.Active() {
		return "[z] " + trimmed[4:]
	}

	return trimmed
}

// Text returns the item's body text. EG, for item
//   - [x] this one is done
//
// the Text() is "this one is done"
func (i Item) Text() string {
	trimmed := i.trimmed()
	if len(trimmed) < 4 {
		return ""
	}
	return i.trimmed()[4:]
}

func (i Item) Tags() []Tag {
	return Tags(i.Text())
}

// SystemTags returns the ## system tags carried by the item.
func (i Item) SystemTags() []SystemTag {
	return SystemTags(i.Text())
}

// IsControl reports whether the item is a file control item, ie carries the
// ##file system tag. A control item stands in for its whole file in the
// aggregate view: its siblings are collapsed behind it.
func (i Item) IsControl() bool {
	for _, t := range i.SystemTags() {
		if t.name == SysFile {
			return true
		}
	}
	return false
}

// Active returns the "active" status for snoozed items.
// Items with `active` tags later than the current date will not
// be shown in the regular view. Defaults to true.
func (i Item) Active() bool {
	for _, t := range i.Tags() {
		if t.name == "active" {
			return parseTagDate(t).Before(time.Now())
		}
	}
	return true
}

// Importance returns the number of leading '!'s in
// the item's Text.
func (i Item) Importance() int {
	count := 0
	txt := i.Text()

	for _, ch := range txt {
		if ch == '!' {
			count++
		} else if ch == '.' {
		} else {
			return count
		}
	}

	return count
}

func (i Item) Created() *time.Time {
	for _, t := range i.Tags() {
		if t.name == "created" {
			return parseTagDate(t)
		}
	}
	for c := range i.file {
		l := len("2006-01-02")
		if c+l < len(i.file) {
			sStr := i.file[c : c+l]
			if d, err := time.Parse("2006-01-02", sStr); err == nil {
				return &d
			}
		}
	}
	return nil
}

func (i Item) Repeat() *time.Duration {
	for _, t := range i.Tags() {
		if t.name == "repeat" {
			return ToDuration(t.value)
		}
	}

	return nil
}

func (i Item) Due() *time.Time {
	for _, t := range i.Tags() {
		if t.name == "due" { //  [ ]!  make a const enum somewhere - appTags or something
			return parseTagDate(t)
		}
	}
	return nil
}

func parseTagDate(t Tag) *time.Time {
	ret, err := time.Parse("2006-01-02", t.value)
	if err != nil {
		// panic(err)
	}

	return &ret
}

// IsTuido inspects a raw string for parsibility into a tuido item.
// It relaxes the [x]it spec in the following ways:
//   - leading whitespace is allowed
//   - markdown style bulleted items are allowed
//   - golang inline "//" comments are parsed for items
//
// [ ] unit #test this w/ a bunch of expected passes & failures
// [ ] #maybe allow numbered md lists (1. [ ] ...)
// [ ] #maybe include a language map for code-comment parsing. ie, {".rb": "#", ".go": "//"}
// [ ] ! #maybe require a file extension for this fcn. Allows for PL specific rules, as well as md
func IsTuido(raw string) bool {
	trimmed := trim(raw)

	for _, status := range statuses {
		if strings.HasPrefix(trimmed, status.String()) {
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
	// [ ] !!!!!!! replace this magic # w/ better named ctors
	if line < 0 { // this is a new item authored in-tui
		newItemRaw := "[ ] "

		// append new blank todo to `file`
		fInfo, err := os.Stat(file)

		if err != nil {
			fmt.Printf("error checking if %s is a directory: %s\n", fInfo, err)
		}

		if fInfo.IsDir() {
			file = filepath.Join(file, time.Now().Format("2006-01-02")+".xit") // xit, md, tbd
		}

		f, err := os.OpenFile(file, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			return Item{}
		}
		defer f.Close()
		f.WriteString(newItemRaw)

		// get the line # of the new item
		f.Seek(0, 0) // reset to beginning of f
		scanner := bufio.NewScanner(f)
		newItemLine := 0
		for scanner.Scan() {
			newItemLine++
		}

		return Item{
			file: file,
			line: newItemLine,
			raw:  newItemRaw,
		}
	}
	return Item{
		file: file,
		line: line,
		raw:  raw,
	}
}

func Tags(s string) []Tag {
	tags := []Tag{}
	split := strings.Split(s, " ")

	for _, token := range split {
		// ## denotes a system tag; it is deliberately excluded from the
		// user tag space (no coloring, no fuzzy filtering). See SystemTags.
		if strings.HasPrefix(token, "##") {
			continue
		}
		if strings.HasPrefix(token, "#") && len(token) > 1 {
			tags = append(tags, NewTag(token[1:]))
		}
	}

	return tags
}

// SystemTag is a tuido-reserved, double-hash (##) tag. System tags are parsed
// and acted upon by tuido itself, and are kept separate from the user-facing
// Tag space: they are not colorized, not fuzzy-filterable, and do not appear
// in Item.Tags().
type SystemTag struct {
	name  string
	value string
}

func (s SystemTag) Name() string  { return s.name }
func (s SystemTag) Value() string { return s.value }
func (s SystemTag) String() string {
	if s.value != "" {
		return fmt.Sprintf("%s=%s", s.name, s.value)
	}
	return s.name
}

// Reserved system tag names.
const (
	// SysFile marks a file's control/summary item. The control item stands
	// in for the whole file in the aggregate view.
	SysFile = "file"
)

// SystemTags parses the ##-prefixed system tags out of s.
func SystemTags(s string) []SystemTag {
	tags := []SystemTag{}
	for _, token := range strings.Split(s, " ") {
		if strings.HasPrefix(token, "##") && len(token) > 2 {
			tags = append(tags, newSystemTag(token[2:]))
		}
	}
	return tags
}

// newSystemTag splits a "name=value" (no ## prefix) into a SystemTag.
func newSystemTag(s string) SystemTag {
	split := strings.SplitN(s, "=", 2)
	if len(split) == 2 {
		return SystemTag{name: split[0], value: split[1]}
	}
	return SystemTag{name: s}
}

// CollapseFileScoped returns items with file-controlled groups collapsed to
// their control item. For any file that contains a control item (##file),
// only the control item is retained and its siblings are dropped. Files with
// no control item are returned unchanged. Input order is preserved.
func CollapseFileScoped(items []*Item) []*Item {
	controlled := map[string]bool{}
	for _, it := range items {
		if it.IsControl() {
			controlled[it.file] = true
		}
	}
	if len(controlled) == 0 {
		return items
	}

	out := make([]*Item, 0, len(items))
	for _, it := range items {
		if controlled[it.file] && !it.IsControl() {
			continue // sibling collapsed behind its control item
		}
		out = append(out, it)
	}
	return out
}

// ChildStats returns the (remaining, total) counts of a control item's
// children - the non-control items sharing its file. remaining counts open
// and ongoing children. Because one file carries at most one control item,
// children are identified as same-file items that are not themselves control
// items (rather than by pointer identity, so a copy of the control works).
func ChildStats(control *Item, all []*Item) (remaining, total int) {
	if control == nil {
		return 0, 0
	}
	for _, it := range all {
		if it.file != control.file || it.IsControl() {
			continue
		}
		total++
		if s := it.Satus(); s == Open || s == Ongoing {
			remaining++
		}
	}
	return remaining, total
}

type Tag struct {
	name  string
	value string
}

func (t Tag) Name() string {
	return t.name
}
func (t Tag) String() string {
	if t.value != "" {
		return fmt.Sprintf("%s=%s", t.name, t.value)
	}

	return t.name
}

// NewTag splits a string token "#name=value" into a Tag struct.
func NewTag(s string) Tag {
	if strings.HasPrefix(s, "#") && len(s) > 1 {
		s = s[1:]
	}

	split := strings.Split(s, "=")
	name := split[0]

	// recombine other split parts into value
	value := strings.Join(split[1:], "=")

	if len(split) >= 2 {
		return Tag{
			name:  name,
			value: value,
		}
	}

	return Tag{
		name: s,
	}
}
