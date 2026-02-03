package replacer

import (
	"regexp"
	"strings"
)

// Matcher defines the interface for pattern matching
type Matcher interface {
	// FindAllMatches returns all match positions in the text
	FindAllMatches(text string) []Match
	// Replace performs the replacement on the text
	Replace(text string, replacement string) string
}

// Match represents a single pattern match
type Match struct {
	Start int
	End   int
	Text  string
}

// LiteralMatcher performs exact string matching
type LiteralMatcher struct {
	pattern       string
	caseSensitive bool
}

// NewLiteralMatcher creates a new literal matcher
func NewLiteralMatcher(pattern string, caseSensitive bool) *LiteralMatcher {
	return &LiteralMatcher{
		pattern:       pattern,
		caseSensitive: caseSensitive,
	}
}

// FindAllMatches finds all literal matches
func (m *LiteralMatcher) FindAllMatches(text string) []Match {
	var matches []Match
	searchText := text
	searchPattern := m.pattern

	if !m.caseSensitive {
		searchText = strings.ToLower(text)
		searchPattern = strings.ToLower(m.pattern)
	}

	start := 0
	for {
		idx := strings.Index(searchText[start:], searchPattern)
		if idx == -1 {
			break
		}

		actualStart := start + idx
		actualEnd := actualStart + len(m.pattern)

		matches = append(matches, Match{
			Start: actualStart,
			End:   actualEnd,
			Text:  text[actualStart:actualEnd],
		})

		start = actualEnd
	}

	return matches
}

// Replace performs literal string replacement
func (m *LiteralMatcher) Replace(text string, replacement string) string {
	if m.caseSensitive {
		return strings.ReplaceAll(text, m.pattern, replacement)
	}

	// Case-insensitive replacement is more complex
	result := text
	matches := m.FindAllMatches(text)

	// Replace from end to start to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		result = result[:match.Start] + replacement + result[match.End:]
	}

	return result
}

// RegexMatcher performs regex pattern matching
type RegexMatcher struct {
	regex *regexp.Regexp
}

// NewRegexMatcher creates a new regex matcher
func NewRegexMatcher(pattern string, caseSensitive bool) (*RegexMatcher, error) {
	// Add multiline flag to make . match newlines
	fullPattern := "(?s)" + pattern

	if !caseSensitive {
		fullPattern = "(?i)" + fullPattern
	}

	re, err := regexp.Compile(fullPattern)
	if err != nil {
		return nil, err
	}

	return &RegexMatcher{
		regex: re,
	}, nil
}

// FindAllMatches finds all regex matches
func (m *RegexMatcher) FindAllMatches(text string) []Match {
	var matches []Match

	indices := m.regex.FindAllStringIndex(text, -1)
	for _, idx := range indices {
		matches = append(matches, Match{
			Start: idx[0],
			End:   idx[1],
			Text:  text[idx[0]:idx[1]],
		})
	}

	return matches
}

// Replace performs regex replacement
func (m *RegexMatcher) Replace(text string, replacement string) string {
	return m.regex.ReplaceAllString(text, replacement)
}

// NewMatcher creates a matcher based on the mode
func NewMatcher(pattern string, isRegex bool, caseSensitive bool) (Matcher, error) {
	if isRegex {
		return NewRegexMatcher(pattern, caseSensitive)
	}
	return NewLiteralMatcher(pattern, caseSensitive), nil
}
