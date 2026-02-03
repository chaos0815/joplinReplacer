package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chaos0815/joplinReplacer/internal/api"
	"github.com/chaos0815/joplinReplacer/internal/config"
	"github.com/chaos0815/joplinReplacer/internal/replacer"
	"github.com/chaos0815/joplinReplacer/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var replaceCmd = &cobra.Command{
	Use:   "replace [search-pattern] [replacement]",
	Short: "Search and replace text in Joplin notes",
	Long: `Search for a pattern in Joplin notes and replace it with new text.
Supports both literal string matching and regex patterns.

Examples:
  # Literal replacement with preview
  joplin-replace replace --dry-run "old text" "new text"

  # Multiline regex replacement
  joplin-replace replace --regex "TODO:.*\n" "DONE:\n"

  # Case-sensitive replacement
  joplin-replace replace --case-sensitive "OldName" "NewName"`,
	Args: cobra.ExactArgs(2),
	Run:  runReplace,
}

func init() {
	rootCmd.AddCommand(replaceCmd)

	// Add flags specific to replace command
	replaceCmd.Flags().Bool("regex", false, "Treat search pattern as regex")
	replaceCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	replaceCmd.Flags().Bool("case-sensitive", false, "Case-sensitive matching")
	replaceCmd.Flags().String("notebook", "", "Filter by notebook ID")
	replaceCmd.Flags().Int("concurrency", 5, "Number of concurrent note updates (1-20)")
	replaceCmd.Flags().Duration("delay", 100*time.Millisecond, "Delay between note updates")

	// Bind flags
	viper.BindPFlag("regex", replaceCmd.Flags().Lookup("regex"))
	viper.BindPFlag("dry-run", replaceCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("case-sensitive", replaceCmd.Flags().Lookup("case-sensitive"))
	viper.BindPFlag("notebook", replaceCmd.Flags().Lookup("notebook"))
}

func runReplace(cmd *cobra.Command, args []string) {
	log := logger.Get()
	defer logger.Sync()

	// Build configuration
	cfg := config.NewConfig()
	cfg.Token = viper.GetString("token")
	cfg.Host = viper.GetString("host")
	cfg.Port = viper.GetInt("port")
	cfg.SearchPattern = args[0]
	cfg.Replacement = args[1]
	cfg.IsRegex, _ = cmd.Flags().GetBool("regex")
	cfg.CaseSensitive, _ = cmd.Flags().GetBool("case-sensitive")
	cfg.DryRun, _ = cmd.Flags().GetBool("dry-run")
	cfg.Verbose = viper.GetBool("verbose")
	cfg.NotebookID, _ = cmd.Flags().GetString("notebook")
	cfg.Concurrency, _ = cmd.Flags().GetInt("concurrency")
	cfg.Delay, _ = cmd.Flags().GetDuration("delay")

	// Handle timeout
	if viper.IsSet("timeout") {
		cfg.Timeout = viper.GetDuration("timeout")
	} else {
		cfg.Timeout = 30 * time.Second
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(cfg.Host, cfg.Port, cfg.Token, cfg.Timeout)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout*2)
	defer cancel()

	// Step 1: Ping Joplin API
	if cfg.Verbose {
		fmt.Println("Connecting to Joplin...")
	}

	if err := client.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	log.Info("Successfully connected to Joplin API")

	// Step 2: Fetch notes
	var notes []api.Note
	var err error

	if cfg.IsRegex {
		// For regex patterns, we need to fetch all notes since the API doesn't support regex search
		if cfg.Verbose {
			fmt.Println("Fetching all notes (regex mode)...")
		}
		notes, err = client.FetchAllNotes(ctx, cfg.NotebookID)
	} else {
		// For literal patterns, use search to only fetch matching notes
		if cfg.Verbose {
			fmt.Println("Searching for matching notes...")
		}
		notes, err = client.FetchMatchingNotes(ctx, cfg.SearchPattern, cfg.NotebookID)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch notes: %v\n", err)
		os.Exit(1)
	}

	log.Info("Fetched notes", zap.Int("count", len(notes)))

	if len(notes) == 0 {
		fmt.Println("No notes found matching the search criteria")
		return
	}

	if cfg.Verbose {
		if cfg.IsRegex {
			fmt.Printf("Fetched %d notes\n", len(notes))
		} else {
			fmt.Printf("Found %d notes containing the search pattern\n", len(notes))
		}
	}

	// Step 3: Create matcher
	matcher, err := replacer.NewMatcher(cfg.SearchPattern, cfg.IsRegex, cfg.CaseSensitive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid pattern: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Create replacer and process notes
	if cfg.Verbose {
		fmt.Println("Searching for matches...")
	}

	repl := replacer.NewReplacer(matcher, cfg.Replacement, cfg.DryRun, cfg.Verbose, cfg.Concurrency, cfg.Delay)
	result, err := repl.ProcessNotes(ctx, client, notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing notes: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Display results
	if cfg.DryRun && result.NotesWithMatches > 0 {
		replacer.PrintPreview(result.MatchedNotes, cfg.Replacement)
		fmt.Println()
		replacer.PrintMatchSummary(result.MatchedNotes)
		fmt.Println()
	}

	fmt.Println(replacer.GetResultSummary(result, cfg.DryRun))

	// Exit with error code if there were failures
	if len(result.FailedUpdates) > 0 {
		os.Exit(1)
	}
}
