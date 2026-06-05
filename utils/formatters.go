package utils

import (
	"strings"

	"discord-tmdb-bot/tmdb"
)

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func FormatCompanies(companies []tmdb.Company, max int) string {
	if len(companies) == 0 {
		return "N/A"
	}
	names := make([]string, 0, max)
	for i, c := range companies {
		if i >= max {
			break
		}
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}
