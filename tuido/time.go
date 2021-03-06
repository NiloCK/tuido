package tuido

import (
	"regexp"
	"strconv"
	"time"
)

func expandDateShorthands(s string) string {
	return rex.ReplaceAllStringFunc(s, repl)
}

var rex regexp.Regexp = *regexp.MustCompile("[r,o,d][0-9]+[h,d,w,m,y]")

func repl(s string) string {
	ret := ""

	switch s[0] {
	case 'r':
		return "#repeat=" + s[1:]
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
//  - 16h (16 hours)
//  - 3d (three days)
//  - 12w (twelve weeks)
//  - 2m (two months)
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
	case 'h':
		t = t.Add(time.Hour * time.Duration(num))
	case 'd':
		t = t.AddDate(0, 0, num)
	case 'w':
		t = t.AddDate(0, 0, num*7)
	case 'm':
		t = t.AddDate(0, num, 0)
	case 'y':
		t = t.AddDate(num, 0, 0)
	}

	return t
}
