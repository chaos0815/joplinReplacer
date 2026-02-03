package replacer

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	contextLines = 3 // Number of lines to show before/after match
)

// PrintPreview prints a preview of all matches
func PrintPreview(matchedNotes []NoteMatch, replacement string) {
	if len(matchedNotes) == 0 {
		fmt.Println("No matches found")
		return
	}

	cyan := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)

	for _, noteMatch := range matchedNotes {
		cyan.Printf("\n=== Note: %s ===\n", noteMatch.Note.Title)
		yellow.Printf("ID: %s\n", noteMatch.Note.ID)
		fmt.Printf("Matches: %d\n\n", len(noteMatch.Matches))

		// Show each match with context
		for i, match := range noteMatch.Matches {
			fmt.Printf("Match %d:\n", i+1)

			// Get context around the match
			before, matchText, after := getContext(noteMatch.Note.Body, match)

			// Print context before
			if before != "" {
				fmt.Print(before)
			}

			// Print the match in red
			red.Print(matchText)

			// Print what it will be replaced with in green
			fmt.Print(" → ")
			green.Print(replacement)
			fmt.Println()

			// Print context after
			if after != "" {
				fmt.Print(after)
			}

			fmt.Println()
		}
	}
}

// getContext extracts text around a match for preview
func getContext(text string, match Match) (before, matchText, after string) {
	lines := strings.Split(text, "\n")

	// Find which line contains the match
	currentPos := 0
	matchLine := 0
	matchLineStart := 0

	for i, line := range lines {
		lineLen := len(line) + 1 // +1 for newline
		if currentPos <= match.Start && match.Start < currentPos+lineLen {
			matchLine = i
			matchLineStart = currentPos
			break
		}
		currentPos += lineLen
	}

	// Calculate context range
	startLine := matchLine - contextLines
	if startLine < 0 {
		startLine = 0
	}

	endLine := matchLine + contextLines + 1
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Build context before match
	if startLine < matchLine {
		before = strings.Join(lines[startLine:matchLine], "\n")
		if before != "" {
			before += "\n"
		}
	}

	// Get the actual match text
	matchText = match.Text

	// Handle multiline matches
	matchEndPos := match.End
	currentPos = matchLineStart
	matchEndLine := matchLine

	for i := matchLine; i < len(lines); i++ {
		lineLen := len(lines[i]) + 1
		if matchEndPos <= currentPos+lineLen {
			matchEndLine = i
			break
		}
		currentPos += lineLen
	}

	// Build context after match
	if matchEndLine+1 < endLine {
		after = "\n" + strings.Join(lines[matchEndLine+1:endLine], "\n")
	}

	return before, matchText, after
}

// PrintMatchSummary prints a summary of matches per note
func PrintMatchSummary(matchedNotes []NoteMatch) {
	cyan := color.New(color.FgCyan, color.Bold)
	fmt.Println("\nMatches by note:")

	for _, noteMatch := range matchedNotes {
		cyan.Printf("  %s", noteMatch.Note.Title)
		fmt.Printf(" (%d matches)\n", len(noteMatch.Matches))
	}
}
