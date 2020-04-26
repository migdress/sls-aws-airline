package internal

import (
	"regexp"
	"strings"
)

func TrimLines(s string) string {
	re := regexp.MustCompile(": +")
	trimmed := string(re.ReplaceAll([]byte(s), []byte(":")))
	trimmed = strings.ReplaceAll(trimmed, "\n", "")
	trimmed = strings.ReplaceAll(trimmed, "\t", "")
	trimmed = strings.TrimSpace(trimmed)
	return trimmed
}
