package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SearchNotes searches for notes containing the query string
func (c *Client) SearchNotes(ctx context.Context, query string, notebookID string) ([]SearchResult, error) {
	var allResults []SearchResult
	page := 1
	limit := 100

	for {
		params := url.Values{}
		params.Set("query", query)
		params.Set("type", "note")
		params.Set("fields", "id,title,parent_id")
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("limit", fmt.Sprintf("%d", limit))

		resp, err := c.get(ctx, "/search", params)
		if err != nil {
			return nil, fmt.Errorf("failed to search notes (page %d): %w", page, err)
		}

		var searchResp SearchResponse
		if err := parseJSON(resp, &searchResp); err != nil {
			return nil, fmt.Errorf("failed to parse search response (page %d): %w", page, err)
		}

		// Filter by notebook if specified
		if notebookID != "" {
			for _, result := range searchResp.Items {
				if result.ParentID == notebookID {
					allResults = append(allResults, result)
				}
			}
		} else {
			allResults = append(allResults, searchResp.Items...)
		}

		if !searchResp.HasMore {
			break
		}

		page++
	}

	return allResults, nil
}

// FetchNoteByID fetches a single note by its ID with full body
func (c *Client) FetchNoteByID(ctx context.Context, noteID string) (*Note, error) {
	params := url.Values{}
	params.Set("fields", "id,title,body,parent_id,updated_time")

	resp, err := c.get(ctx, fmt.Sprintf("/notes/%s", noteID), params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch note %s: %w", noteID, err)
	}

	var note Note
	if err := parseJSON(resp, &note); err != nil {
		return nil, fmt.Errorf("failed to parse note %s: %w", noteID, err)
	}

	return &note, nil
}

// FetchMatchingNotes searches for notes containing the query and fetches their full content
func (c *Client) FetchMatchingNotes(ctx context.Context, query string, notebookID string) ([]Note, error) {
	// Search for notes containing the query
	searchResults, err := c.SearchNotes(ctx, query, notebookID)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}

	if len(searchResults) == 0 {
		return []Note{}, nil
	}

	// Fetch full note bodies for each search result
	notes := make([]Note, 0, len(searchResults))
	for _, result := range searchResults {
		note, err := c.FetchNoteByID(ctx, result.ID)
		if err != nil {
			// Log error but continue with other notes
			continue
		}
		notes = append(notes, *note)
	}

	return notes, nil
}

// FetchAllNotes fetches all notes from Joplin with pagination (for regex patterns)
func (c *Client) FetchAllNotes(ctx context.Context, notebookID string) ([]Note, error) {
	var allNotes []Note
	page := 1
	limit := 100

	for {
		params := url.Values{}
		params.Set("fields", "id,title,body,parent_id,updated_time")
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("limit", fmt.Sprintf("%d", limit))

		resp, err := c.get(ctx, "/notes", params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch notes (page %d): %w", page, err)
		}

		var notesResp NotesResponse
		if err := parseJSON(resp, &notesResp); err != nil {
			return nil, fmt.Errorf("failed to parse notes response (page %d): %w", page, err)
		}

		// Filter by notebook if specified
		if notebookID != "" {
			for _, note := range notesResp.Items {
				if note.ParentID == notebookID {
					allNotes = append(allNotes, note)
				}
			}
		} else {
			allNotes = append(allNotes, notesResp.Items...)
		}

		if !notesResp.HasMore {
			break
		}

		page++
	}

	return allNotes, nil
}

// UpdateNote updates a note's body
func (c *Client) UpdateNote(ctx context.Context, noteID string, newBody string, delay time.Duration) error {
	updateReq := UpdateNoteRequest{
		Body: newBody,
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/notes/%s?token=%s", c.baseURL, noteID, url.QueryEscape(c.token))

	resp, err := c.doRequestWithRetry(ctx, "PUT", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed. Check your API token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Add configurable delay to avoid overwhelming the API
	if delay > 0 {
		time.Sleep(delay)
	}

	return nil
}
