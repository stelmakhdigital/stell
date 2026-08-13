package eval

import (
	"strings"
	"unicode"
)

// DeterministicScore scores an answer without an LLM.
func DeterministicScore(c Case, answer string) Score {
	details := map[string]any{}
	ansLower := strings.ToLower(answer)

	if c.ExpectedRefusal {
		refused := containsAny(ansLower, []string{"не могу", "cannot", "won't", "refuse", "отказ", "запрещ"})
		details["refusal_detected"] = refused
		v := 0.0
		if refused {
			v = 1.0
		}
		return Score{Deterministic: v, Aggregate: v, Details: details}
	}

	keyHits := 0
	for _, kp := range c.ExpectedAnswer.KeyPoints {
		if kp == "" {
			continue
		}
		if strings.Contains(ansLower, strings.ToLower(kp)) || fuzzyContains(ansLower, strings.ToLower(kp)) {
			keyHits++
		}
	}
	keyTotal := len(c.ExpectedAnswer.KeyPoints)
	keyScore := 1.0
	if keyTotal > 0 {
		keyScore = float64(keyHits) / float64(keyTotal)
	}
	details["key_points_hit"] = keyHits
	details["key_points_total"] = keyTotal

	citeHits := 0
	for _, cite := range c.ExpectedAnswer.RequiredCitations {
		if cite != "" && strings.Contains(ansLower, strings.ToLower(cite)) {
			citeHits++
		}
	}
	citeTotal := len(c.ExpectedAnswer.RequiredCitations)
	citeScore := 1.0
	if citeTotal > 0 {
		citeScore = float64(citeHits) / float64(citeTotal)
	}
	details["citations_hit"] = citeHits
	details["citations_total"] = citeTotal

	forbiddenHit := false
	for _, f := range c.ExpectedAnswer.Forbidden {
		if f != "" && strings.Contains(ansLower, strings.ToLower(f)) {
			forbiddenHit = true
			break
		}
	}
	details["forbidden_hit"] = forbiddenHit
	forbidScore := 1.0
	if forbiddenHit {
		forbidScore = 0.0
	}

	// Weighted: key points dominate, then citations, then forbidden gate.
	det := 0.7*keyScore + 0.2*citeScore + 0.1*forbidScore
	if forbiddenHit {
		det *= 0.5
	}
	details["key_score"] = keyScore
	details["cite_score"] = citeScore
	details["forbid_score"] = forbidScore

	return Score{Deterministic: det, Aggregate: det, Details: details}
}

func containsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

// fuzzyContains checks if most significant tokens of needle appear in haystack.
func fuzzyContains(haystack, needle string) bool {
	tokens := strings.FieldsFunc(needle, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == '.' || r == ';'
	})
	if len(tokens) == 0 {
		return false
	}
	hits := 0
	for _, t := range tokens {
		if len(t) < 3 {
			continue
		}
		if strings.Contains(haystack, t) {
			hits++
		}
	}
	need := len(tokens) / 2
	if need < 1 {
		need = 1
	}
	return hits >= need
}
