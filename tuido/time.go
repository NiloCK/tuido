package tuido

import (
	"regexp"
	"strconv"
	"time"
)

// expandDateShorthands takes an input string (a user supplied tuido)
// and expands its date shorthads via a regex matcher.
//
// Expansion Examples:
//  - "r1w" -> "#repeat=1w" (repeat weekly)
//  - "a5w" -> "#active=[datestring for 5 weeks from now]" (hide until 5 weeks)
//  - "e25m" -> "#estimate=25m" (estimate 25 minutes task time)
//  - "d7d" -> "#due=[datestring for 7 days from now]" (set a due date)
func expandDateShorthands(s string) string {
	return rex.ReplaceAllStringFunc(s, repl)
}

var rex regexp.Regexp = *regexp.MustCompile("[r,e,a,d][0-9]+[h,d,w,m,y,M]")

func repl(s string) string {
	ret := ""

	// returned directly, w/ shorthand durations
	switch s[0] {
	case 'r':
		return "#repeat=" + s[1:]
	case 'e':
		return "#estimate=" + s[1:]
	}

	// durations converted to dates, then returned
	switch s[0] {
	case 'd':
		ret += "#due="
	case 'a':
		ret += "#active="
	}

	t := toDate(s[1:])
	ret += t.Format("2006-01-02")

	return ret
}

// toDate parses duration shorthands like
//  - 2m (two minutes)
//  - 16h (16 hours)
//  - 3d (three days)
//  - 12w (twelve weeks)
//  - 1y (one year)
// into time.Time structs that far from now.
//
// [ ] #test #parsing
func toDate(dStr string) time.Time {
	// fmt.Println("toDate(" + dStr + ")")
	t := time.Now()

	num, err := strconv.Atoi(dStr[:len(dStr)-1])

	if err != nil {
		// fmt.Println("err: ", err)
		return time.Time{}
	}

	switch dStr[len(dStr)-1] {
	case 'm':
		t = t.Add(time.Minute * time.Duration(num))
	case 'h':
		t = t.Add(time.Hour * time.Duration(num))
	case 'd':
		t = t.AddDate(0, 0, num)
	case 'w':
		t = t.AddDate(0, 0, num*7)
	case 'M':
		t = t.AddDate(0, num, 0)
	case 'y':
		t = t.AddDate(num, 0, 0)
	}

	return t
}

// ToDuration parses duration shorthands like
//  - 16h (16 hours)
//  - 3d (three days)
//  - 12w (twelve weeks)
//  - 2m (two months)
//  - 1y (one year)
// into time.Durations structs of that duration.
//
// Note, 1m from now will producea different durations
// depending on the current month
//
// [ ] #test #parsing
func ToDuration(dStr string) *time.Duration {
	num, err := strconv.Atoi(dStr[:len(dStr)-1])

	var d time.Duration

	if err != nil {
		// fmt.Println("err: ", err)
		return nil
	}

	switch dStr[len(dStr)-1] {
	case 'h':
		d = time.Hour * time.Duration(num)
	case 'd':
		d = time.Hour * time.Duration(24*num)
	case 'w':
		d = time.Hour * time.Duration(24*7*num)
	case 'm':
		nextM := time.Now().AddDate(0, 1, 0)
		d = nextM.Sub(time.Now())
	case 'y':
		nextY := time.Now().AddDate(1, 0, 0)
		d = nextY.Sub(time.Now())
	}

	return &d
}
