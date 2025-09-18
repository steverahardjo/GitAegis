package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"GitAegis/core"
)

func main() {
	// Resolve the full path for safety (tilde doesn't expand automatically in Go)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to get home directory:", err)
	}
	projectPath := filepath.Join(home, "Documents", "Projects", "Personal", "GitAegis")

	// Setup filters
	filters := core.AllFilters(
		core.EntropyFilter(5.0),
		core.RegexFilter(),
	)
	// Run folder iteration
	results, err := core.IterFolder(projectPath, filters)

	// Pretty print the results
	core.PrettyPrintResults(results)
	fmt.Println("Scan complete!")
}
