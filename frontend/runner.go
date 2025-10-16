package frontend

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
	"path/filepath"
	"time"
)

// Add will scan given paths and git-add them if no secrets are found.
func (rv *RuntimeValue) Add(paths ...string) error {
	// perform scan using rv configuration
	secretsFound, err := rv.Scan(paths...)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	if secretsFound {
		return fmt.Errorf("secrets detected! aborting add")
	}

	for _, f := range paths {
		if err := core.GitAdd(f); err != nil {
			return fmt.Errorf("git add failed for %s: %w", f, err)
		}
	}
	return nil
}

// Scan scans the provided project paths using the RuntimeValue's configuration.
// PrettyPrint() the result into the console
// Returns (true, nil) if secrets were found, (false, nil) when none found,
// or (false, err) on error.
func (rv *RuntimeValue) Scan(projectPaths ...string) (bool, error) {
	if len(projectPaths) == 0 {
		projectPaths = []string{"."}
	}

	// ensure result container exists and initialized
	if rv.Result == nil {
		rv.Result = &core.ScanResult{}
	}
	rv.Result.Init()

	fmt.Println("Scanning paths:", projectPaths)
	// small pause to allow user to read if used in interactive CLI
	time.Sleep(1 * time.Second)
	rv.SetTreeSitterPath("/home/holyknight101/.private/helix/runtime/grammars")
	core.IntegrateTreeSitter(rv.TreeSitterPath)

	for _, path := range projectPaths {
		if err := rv.Result.IterFolder(path, rv.Filters, rv.UseGitignore, int64(rv.MaxFileSize)); err != nil {
			return false, fmt.Errorf("scan failed for %s: %w", path, err)
		}
	}
	if rv.Result.IsFilenameMapEmpty() {
		fmt.Println("Nothing is found in the fileMap")
	}
	rv.Result.PrettyPrintResults()

	saveRoot, err := filepath.Abs(".")
	if err != nil {
		return true, fmt.Errorf("failed to resolve save path: %w", err)
	}

	if rv.LoggingEnabled {
		if err := rv.Result.SaveFilenameMap(saveRoot); err != nil {
			return true, fmt.Errorf("failed to save scan results: %w", err)
		}
	}

	return true, nil
}

// RunObfuscate loads obfuscation ruleGetFilenames from the current working directory
// and performs the obfuscation via core.LoadObfuscation.
func RunObfuscate() error {
	/*
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	if err := core.LoadObfuscation(root); err != nil {
		return fmt.Errorf("failed to obfuscate secrets: %w", err)
	}

	fmt.Println("Secrets obfuscated successfully.")
	return nil
	*/
	return nil
}
