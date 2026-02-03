package replacer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chaos0815/joplinReplacer/internal/api"
)

// Result represents the outcome of a replacement operation
type Result struct {
	NotesScanned     int
	NotesWithMatches int
	TotalMatches     int
	NotesUpdated     int
	FailedUpdates    []FailedUpdate
	MatchedNotes     []NoteMatch
}

// FailedUpdate represents a failed note update
type FailedUpdate struct {
	NoteID    string
	NoteTitle string
	Error     error
}

// NoteMatch represents matches found in a note
type NoteMatch struct {
	Note    api.Note
	Matches []Match
}

// Replacer coordinates the search and replace operations
type Replacer struct {
	matcher     Matcher
	replacement string
	dryRun      bool
	verbose     bool
	concurrency int
	delay       time.Duration
}

// NewReplacer creates a new replacer
func NewReplacer(matcher Matcher, replacement string, dryRun bool, verbose bool, concurrency int, delay time.Duration) *Replacer {
	return &Replacer{
		matcher:     matcher,
		replacement: replacement,
		dryRun:      dryRun,
		verbose:     verbose,
		concurrency: concurrency,
		delay:       delay,
	}
}

// ProcessNotes processes all notes and performs replacements
func (r *Replacer) ProcessNotes(ctx context.Context, client *api.Client, notes []api.Note) (*Result, error) {
	result := &Result{
		NotesScanned: len(notes),
	}

	// Find all matches
	for _, note := range notes {
		matches := r.matcher.FindAllMatches(note.Body)
		if len(matches) > 0 {
			result.NotesWithMatches++
			result.TotalMatches += len(matches)
			result.MatchedNotes = append(result.MatchedNotes, NoteMatch{
				Note:    note,
				Matches: matches,
			})
		}
	}

	// If dry-run, don't perform updates
	if r.dryRun {
		return result, nil
	}

	// Perform updates with concurrency
	return r.performConcurrentUpdates(ctx, client, result)
}

// performConcurrentUpdates updates notes concurrently with progress reporting
func (r *Replacer) performConcurrentUpdates(ctx context.Context, client *api.Client, result *Result) (*Result, error) {
	totalNotes := len(result.MatchedNotes)
	if totalNotes == 0 {
		return result, nil
	}

	// Create channels for work distribution
	jobs := make(chan NoteMatch, totalNotes)
	results := make(chan updateResult, totalNotes)

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < r.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.updateWorker(ctx, client, jobs, results)
		}()
	}

	// Send jobs to workers
	for _, noteMatch := range result.MatchedNotes {
		jobs <- noteMatch
	}
	close(jobs)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with progress reporting
	processed := 0
	for updateRes := range results {
		processed++
		if updateRes.err != nil {
			result.FailedUpdates = append(result.FailedUpdates, FailedUpdate{
				NoteID:    updateRes.noteID,
				NoteTitle: updateRes.noteTitle,
				Error:     updateRes.err,
			})
		} else {
			result.NotesUpdated++
		}

		// Show progress
		if r.verbose {
			percentage := float64(processed) / float64(totalNotes) * 100
			fmt.Printf("\rUpdating notes: %d/%d (%.1f%%)", processed, totalNotes, percentage)
		}
	}

	if r.verbose {
		fmt.Println() // New line after progress
	}

	return result, nil
}

// updateResult represents the result of a single note update
type updateResult struct {
	noteID    string
	noteTitle string
	err       error
}

// updateWorker processes note updates from the jobs channel
func (r *Replacer) updateWorker(ctx context.Context, client *api.Client, jobs <-chan NoteMatch, results chan<- updateResult) {
	for noteMatch := range jobs {
		newBody := r.matcher.Replace(noteMatch.Note.Body, r.replacement)
		err := client.UpdateNote(ctx, noteMatch.Note.ID, newBody, r.delay)
		results <- updateResult{
			noteID:    noteMatch.Note.ID,
			noteTitle: noteMatch.Note.Title,
			err:       err,
		}
	}
}

// GetResultSummary returns a human-readable summary of the results
func GetResultSummary(result *Result, dryRun bool) string {
	if result.NotesWithMatches == 0 {
		return "No matches found in any notes"
	}

	summary := fmt.Sprintf("Scanned %d notes\n", result.NotesScanned)
	summary += fmt.Sprintf("Found %d matches in %d notes\n", result.TotalMatches, result.NotesWithMatches)

	if dryRun {
		summary += "\nDRY-RUN MODE: No changes were made"
	} else {
		summary += fmt.Sprintf("\nSuccessfully updated %d notes", result.NotesUpdated)
		if len(result.FailedUpdates) > 0 {
			summary += fmt.Sprintf("\nFailed to update %d notes:", len(result.FailedUpdates))
			for _, failed := range result.FailedUpdates {
				summary += fmt.Sprintf("\n  - %s (%s): %v", failed.NoteTitle, failed.NoteID, failed.Error)
			}
		}
	}

	return summary
}
