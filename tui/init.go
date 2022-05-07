package tui

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix()) // a fresh set of tag colors on each run. Spice of life.
}
