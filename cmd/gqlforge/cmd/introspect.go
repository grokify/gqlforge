package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/grokify/goauth"
	"github.com/grokify/goauth/authutil"
	"github.com/grokify/gqlforge/introspection"
	"github.com/spf13/cobra"
)

var (
	introspectToken    string
	introspectEndpoint string
	outputSDL          bool
	outputJSON         bool
	schemaName         string
)

var introspectCmd = &cobra.Command{
	Use:   "introspect [endpoint]",
	Short: "Fetch GraphQL schema via introspection",
	Long: `Fetch a GraphQL schema from a server using introspection queries.

The schema can be output in two formats:
  - SDL (Schema Definition Language): schema.graphql (default)
  - JSON (raw introspection result): schema.json

Authentication options:
  1. goauth credentials file:
     gqlforge introspect --creds credentials.json --account myapi https://api.example.com/graphql

  2. Bearer token flag:
     gqlforge introspect --token <token> https://api.example.com/graphql

  3. Environment variables:
     GQLFORGE_TOKEN=<token> gqlforge introspect https://api.example.com/graphql

Examples:
  # Introspect and save as SDL
  gqlforge introspect https://api.example.com/graphql -o ./schema

  # Introspect with authentication
  gqlforge introspect --creds ~/.config/goauth/credentials.json --account myapi https://api.example.com/graphql

  # Output both SDL and JSON
  gqlforge introspect --sdl --json https://api.example.com/graphql`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIntrospect,
}

func init() {
	rootCmd.AddCommand(introspectCmd)

	introspectCmd.Flags().StringVar(&introspectToken, "token", "", "Bearer token for authentication")
	introspectCmd.Flags().StringVar(&introspectEndpoint, "endpoint", "", "GraphQL endpoint (alternative to positional arg)")
	introspectCmd.Flags().BoolVar(&outputSDL, "sdl", true, "Output schema as SDL (.graphql)")
	introspectCmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw introspection result (.json)")
	introspectCmd.Flags().StringVar(&schemaName, "name", "schema", "Base name for output files")
}

func runIntrospect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine endpoint
	endpoint := introspectEndpoint
	if len(args) > 0 {
		endpoint = args[0]
	}
	if endpoint == "" {
		endpoint = os.Getenv("GQLFORGE_ENDPOINT")
	}
	if endpoint == "" {
		return fmt.Errorf("endpoint is required: provide as argument, --endpoint flag, or GQLFORGE_ENDPOINT env var")
	}

	// Get HTTP client with authentication
	httpClient, err := getHTTPClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	verboseLog("Introspecting: %s", endpoint)

	// Create introspection client
	client := introspection.NewClient(endpoint, httpClient)

	// Perform introspection
	result, err := client.Introspect(ctx)
	if err != nil {
		return fmt.Errorf("introspection failed: %w", err)
	}

	verboseLog("Introspection successful: found %d types", len(result.Schema.Types))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Output SDL
	if outputSDL {
		sdlPath := filepath.Join(outputDir, schemaName+".graphql")
		sdl, err := result.ToSDL()
		if err != nil {
			return fmt.Errorf("failed to convert to SDL: %w", err)
		}
		if err := os.WriteFile(sdlPath, []byte(sdl), 0o600); err != nil {
			return fmt.Errorf("failed to write SDL: %w", err)
		}
		fmt.Printf("Wrote SDL schema: %s\n", sdlPath)
	}

	// Output JSON
	if outputJSON {
		jsonPath := filepath.Join(outputDir, schemaName+".json")
		jsonData, err := result.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to convert to JSON: %w", err)
		}
		if err := os.WriteFile(jsonPath, jsonData, 0o600); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
		fmt.Printf("Wrote JSON schema: %s\n", jsonPath)
	}

	return nil
}

func getHTTPClient(ctx context.Context) (*http.Client, error) {
	// Try goauth credentials first
	if credsFile != "" && account != "" {
		verboseLog("Loading credentials from %s (account: %s)", credsFile, account)
		return goauth.NewClient(ctx, credsFile, account)
	}

	// Try token from flag or environment
	token := introspectToken
	if token == "" {
		token = os.Getenv("GQLFORGE_TOKEN")
	}

	if token != "" {
		verboseLog("Using bearer token authentication")
		return authutil.NewClientToken(authutil.TokenBearer, token, false), nil
	}

	// No authentication - return default client with warning
	verboseLog("Warning: No authentication provided, introspection may fail")
	return http.DefaultClient, nil
}
