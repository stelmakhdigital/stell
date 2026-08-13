package guardrails

import (
	"regexp"

	"github.com/budaev/stell/pkg/observability"
)

var (
	pemRe    = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)
	skLiveRe = regexp.MustCompile(`(?i)sk-(live|test)-[A-Za-z0-9]{16,}`)
	openaiRe = regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9]{20,}\b`)
	awsKeyRe = regexp.MustCompile(`(?i)\bAKIA[0-9A-Z]{16}\b`)
	githubRe = regexp.MustCompile(`(?i)\bghp_[A-Za-z0-9]{20,}\b`)
)

// OutputPolicy redacts secrets and can block leaks.
type OutputPolicy struct {
	DenySubstrings []string
}

// DefaultOutputPolicy redacts common secret patterns.
func DefaultOutputPolicy() *OutputPolicy {
	return &OutputPolicy{
		DenySubstrings: []string{
			"BEGIN RSA PRIVATE KEY",
			"BEGIN OPENSSH PRIVATE KEY",
		},
	}
}

// Filter redacts secrets. Deny if a private key dump remains obvious.
func (p *OutputPolicy) Filter(text string) (string, Decision) {
	if p == nil {
		return text, Decision{Action: Allow}
	}
	out := observability.Redact(text)
	out = pemRe.ReplaceAllString(out, "[REDACTED PRIVATE KEY]")
	out = skLiveRe.ReplaceAllString(out, "[REDACTED API KEY]")
	out = openaiRe.ReplaceAllString(out, "[REDACTED API KEY]")
	out = awsKeyRe.ReplaceAllString(out, "[REDACTED AWS KEY]")
	out = githubRe.ReplaceAllString(out, "[REDACTED GITHUB TOKEN]")

	if containsAnyFold(out, p.DenySubstrings) {
		return "отказ: вывод заблокирован из-за возможной утечки секрета", Decision{
			Action:  Deny,
			Message: "secret leak",
		}
	}
	if out != text {
		return out, Decision{Action: Modify, Rewrite: out}
	}
	return out, Decision{Action: Allow}
}
