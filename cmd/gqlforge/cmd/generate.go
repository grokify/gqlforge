package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Khan/genqlient/generate"
	"github.com/spf13/cobra"
)

var (
	generateConfig string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go client code using genqlient",
	Long: `Generate type-safe Go client code from GraphQL operations.

This command uses genqlient as a library to generate Go code from your schema and operations.

Prerequisites:
  1. Have a valid schema.graphql (run 'gqlforge introspect' first)
  2. Define operations in operations/*.graphql files
  3. Configure genqlient.yaml

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
	// Resolve config path
	configPath := generateConfig
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(outputDir, configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("genqlient config not found: %s\nRun 'gqlforge init' to create project structure", configPath)
	}

	verboseLog("Using config: %s", configPath)

	// Load and validate config
	config, err := generate.ReadAndValidateConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	verboseLog("Config loaded: schema=%s, operations=%v", config.Schema, config.Operations)

	// Generate code
	generated, err := generate.Generate(config)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Write generated files
	for filename, content := range generated {
		// Ensure directory exists
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		if err := os.WriteFile(filename, content, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
		fmt.Printf("Wrote: %s\n", filename)
	}

	fmt.Println("Generation complete!")
	return nil
}
