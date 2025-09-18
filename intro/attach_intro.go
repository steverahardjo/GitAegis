package intro

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var bashrcWrapper = `
# >>> Git Aegis wrapper >>>
function git() {
    if [[ "$1" == "add" ]]; then
        shift
        git-aegis add "$@"
    else
        command git "$@"
    fi
}
# <<< Git Aegis wrapper <<<
`
func detectAttachShellConfig() {
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

func main() {
	detectAttachShellConfig()
}
