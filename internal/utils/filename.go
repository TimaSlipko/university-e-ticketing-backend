package utils

import (
	"regexp"
	"strings"
)

// SanitizeFilename removes or replaces characters that are not safe for filenames
func SanitizeFilename(filename string) string {
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove special characters except dots, hyphens, and underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = reg.ReplaceAllString(filename, "")

	// Limit length to 100 characters
	if len(filename) > 100 {
		// Keep the extension
		if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
			ext := filename[dotIndex:]
			name := filename[:dotIndex]
			if len(name) > 100-len(ext) {
				name = name[:100-len(ext)]
			}
			filename = name + ext
		} else {
			filename = filename[:100]
		}
	}

	return filename
}
