package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	credsFile string
	account   string
	verbose   bool
	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "gqlforge",
	Short: "GraphQL client SDK generator for Go",
	Long: `gqlforge is a tool for generating type-safe GraphQL client SDKs for Go.

It uses GraphQL introspection to fetch schemas and genqlient to generate
type-safe client code.

Workflow:
  1. gqlforge introspect <endpoint>  # Fetch and save GraphQL schema
  2. gqlforge generate               # Generate Go client code

Authentication:
  Use goauth credentials file format:
    --creds <file>     Path to goauth credentials JSON file
    --account <key>    Account key within credentials file

  Or environment variables:
    GQLFORGE_TOKEN     Bearer token for authentication
    GQLFORGE_ENDPOINT  GraphQL endpoint URL`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&credsFile, "creds", "", "Path to goauth credentials file")
	rootCmd.PersistentFlags().StringVar(&account, "account", "", "Account key in credentials file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", ".", "Output directory")
}

func verboseLog(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
