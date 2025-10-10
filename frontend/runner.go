package frontend

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
	"os"
	"path/filepath"
	"time"
)

func Add(logging bool, paths ...string) error {
	secretsFound, err := Scan(5.0, logging, global_gitignore, int(global_filemaxsize), global_filters, paths...)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	if secretsFound {
		return fmt.Errorf("secrets detected! aborting add")
	}
	for _, f := range paths {
		if err := core.GitAdd(f); err != nil {
			return err
		}
	}
	return nil
}
//Runner function in the frontend that stick together all core functions
func Scan(entropyLimit float64, logging bool, global_gitignore bool, filesize_limit int, filters core.LineFilter, projectPaths ...string) (bool, error) {
	if len(projectPaths) == 0 {
		projectPaths = []string{"."}
	}

	global_result = &core.ScanResult{}
	global_result.Init()

	fmt.Println("Scanning paths:", projectPaths)
	time.Sleep(1 * time.Second)

	foundSecrets := false

	for _, path := range projectPaths {
		err := global_result.IterFolder(path, filters, global_gitignore, int64(filesize_limit))
		if err != nil {
			return foundSecrets, fmt.Errorf("scan failed for %s: %w", path, err)
		}
	}

	if global_result.IsFilenameMapEmpty() {
		return false, nil
	}

	global_result.PrettyPrintResults()

	saveRoot, err := filepath.Abs(".")
	if err != nil {
		return true, fmt.Errorf("failed to resolve save path: %w", err)
	}
	if logging == true {
		if err := core.SaveFilenameMap(saveRoot, global_result.GetFilenameMap()); err != nil {
			return true, fmt.Errorf("failed to save scan results: %w", err)
		}
	}

	return true, nil
}

func runObfuscate() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := core.LoadObfuscation(root); err != nil {
		return fmt.Errorf("failed to obfuscate secrets: %w", err)
	}

	fmt.Println("Secrets obfuscated successfully.")
	return nil
}