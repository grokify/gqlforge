package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchConfig   string
	watchDebounce time.Duration
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch files and re-generate on changes",
	Long: `Watch schema and operation files for changes and automatically regenerate.

This command monitors your GraphQL files and runs the generate command
whenever changes are detected.

Examples:
  gqlforge watch
  gqlforge watch --config genqlient.yaml
  gqlforge watch --debounce 500ms`,
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().StringVar(&watchConfig, "config", "genqlient.yaml", "Path to genqlient.yaml configuration")
	watchCmd.Flags().DurationVar(&watchDebounce, "debounce", 300*time.Millisecond, "Debounce duration for rapid changes")
}

func runWatch(cmd *cobra.Command, args []string) error {
	// Resolve config path
	configPath := watchConfig
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(outputDir, configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("genqlient config not found: %s\nRun 'gqlforge init' to create project structure", configPath)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Get directories to watch
	configDir := filepath.Dir(configPath)
	schemaPath := filepath.Join(configDir, "schema.graphql")
	operationsDir := filepath.Join(configDir, "operations")

	// Add schema file to watch
	if _, err := os.Stat(schemaPath); err == nil {
		if err := watcher.Add(schemaPath); err != nil {
			return fmt.Errorf("failed to watch schema: %w", err)
		}
		verboseLog("Watching: %s", schemaPath)
	}

	// Add operations directory to watch
	if _, err := os.Stat(operationsDir); err == nil {
		if err := watcher.Add(operationsDir); err != nil {
			return fmt.Errorf("failed to watch operations: %w", err)
		}
		verboseLog("Watching: %s", operationsDir)
	}

	// Add config file to watch
	if err := watcher.Add(configPath); err != nil {
		return fmt.Errorf("failed to watch config: %w", err)
	}
	verboseLog("Watching: %s", configPath)

	fmt.Println("Watching for changes... (press Ctrl+C to stop)")

	// Initial generation
	fmt.Println("\nRunning initial generation...")
	generateConfig = watchConfig
	if err := runGenerate(cmd, nil); err != nil {
		fmt.Printf("Generation failed: %v\n", err)
	}

	// Debounce timer
	var debounceTimer *time.Timer
	debounceTimer = time.NewTimer(0)
	<-debounceTimer.C // Drain initial timer

	regenerate := func() {
		fmt.Println("\n--- File changed, regenerating... ---")
		generateConfig = watchConfig
		if err := runGenerate(cmd, nil); err != nil {
			fmt.Printf("Generation failed: %v\n", err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only respond to write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				// Only watch .graphql and .yaml files
				ext := filepath.Ext(event.Name)
				if ext == ".graphql" || ext == ".yaml" || ext == ".yml" {
					verboseLog("Change detected: %s (%s)", event.Name, event.Op)

					// Reset debounce timer
					if !debounceTimer.Stop() {
						select {
						case <-debounceTimer.C:
						default:
						}
					}
					debounceTimer.Reset(watchDebounce)
				}
			}

		case <-debounceTimer.C:
			regenerate()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}
