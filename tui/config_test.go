package tui

import (
	"reflect"
	"testing"
)

// TestApplyConfig_allFieldsPropagated has two jobs:
//
//  1. It verifies that every field in the config struct is handled by applyConfig.
//     If a field is added to config and applyConfig is not updated, reflect.DeepEqual
//     will catch the mismatch.
//
//  2. It verifies that this test itself covers every field. If a field is added to
//     config but not set in `full` below, IsZero() will fail the test before applyConfig
//     is even called — so the test cannot silently pass with incomplete coverage.
func TestApplyConfig_allFieldsPropagated(t *testing.T) {
	full := config{
		extensions:        []string{"xit"},
		writeto:           "/tmp/test-writeto",
		frictionThreshold: 99,
		exclude:           []string{"vendor"},
	}

	// Guard: ensure this test sets every field to a non-zero value.
	// If a new field is added to config without updating `full`, this fails.
	v := reflect.ValueOf(full)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			t.Errorf("field %q is zero in test config — set it to a non-zero value when adding config fields",
				v.Type().Field(i).Name)
		}
	}

	saved := runConfig
	defer func() { runConfig = saved }()
	runConfig = config{}

	applyConfig(&full)

	if !reflect.DeepEqual(runConfig, full) {
		t.Errorf("applyConfig did not propagate all fields\ngot:  %+v\nwant: %+v", runConfig, full)
	}
}
