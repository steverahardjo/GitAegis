package intro

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// config to add bash script to run Scan before doing add
var bashrcWrapper = `
# >>> Git Aegis wrapper >>>
function git() {
    if [[ "$1" == "add" ]]; then
        shift
        gitaegis add "$@"
    else
        command git "$@"
    fi
}
# <<< Git Aegis wrapper <<<
`
func AttachShellConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Unable to find a home directory to attach to a shell config file (.bashrc)")
		return
	}

	bashrc := filepath.Join(homeDir, ".bashrc")

	if _, err := os.Stat(bashrc); err == nil {
		fmt.Println("Found a .bashrc file")

		data, err := os.ReadFile(bashrc)
		if err != nil {
			fmt.Println("Unable to read .bashrc file")
			return
		}

		content := string(data)
		if strings.Contains(content, "git-aegis") {
			fmt.Println("Git Aegis is already attached to .bashrc file")
			return
		}

		f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println("Unable to open .bashrc file for writing")
			return
		}
		defer f.Close()

		if _, err := f.WriteString("\n\n" + bashrcWrapper + "\n"); err != nil {
			fmt.Println("Error writing to .bashrc file:", err)
			return
		}
		cmd := exec.Command("source", "~/.bashrc")
		cmd.Run()
		fmt.Println("Successfully attached Git Aegis to .bashrc file")
	} else {
		fmt.Println("No .bashrc file found in home directory")
	}
}
//enable gitaegis scan . as git hook based on  the flow
func GitPreHookInit(root string) error {
	hooksDir := filepath.Join(root, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("git hooks directory not found: %s", hooksDir)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Pre-commit hook content
	hookContent := `#!/bin/sh
# gitaegis pre-commit hook
echo "Running gitaegis scan..."
gitaegis scan . 
RESULT=$?
if [ $RESULT -ne 0 ]; then
  echo "GitAegis scan failed. Commit aborted."
  exit 1
fi
`
	// Write the hook file
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write pre-commit hook: %v", err)
	}

	fmt.Println("GitAegis pre-commit hook installed successfully.")
	return nil
}
