package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initEndpoint string
	initPackage  string
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new gqlforge project",
	Long: `Initialize a new gqlforge project with the standard directory structure.

Creates:
  <name>/
  ├── genqlient.yaml        # genqlient configuration
  ├── schema.graphql        # GraphQL schema (empty, run introspect to populate)
  ├── operations/           # Your GraphQL operations
  │   └── example.graphql   # Example operation file
  └── generated/            # Generated Go code (after running generate)

Example:
  gqlforge init myapi
  cd myapi
  gqlforge introspect --token xxx https://api.example.com/graphql
  # Edit operations/*.graphql
  gqlforge generate`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initEndpoint, "endpoint", "", "GraphQL endpoint URL (saved to config)")
	initCmd.Flags().StringVar(&initPackage, "package", "graphql", "Go package name for generated code")
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	projectDir := filepath.Join(outputDir, name)

	// Create directory structure
	dirs := []string{
		projectDir,
		filepath.Join(projectDir, "operations"),
		filepath.Join(projectDir, "generated"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		verboseLog("Created directory: %s", dir)
	}

	// Create genqlient.yaml
	genqlientConfig := fmt.Sprintf(`# genqlient configuration
# See: https://github.com/Khan/genqlient

schema:
  - schema.graphql

operations:
  - operations/*.graphql

generated: generated/client.go

package: %s

# Optional: customize type mappings
# bindings:
#   DateTime:
#     type: time.Time

# Optional: context handling
# context_type: context.Context
`, initPackage)

	configPath := filepath.Join(projectDir, "genqlient.yaml")
	if err := os.WriteFile(configPath, []byte(genqlientConfig), 0644); err != nil {
		return fmt.Errorf("failed to write genqlient.yaml: %w", err)
	}
	fmt.Printf("Created: %s\n", configPath)

	// Create empty schema.graphql with instructions
	schemaContent := `# GraphQL Schema
#
# This file should contain your GraphQL schema in SDL format.
#
# To populate from a GraphQL endpoint:
#   gqlforge introspect --token YOUR_TOKEN https://api.example.com/graphql
#
# Or copy your schema here manually.

# Example:
# type Query {
#   hello: String!
# }
`
	schemaPath := filepath.Join(projectDir, "schema.graphql")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		return fmt.Errorf("failed to write schema.graphql: %w", err)
	}
	fmt.Printf("Created: %s\n", schemaPath)

	// Create example operations file
	exampleOps := `# Example GraphQL Operations
#
# Define your queries, mutations, and subscriptions here.
# genqlient will generate Go functions for each operation.
#
# See: https://github.com/Khan/genqlient#operations

# Example query:
# query GetUser($id: ID!) {
#   user(id: $id) {
#     id
#     name
#     email
#   }
# }

# Example mutation:
# mutation CreateUser($input: CreateUserInput!) {
#   createUser(input: $input) {
#     id
#     name
#   }
# }
`
	examplePath := filepath.Join(projectDir, "operations", "example.graphql")
	if err := os.WriteFile(examplePath, []byte(exampleOps), 0644); err != nil {
		return fmt.Errorf("failed to write example.graphql: %w", err)
	}
	fmt.Printf("Created: %s\n", examplePath)

	// Create .gitignore
	gitignore := `# Generated files
generated/

# Secrets
.env
*.token
`
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}
	fmt.Printf("Created: %s\n", gitignorePath)

	// Create gqlforge.yaml for project-specific config
	forgeConfig := fmt.Sprintf(`# gqlforge project configuration

# GraphQL endpoint for introspection
endpoint: %q

# Authentication
# token: "" # Or use GQLFORGE_TOKEN env var
# creds: ""  # Path to goauth credentials file
# account: "" # Account key in credentials file

# Output settings
schema: schema.graphql
operations_dir: operations
`, initEndpoint)

	forgeConfigPath := filepath.Join(projectDir, "gqlforge.yaml")
	if err := os.WriteFile(forgeConfigPath, []byte(forgeConfig), 0644); err != nil {
		return fmt.Errorf("failed to write gqlforge.yaml: %w", err)
	}
	fmt.Printf("Created: %s\n", forgeConfigPath)

	fmt.Printf("\nProject initialized: %s\n", projectDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  gqlforge introspect --token YOUR_TOKEN https://api.example.com/graphql")
	fmt.Println("  # Edit operations/*.graphql with your queries/mutations")
	fmt.Println("  gqlforge generate")

	return nil
}
