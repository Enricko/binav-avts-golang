package helper

import (
	"html"
	"regexp"
	"strings"
)

func Protect(data string) string {
	// Menghapus karakter backslashes
	filter := strings.ReplaceAll(data, "\\", "")

	// Menghapus tag HTML dan karakter khusus
	filter = html.EscapeString(filter)

	// Menghapus karakter HTML entities
	re := regexp.MustCompile(`&(?:[a-zA-Z]+|#\d+);`)
	filter = re.ReplaceAllString(filter, "")

	return filter
}
