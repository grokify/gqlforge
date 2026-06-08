package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

var (
	validateSchema     string
	validateOperations string
	validateStrict     bool
	validateJSON       bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GraphQL operations against schema",
	Long: `Validate GraphQL operations against a schema using gqlparser.

This command checks that your operations are syntactically correct and
valid against the schema before generation.

Examples:
  gqlforge validate
  gqlforge validate --schema schema.graphql --operations "operations/*.graphql"
  gqlforge validate --strict  # Fail on warnings
  gqlforge validate --json    # Output as JSON`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateSchema, "schema", "schema.graphql", "Path to schema file")
	validateCmd.Flags().StringVar(&validateOperations, "operations", "operations/*.graphql", "Glob pattern for operations files")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Treat warnings as errors")
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "Output validation results as JSON")
}

// ValidationResult represents the result of validating operations.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Rule    string `json:"rule,omitempty"`
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Resolve schema path
	schemaPath := validateSchema
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

	// Find operation files
	opsPattern := validateOperations
	if !filepath.IsAbs(opsPattern) {
		opsPattern = filepath.Join(outputDir, opsPattern)
	}

	matches, err := filepath.Glob(opsPattern)
	if err != nil {
		return fmt.Errorf("invalid operations pattern: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no operation files found matching: %s", opsPattern)
	}

	verboseLog("Found %d operation files", len(matches))

	// Load all operation files and combine into single string
	var combinedOps strings.Builder
	for _, match := range matches {
		content, err := os.ReadFile(match)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", match, err)
		}
		combinedOps.WriteString(fmt.Sprintf("# Source: %s\n", match))
		combinedOps.Write(content)
		combinedOps.WriteString("\n")
	}

	// Parse operations
	doc, gqlErrs := gqlparser.LoadQuery(schema, combinedOps.String())
	if gqlErrs != nil {
		// Convert parse errors to validation errors
		result := ValidationResult{Valid: false}
		for _, e := range gqlErrs {
			ve := ValidationError{
				Message: e.Message,
				Rule:    e.Rule,
			}
			if len(e.Locations) > 0 {
				ve.Line = e.Locations[0].Line
				ve.Column = e.Locations[0].Column
			}
			result.Errors = append(result.Errors, ve)
		}
		return outputValidationResult(result)
	}

	// Validate operations
	errs := validator.Validate(schema, doc)

	result := ValidationResult{Valid: len(errs) == 0}
	for _, e := range errs {
		ve := ValidationError{
			Message: e.Message,
			Rule:    e.Rule,
		}
		if len(e.Locations) > 0 {
			ve.Line = e.Locations[0].Line
			ve.Column = e.Locations[0].Column
		}
		result.Errors = append(result.Errors, ve)
	}

	return outputValidationResult(result)
}

func outputValidationResult(result ValidationResult) error {
	if validateJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if result.Valid {
		fmt.Println("✓ All operations are valid")
		return nil
	}

	fmt.Println("Validation errors:")
	for _, e := range result.Errors {
		location := ""
		if e.File != "" {
			location = fmt.Sprintf("%s:", e.File)
		}
		if e.Line > 0 {
			location += fmt.Sprintf("%d:%d: ", e.Line, e.Column)
		}
		rule := ""
		if e.Rule != "" {
			rule = fmt.Sprintf(" [%s]", e.Rule)
		}
		fmt.Printf("  ✗ %s%s%s\n", location, e.Message, rule)
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w.Message)
		}
	}

	if validateStrict && len(result.Warnings) > 0 {
		return fmt.Errorf("validation failed with %d errors and %d warnings (strict mode)", len(result.Errors), len(result.Warnings))
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}
