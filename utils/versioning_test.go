package utils_test

import (
	"testing"

	"github.com/nilock/tuido/utils"
)

// TestVersion sanity-checks the version reporting. The concrete version is
// injected at build time by GoReleaser via -ldflags (see .goreleaser.yaml),
// so for plain `go test` builds it is expected to report the "dev" fallback.
// The old test that compared a hardcoded const against the latest git tag is
// intentionally gone: the git tag is now the single source of version truth.
func TestVersion(t *testing.T) {
	if got := utils.Version(); got == "" {
		t.Errorf("Version() returned empty string")
	}
}
