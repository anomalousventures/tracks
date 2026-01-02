package database

import (
	"net/url"
	"strings"
)

func SanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid URL]"
	}
	if u.User == nil {
		return rawURL
	}

	var result strings.Builder
	result.WriteString(u.Scheme)
	result.WriteString("://")
	result.WriteString("****:****@")
	result.WriteString(u.Host)
	result.WriteString(u.Path)
	if u.RawQuery != "" {
		result.WriteString("?")
		result.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		result.WriteString("#")
		result.WriteString(u.Fragment)
	}
	return result.String()
}
