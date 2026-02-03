package api

// Note represents a Joplin note
type Note struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ParentID    string `json:"parent_id"`
	UpdatedTime int64  `json:"updated_time"`
}

// NotesResponse represents the paginated response from /notes endpoint
type NotesResponse struct {
	Items   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

// UpdateNoteRequest represents the request body for updating a note
type UpdateNoteRequest struct {
	Body string `json:"body"`
}

// PingResponse represents the response from /ping endpoint
type PingResponse struct {
	Status string `json:"status"`
}

// SearchResult represents a single search result from /search endpoint
type SearchResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ParentID string `json:"parent_id"`
	Type     string `json:"type_"`
}

// SearchResponse represents the paginated response from /search endpoint
type SearchResponse struct {
	Items   []SearchResult `json:"items"`
	HasMore bool           `json:"has_more"`
}
