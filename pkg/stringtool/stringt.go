package stringtool

import (
	"strings"
)

// Empty checks if string is empty or consists only with spaces
func Empty(s string) bool {
	return strings.TrimSpace(s) == ""
}
