package observability

import (
	"regexp"
	"strings"
)

var (
	bearerRe = regexp.MustCompile(`(?i)(Bearer\s+)[A-Za-z0-9\-._~+/]+=*`)
	apiKeyRe = regexp.MustCompile(`(?i)(api[_-]?key\s*[=:]\s*)([^\s&'"<>]+)`)
	tokenRe  = regexp.MustCompile(`(?i)(token\s*[=:]\s*)([^\s&'"<>]+)`)
)

// Redact masks common secrets in text.
func Redact(s string) string {
	if s == "" {
		return s
	}
	out := bearerRe.ReplaceAllString(s, "${1}***")
	out = apiKeyRe.ReplaceAllString(out, "${1}***")
	out = tokenRe.ReplaceAllString(out, "${1}***")
	return out
}

// RedactMap returns a shallow copy with string values redacted.
func RedactMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case string:
			if strings.Contains(strings.ToLower(k), "key") || strings.Contains(strings.ToLower(k), "token") || strings.Contains(strings.ToLower(k), "secret") {
				out[k] = "***"
			} else {
				out[k] = Redact(t)
			}
		default:
			out[k] = v
		}
	}
	return out
}
