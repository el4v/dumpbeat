package version

import (
	"fmt"
	"strings"
)

var (
	version   string
	buildTime string
	commit    string
)

// AsString formatted string for beautiful version
func AsString() string {
	bt := strings.Replace(buildTime, "_", " ", -1)
	return fmt.Sprintf("%s (Build Time: %s, Commit: %s)", version, bt, commit)
}
