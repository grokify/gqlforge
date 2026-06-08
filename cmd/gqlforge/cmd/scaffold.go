package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/gqlforge/scaffold"
	"github.com/spf13/cobra"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	scaffoldType    string
	scaffoldDepth   int
	scaffoldInclude string
	scaffoldExclude string
	scaffoldOutput  string
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold <schema.graphql>",
	Short: "Generate stub operations from schema types",
	Long: `Generate stub GraphQL operations from Query and Mutation types in a schema.

This command analyzes your schema and generates template operations for all
fields on Query and Mutation types with appropriate selection sets.

Examples:
  gqlforge scaffold schema.graphql -o operations/
  gqlforge scaffold schema.graphql --type Query --depth 2
  gqlforge scaffold schema.graphql --include "user*,feature*"
  gqlforge scaffold schema.graphql --exclude "internal*"`,
	Args: cobra.ExactArgs(1),
	RunE: runScaffold,
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)

	scaffoldCmd.Flags().StringVar(&scaffoldType, "type", "both", "Type to scaffold: Query, Mutation, or both")
	scaffoldCmd.Flags().IntVar(&scaffoldDepth, "depth", 2, "Max field depth for selection sets")
	scaffoldCmd.Flags().StringVar(&scaffoldInclude, "include", "", "Glob pattern for fields to include (comma-separated)")
	scaffoldCmd.Flags().StringVar(&scaffoldExclude, "exclude", "", "Glob pattern for fields to exclude (comma-separated)")
	scaffoldCmd.Flags().StringVarP(&scaffoldOutput, "output", "o", "operations", "Output directory")
}

func runScaffold(cmd *cobra.Command, args []string) error {
	schemaPath := args[0]
	if !filepath.IsAbs(schemaPath) {
		schemaPath = filepath.Join(outputDir, schemaPath)
	}

	// Load schema
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	verboseLog("Loading schema: %s", schemaPath)

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  schemaPath,
		Input: string(schemaContent),
	})
	if gqlErr != nil {
		return fmt.Errorf("failed to parse schema: %w", gqlErr)
	}

	// Create generator
	gen := scaffold.NewGenerator(schema)
	gen.MaxDepth = scaffoldDepth

	// Set type names
	switch strings.ToLower(scaffoldType) {
	case "query":
		gen.TypeNames = []string{"Query"}
	case "mutation":
		gen.TypeNames = []string{"Mutation"}
	case "both", "":
		gen.TypeNames = []string{"Query", "Mutation"}
	default:
		return fmt.Errorf("invalid type: %s (use Query, Mutation, or both)", scaffoldType)
	}

	// Set patterns
	gen.Include = scaffold.ParsePatterns(scaffoldInclude)
	gen.Exclude = scaffold.ParsePatterns(scaffoldExclude)

	verboseLog("Scaffolding %v with depth %d", gen.TypeNames, gen.MaxDepth)

	// Generate operations
	files, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No operations generated (no Query/Mutation types found)")
		return nil
	}

	// Resolve output directory
	outDir := scaffoldOutput
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(outputDir, outDir)
	}

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write files
	for filename, content := range files {
		path := filepath.Join(outDir, filename)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
		fmt.Printf("Wrote: %s\n", path)
	}

	fmt.Printf("Generated %d operation file(s)\n", len(files))
	return nil
}
