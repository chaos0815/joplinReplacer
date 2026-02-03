package cmd

import (
	"fmt"
	"os"

	"github.com/chaos0815/joplinReplacer/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "joplin-replace",
	Short: "Search and replace text in Joplin notes",
	Long: `A CLI tool to search and replace multiline patterns in Joplin notes via the local REST API.
Supports both literal and regex patterns with dry-run mode for safety.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.joplin-replace.yaml)")
	rootCmd.PersistentFlags().String("token", "", "Joplin API token (or set JOPLIN_TOKEN env var)")
	rootCmd.PersistentFlags().String("host", "localhost", "Joplin API host")
	rootCmd.PersistentFlags().Int("port", 41184, "Joplin API port")
	rootCmd.PersistentFlags().Duration("timeout", 0, "API request timeout (default: 30s)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Bind flags to viper
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Bind environment variables
	viper.BindEnv("token", "JOPLIN_TOKEN")
	viper.SetEnvPrefix("JOPLIN")
	viper.AutomaticEnv()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".joplin-replace")
	}

	// Read config file (optional)
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	// Initialize logger
	if err := logger.Init(viper.GetBool("verbose")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}
