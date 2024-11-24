package internal

import "strings"

func TrimExtension(s string) string {
	idx := strings.LastIndex(s, ".pdf")

	// No such file extension exists
	if idx == -1 {
		return s
	}

	return s[:idx]
}
