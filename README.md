# Joplin Replace

A CLI tool to search and replace multiline patterns in Joplin notes via the local REST API. Supports both literal and regex patterns with dry-run mode for safety.

## Features

- **Pattern Matching**: Supports both literal string matching and regex patterns
- **Multiline Support**: Handle patterns that span multiple lines
- **Dry-Run Mode**: Preview changes before applying them
- **Case Sensitivity**: Optional case-sensitive or case-insensitive matching
- **Batch Operations**: Process all notes or filter by notebook
- **Error Recovery**: Retry logic with exponential backoff
- **Colored Output**: Highlighted preview of changes

## Prerequisites

- **Joplin Desktop**: Must be running with Web Clipper enabled
  - Go to: Tools → Options → Web Clipper
  - Enable the Web Clipper service
  - Note down the API token
- **Go**: Version 1.21 or later

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/chaos0815/joplinReplacer.git
cd joplinReplacer

# Build the binary
go build -o joplin-replace

# Optional: Install to system
go install
```

## Configuration

### API Token

You need a Joplin API token. Get it from:
1. Open Joplin Desktop
2. Go to Tools → Options → Web Clipper
3. Copy the authorization token

Provide the token via:
- **Flag**: `--token YOUR_TOKEN`
- **Environment variable**: `JOPLIN_TOKEN=YOUR_TOKEN`
- **Config file**: `~/.joplin-replace.yaml`

### Environment Variables

Create a `.env` file or export variables:

```bash
export JOPLIN_TOKEN="your_api_token_here"
export JOPLIN_HOST="localhost"  # optional
export JOPLIN_PORT="41184"      # optional
```

## Usage

### Basic Syntax

```bash
joplin-replace replace [flags] <search-pattern> <replacement>
```

### Examples

#### 1. Literal Replacement with Preview

```bash
joplin-replace replace --dry-run "old text" "new text"
```

#### 2. Multiline Regex Replacement

```bash
# Mark all TODO items as done
joplin-replace replace --regex "- \[ \] (.+)" "- [x] $1"
```

#### 3. Case-Sensitive Replacement

```bash
joplin-replace replace --case-sensitive "OldName" "NewName"
```

#### 4. Using Environment Variables

```bash
export JOPLIN_TOKEN="abc123"
joplin-replace replace "search" "replace"
```

#### 5. Filter by Notebook

```bash
joplin-replace replace --notebook="notebook_id_here" "old" "new"
```

#### 6. Multiline Pattern Examples

```bash
# Replace code blocks
joplin-replace replace --regex "```python\n.*?\n```" "```javascript\n// code here\n```"

# Update section headers
joplin-replace replace --regex "## Old Section.*?\n" "## New Section\n"
```

### Flags

#### Required (or via environment)
- `--token string`: Joplin API token (env: `JOPLIN_TOKEN`)

#### Optional
- `--host string`: API host (default: "localhost")
- `--port int`: API port (default: 41184)
- `--regex`: Treat search as regex pattern
- `--dry-run`: Preview changes without applying
- `--case-sensitive`: Case-sensitive matching (default: false)
- `--notebook string`: Filter by notebook ID
- `--verbose`: Enable verbose logging
- `--timeout duration`: API timeout (default: 30s)

## How It Works

1. **Connect**: Pings the Joplin API to verify connection
2. **Fetch**: Retrieves all notes (with pagination)
3. **Search**: Finds all matches using literal or regex patterns
4. **Preview/Apply**:
   - In dry-run mode: Shows preview of changes
   - In apply mode: Updates notes via API
5. **Report**: Displays summary of operations

## API Details

### Joplin Web Clipper API

The tool connects to the local Joplin API:
- **Default URL**: `http://localhost:41184`
- **Authentication**: Token-based via query parameter
- **Endpoints Used**:
  - `GET /ping`: Health check
  - `GET /notes`: Fetch notes (paginated)
  - `PUT /notes/:id`: Update note body

### Rate Limiting

The tool includes:
- 100ms delay between updates
- Exponential backoff retry (3 attempts)
- Configurable timeout (default: 30s)

## Safety Features

1. **Dry-Run Mode**: Always test with `--dry-run` first
2. **Preview Display**: Shows exact changes with context
3. **Metadata Preservation**: Only updates note body, preserves title, tags, timestamps
4. **Error Handling**: Logs errors and continues with other notes
5. **No Deletion**: Only modifies note content, never deletes

## Error Messages

| Message | Cause | Solution |
|---------|-------|----------|
| "Cannot connect to Joplin" | Joplin not running or Web Clipper disabled | Start Joplin Desktop and enable Web Clipper |
| "Authentication failed" | Invalid API token | Check token in Joplin settings |
| "Invalid regex pattern" | Malformed regex | Fix regex syntax |
| "No matches found" | Pattern not found | Check pattern spelling/format |

## Development

### Project Structure

```
joplinReplacer/
├── main.go                    # Entry point
├── go.mod                     # Module definition
├── README.md                  # Documentation
├── .env.example               # Example environment variables
├── cmd/
│   ├── root.go               # Root command setup
│   └── replace.go            # Replace command implementation
├── internal/
│   ├── api/
│   │   ├── types.go          # API data structures
│   │   ├── client.go         # HTTP client with auth/retry
│   │   └── notes.go          # Note fetch/update operations
│   ├── replacer/
│   │   ├── matcher.go        # Pattern matching (literal/regex)
│   │   ├── replacer.go       # Core search/replace logic
│   │   └── preview.go        # Dry-run preview formatting
│   └── config/
│       └── config.go         # Configuration management
└── pkg/
    └── logger/
        └── logger.go         # Logging utilities
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/replacer/...
```

### Building

```bash
# Build for current platform
go build -o joplin-replace

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o joplin-replace-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o joplin-replace-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o joplin-replace-windows-amd64.exe
```

## Troubleshooting

### "Cannot connect to Joplin"

1. Verify Joplin Desktop is running
2. Check Web Clipper is enabled (Tools → Options → Web Clipper)
3. Verify the port (default: 41184)
4. Try accessing `http://localhost:41184/ping` in a browser

### "Authentication failed"

1. Get a fresh token from Joplin (Tools → Options → Web Clipper)
2. Verify the token has no extra spaces or characters
3. Try using the `--token` flag directly instead of environment variable

### Regex Not Matching Multiline Content

- The tool automatically adds `(?s)` flag for multiline support
- Test your regex at [regex101.com](https://regex101.com) with the "dotall" flag
- Remember to escape special characters: `\n`, `\t`, `\[`, `\]`, etc.

## Contributing

If you really must:
1. Fork the repository
2. ...
3. PROFIT!

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration
- Colored output via [color](https://github.com/fatih/color)
- Logging with [zap](https://github.com/uber-go/zap)

## Support

Sorry for the AI slop, but I was in dire need of fixing >4000 borked notes.
I can't give support for this as I hope it was a once only use.

Putting it up, as it worked for me, and in case someone is in a similar situation.

