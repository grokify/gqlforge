package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	generateConfig string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go client code using genqlient",
	Long: `Generate type-safe Go client code from GraphQL operations.

This command wraps genqlient to generate Go code from your schema and operations.

Prerequisites:
  1. Install genqlient: go install github.com/Khan/genqlient@latest
  2. Have a valid schema.graphql (run 'gqlforge introspect' first)
  3. Define operations in operations/*.graphql files
  4. Configure genqlient.yaml

Example:
  gqlforge generate
  gqlforge generate --config custom-genqlient.yaml`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateConfig, "config", "genqlient.yaml", "Path to genqlient.yaml configuration")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Check if genqlient is installed
	genqlientPath, err := exec.LookPath("genqlient")
	if err != nil {
		return fmt.Errorf("genqlient not found in PATH. Install with: go install github.com/Khan/genqlient@latest")
	}
	verboseLog("Found genqlient: %s", genqlientPath)

	// Check if config exists
	configPath := generateConfig
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(outputDir, configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("genqlient config not found: %s\nRun 'gqlforge init' to create project structure", configPath)
	}

	verboseLog("Using config: %s", configPath)

	// Run genqlient
	genqlientCmd := exec.Command("genqlient", configPath)
	genqlientCmd.Dir = filepath.Dir(configPath)
	genqlientCmd.Stdout = os.Stdout
	genqlientCmd.Stderr = os.Stderr

	fmt.Printf("Running: genqlient %s\n", configPath)

	if err := genqlientCmd.Run(); err != nil {
		return fmt.Errorf("genqlient failed: %w", err)
	}

	fmt.Println("Generation complete!")
	return nil
}
