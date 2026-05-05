package nats

import "strings"

func mergeStreamSubjects(current, required []string) ([]string, bool) {
	merged := make([]string, 0, len(current)+len(required))
	for _, subject := range current {
		if subject == "" || hasExactSubject(merged, subject) {
			continue
		}
		merged = append(merged, subject)
	}

	changed := len(merged) != len(current)
	for _, subject := range required {
		if subject == "" || streamSubjectsCover(merged, subject) {
			continue
		}
		merged = append(merged, subject)
		changed = true
	}

	return merged, changed
}

func streamSubjectsCover(patterns []string, subject string) bool {
	for _, pattern := range patterns {
		if subjectPatternCovers(pattern, subject) {
			return true
		}
	}
	return false
}

func hasExactSubject(subjects []string, subject string) bool {
	for _, existing := range subjects {
		if existing == subject {
			return true
		}
	}
	return false
}

func subjectPatternCovers(pattern, subject string) bool {
	if pattern == "" || subject == "" {
		return false
	}

	patternTokens := strings.Split(pattern, ".")
	subjectTokens := strings.Split(subject, ".")

	for i, token := range patternTokens {
		if token == ">" {
			return i == len(patternTokens)-1 && i < len(subjectTokens)
		}
		if i >= len(subjectTokens) {
			return false
		}
		if token == "*" {
			continue
		}
		if token != subjectTokens[i] {
			return false
		}
	}

	return len(patternTokens) == len(subjectTokens)
}
