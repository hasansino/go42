package tools

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
)

// nolint:unused
// monkeyTime patches time.Now() function to return predefined value
func monkeyTime(with time.Time) func() {
	p := gomonkey.ApplyFunc(time.Now, func() time.Time {
		return with
	})
	return p.Reset
}
