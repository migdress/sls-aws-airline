package common

import (
	"strings"
)

func TrimLines(s string) string {
	trimmed := strings.ReplaceAll(s, "\n", "")
	trimmed = strings.ReplaceAll(trimmed, "\t", "")
	trimmed = strings.TrimSpace(trimmed)
	return trimmed
}
