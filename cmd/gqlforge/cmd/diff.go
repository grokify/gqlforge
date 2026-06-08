package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/grokify/gqlforge/introspection"
	"github.com/spf13/cobra"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	diffToken        string
	diffBreakingOnly bool
	diffJSONOutput   bool
)

var diffCmd = &cobra.Command{
	Use:   "diff <local-schema> <remote-endpoint>",
	Short: "Compare local schema to remote endpoint",
	Long: `Compare a local GraphQL schema file to a remote endpoint to detect drift.

This command introspects the remote endpoint and compares it to your local
schema, highlighting added, removed, and changed types and fields.

Examples:
  gqlforge diff schema.graphql https://api.example.com/graphql --token xxx
  gqlforge diff schema.graphql --creds creds.json --account myapi https://api.example.com/graphql
  gqlforge diff schema.graphql https://api.example.com/graphql --breaking-only
  gqlforge diff schema.graphql https://api.example.com/graphql --json`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&diffToken, "token", "", "Bearer token for remote introspection")
	diffCmd.Flags().BoolVar(&diffBreakingOnly, "breaking-only", false, "Only show breaking changes")
	diffCmd.Flags().BoolVar(&diffJSONOutput, "json", false, "Output diff as JSON")
}

// SchemaDiff represents the differences between two schemas.
type SchemaDiff struct {
	AddedTypes    []string                 `json:"addedTypes,omitempty"`
	RemovedTypes  []string                 `json:"removedTypes,omitempty"`
	AddedFields   map[string][]string      `json:"addedFields,omitempty"`
	RemovedFields map[string][]string      `json:"removedFields,omitempty"`
	ChangedFields map[string][]FieldChange `json:"changedFields,omitempty"`
	Breaking      bool                     `json:"breaking"`
	BreakingCount int                      `json:"breakingCount"`
}

// FieldChange represents a change to a field.
type FieldChange struct {
	Name     string `json:"name"`
	OldType  string `json:"oldType"`
	NewType  string `json:"newType"`
	Breaking bool   `json:"breaking"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	localPath := args[0]
	remoteURL := args[1]

	// Resolve local path
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(outputDir, localPath)
	}

	// Load local schema
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local schema: %w", err)
	}

	verboseLog("Loading local schema: %s", localPath)

	localSchema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  localPath,
		Input: string(localContent),
	})
	if gqlErr != nil {
		return fmt.Errorf("failed to parse local schema: %w", gqlErr)
	}

	// Get HTTP client with authentication (reuse token if provided)
	introspectToken = diffToken
	httpClient, err := getHTTPClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	verboseLog("Introspecting remote: %s", remoteURL)

	// Introspect remote
	client := introspection.NewClient(remoteURL, httpClient)
	remoteResult, err := client.Introspect(ctx)
	if err != nil {
		return fmt.Errorf("introspection failed: %w", err)
	}

	// Compare schemas
	diff := compareSchemas(localSchema, &remoteResult.Schema)

	// Output result
	if diffJSONOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(diff)
	}

	printDiff(localPath, remoteURL, diff)
	return nil
}

func compareSchemas(local *ast.Schema, remote *introspection.Schema) *SchemaDiff {
	diff := &SchemaDiff{
		AddedFields:   make(map[string][]string),
		RemovedFields: make(map[string][]string),
		ChangedFields: make(map[string][]FieldChange),
	}

	// Build type maps
	localTypes := make(map[string]*ast.Definition)
	for _, t := range local.Types {
		if !isBuiltInType(t.Name) {
			localTypes[t.Name] = t
		}
	}

	remoteTypes := make(map[string]*introspection.FullType)
	for i := range remote.Types {
		t := &remote.Types[i]
		if !isBuiltInType(t.Name) {
			remoteTypes[t.Name] = t
		}
	}

	// Find added types (in remote but not local)
	for name := range remoteTypes {
		if _, exists := localTypes[name]; !exists {
			diff.AddedTypes = append(diff.AddedTypes, name)
		}
	}
	sort.Strings(diff.AddedTypes)

	// Find removed types (in local but not remote) - BREAKING
	for name := range localTypes {
		if _, exists := remoteTypes[name]; !exists {
			diff.RemovedTypes = append(diff.RemovedTypes, name)
			diff.Breaking = true
			diff.BreakingCount++
		}
	}
	sort.Strings(diff.RemovedTypes)

	// Compare fields for types that exist in both
	for name, localType := range localTypes {
		remoteType, exists := remoteTypes[name]
		if !exists {
			continue
		}

		// Build field maps
		localFields := make(map[string]*ast.FieldDefinition)
		for _, f := range localType.Fields {
			localFields[f.Name] = f
		}

		remoteFields := make(map[string]*introspection.Field)
		for i := range remoteType.Fields {
			f := &remoteType.Fields[i]
			remoteFields[f.Name] = f
		}

		// Find added fields
		for fieldName := range remoteFields {
			if _, exists := localFields[fieldName]; !exists {
				diff.AddedFields[name] = append(diff.AddedFields[name], fieldName)
			}
		}

		// Find removed fields - BREAKING
		for fieldName := range localFields {
			if _, exists := remoteFields[fieldName]; !exists {
				diff.RemovedFields[name] = append(diff.RemovedFields[name], fieldName)
				diff.Breaking = true
				diff.BreakingCount++
			}
		}

		// Find changed fields
		for fieldName, localField := range localFields {
			remoteField, exists := remoteFields[fieldName]
			if !exists {
				continue
			}

			localTypeStr := localField.Type.String()
			remoteTypeStr := typeRefToString(&remoteField.Type)

			if localTypeStr != remoteTypeStr {
				change := FieldChange{
					Name:    fieldName,
					OldType: localTypeStr,
					NewType: remoteTypeStr,
				}
				// Type changes are breaking if they're not compatible
				// For simplicity, we mark all type changes as breaking
				change.Breaking = true
				diff.ChangedFields[name] = append(diff.ChangedFields[name], change)
				if change.Breaking {
					diff.Breaking = true
					diff.BreakingCount++
				}
			}
		}
	}

	return diff
}

func typeRefToString(tr *introspection.TypeRef) string {
	if tr == nil {
		return ""
	}
	switch tr.Kind {
	case introspection.KindNonNull:
		return typeRefToString(tr.OfType) + "!"
	case introspection.KindList:
		return "[" + typeRefToString(tr.OfType) + "]"
	default:
		return tr.Name
	}
}

func isBuiltInType(name string) bool {
	switch name {
	case "String", "Int", "Float", "Boolean", "ID":
		return true
	}
	return len(name) > 2 && name[:2] == "__"
}

func printDiff(localPath, remoteURL string, diff *SchemaDiff) {
	fmt.Printf("Schema Diff: %s vs %s\n\n", localPath, remoteURL)

	if len(diff.AddedTypes) > 0 && !diffBreakingOnly {
		fmt.Println("+ Added Types:")
		for _, t := range diff.AddedTypes {
			fmt.Printf("  + %s\n", t)
		}
		fmt.Println()
	}

	if len(diff.RemovedTypes) > 0 {
		fmt.Println("- Removed Types (BREAKING):")
		for _, t := range diff.RemovedTypes {
			fmt.Printf("  - %s\n", t)
		}
		fmt.Println()
	}

	// Collect types with changes
	changedTypes := make(map[string]bool)
	for typeName := range diff.AddedFields {
		changedTypes[typeName] = true
	}
	for typeName := range diff.RemovedFields {
		changedTypes[typeName] = true
	}
	for typeName := range diff.ChangedFields {
		changedTypes[typeName] = true
	}

	if len(changedTypes) > 0 {
		fmt.Println("~ Changed Types:")
		typeNames := make([]string, 0, len(changedTypes))
		for name := range changedTypes {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)

		for _, typeName := range typeNames {
			hasBreaking := len(diff.RemovedFields[typeName]) > 0 || len(diff.ChangedFields[typeName]) > 0
			if diffBreakingOnly && !hasBreaking {
				continue
			}

			fmt.Printf("  ~ %s:\n", typeName)

			if !diffBreakingOnly {
				for _, f := range diff.AddedFields[typeName] {
					fmt.Printf("    + %s\n", f)
				}
			}

			for _, f := range diff.RemovedFields[typeName] {
				fmt.Printf("    - %s (BREAKING)\n", f)
			}

			for _, c := range diff.ChangedFields[typeName] {
				breaking := ""
				if c.Breaking {
					breaking = " (BREAKING)"
				}
				fmt.Printf("    ~ %s: %s → %s%s\n", c.Name, c.OldType, c.NewType, breaking)
			}
		}
		fmt.Println()
	}

	// Summary
	addCount := len(diff.AddedTypes)
	for _, fields := range diff.AddedFields {
		addCount += len(fields)
	}
	removeCount := len(diff.RemovedTypes)
	for _, fields := range diff.RemovedFields {
		removeCount += len(fields)
	}

	if addCount == 0 && removeCount == 0 && diff.BreakingCount == 0 {
		fmt.Println("No differences found.")
	} else {
		fmt.Printf("Summary: %d added, %d removed, %d breaking changes\n", addCount, removeCount, diff.BreakingCount)
	}
}
