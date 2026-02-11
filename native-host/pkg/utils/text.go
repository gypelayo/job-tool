package utils

import "strings"

func CleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response)
}

func ExtractURL(text string) string {
	if idx := strings.Index(text, "URL:"); idx != -1 {
		urlLine := text[idx:]
		if endIdx := strings.Index(urlLine, "\n"); endIdx != -1 {
			urlLine = urlLine[:endIdx]
		}
		url := strings.TrimSpace(strings.TrimPrefix(urlLine, "URL:"))
		return url
	}
	return ""
}
