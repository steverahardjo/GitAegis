package frontend

import (
	"fmt"
	"path/filepath"
	"time"
	"os/exec"
	"os"
	"strings"
	core "github.com/steverahardjo/gitaegis/core"
)

// Add will scan given paths and git-add them if no secrets are found.
func (rv *RuntimeValue) Add(repoPath string, paths ...string) error {
	// perform scan using rv configuration
	secretsFound, err := rv.Scan(paths...)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	if secretsFound {
		return fmt.Errorf("secrets detected! aborting add")
	}

	for _, f := range paths {
		if err := core.GitAdd(repoPath, f); err != nil {
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

	if rv.Result == nil {
		rv.Result = &core.ScanResult{}
	}

	fmt.Println("Scanning paths:", projectPaths)
	time.Sleep(1 * time.Second)
	rv.SetTreeSitterPath("/home/holyknight101/.private/helix/runtime/grammars")
	core.IntegrateTreeSitter(rv.TreeSitterPath)
	filter := core.AnyFilters(
		rv.Filters,
		core.EntropyFilter(rv.EntropyLimit),
	)
	for _, path := range projectPaths {
		if rv.GitDiffScan {
			
			files := core.GetUntrackedFile(path)
			if err := rv.Result.IterFiles(files, filter, int64(rv.MaxFileSize)); err != nil {
				return false, fmt.Errorf("scan failed for files in %s: %w", path, err)
			}
		} else {
			if err := rv.Result.IterFolder(path, filter, rv.UseGitignore, int64(rv.MaxFileSize)); err != nil {
				return false, fmt.Errorf("scan failed for %s: %w", path, err)
			}
		}
	}
	res := rv.Result.IsFilenameMapEmpty()
	if res {
		fmt.Println("Nothing is found in the fileMap")
		return false, nil
	}
	rv.Result.PrettyPrintResults()

	saveRoot, err := filepath.Abs(".")
	if err != nil {
		return true, fmt.Errorf("[runner.Scan] failed to resolve save path: %w", err)
	}

	if rv.LoggingEnabled && !res {
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

func UninstallSelf() error {
	pathBytes, err := exec.Command("which", "gitaegis").Output()
	if err != nil {
		return fmt.Errorf("could not find gitaegis in PATH: %w", err)
	}

	binPath := strings.TrimSpace(string(pathBytes))

	cleanCmd := "sed -i '' '/gitaegis/d' ~/.bashrc ~/.zshrc 2>/dev/null || true"
	exec.Command("bash", "-c", cleanCmd).Run()

	fmt.Printf("Deleting binary at: %s\n", binPath)
	return os.Remove(binPath)
}
