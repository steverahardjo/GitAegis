package frontend

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
	"log"
	"os"
	"path/filepath"

	cobra "github.com/spf13/cobra"
)


var rootCmd = &cobra.Command{
	Use:   "gitaegis",
	Short: "API key scanner in Go",
	Long:  "Lightweight API key scanner using entropy and tree-sitter in Golang",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		global_result = &core.ScanResult{}
		global_result.Init()
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a directory or file for secrets",
	Long:  "Scan a specified directory or file for secrets using entropy and regex for API keys. Defaults to the current directory if no path is provided.",
	Run: func(cmd *cobra.Command, args []string) {
		var targetPath string

		// Use provided path or fallback to current dir
		if len(args) > 0 {
			targetPath = args[0]
		} else {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal("Unable to get current working directory:", err)
			}
			targetPath = wd
		}
		global_logging, _ := cmd.Flags().GetBool("logging")

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			log.Fatal("Unable to resolve absolute path:", err)
		}

		LazyInitConfig()

		fmt.Println("START SCANNING...")
		fmt.Println("Target path:", absPath)

		found, err := Scan(global_entLimit, global_logging, global_git_integration, int(global_filemaxsize), global_filters,absPath)
		if err != nil {
			log.Fatal(err)
		}

		if found {
			fmt.Println("\nSecrets detected! You may run `gitaegis obfuscate` to mask them.")
		} else {
			fmt.Println("\nNo secrets found. Nothing to obfuscate.")
		}
	},
}

var gitignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Generate or update .gitignore from previous scan run",
	Run: func(cmd *cobra.Command, args []string) {
		if err := core.UpdateGitignore(); err != nil {
			log.Fatal(err)
		}
	},
}

var obfuscateCmd = &cobra.Command{
	Use:   "obfuscate",
	Short: "Obfuscate detected secrets in the codebase",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runObfuscate(); err != nil {
			log.Fatal(err)
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Scan before a git add",
	Long:  "Couple GitAegis with git add to block commits containing secrets.",
	Run: func(cmd *cobra.Command, args []string) {
		logging, _ := cmd.Flags().GetBool("logging")
		if err := Add(logging, args...); err != nil {
			log.Fatal(err)
		}
	},
}



var sitter = &cobra.Command{
	Use:   "sitter",
	Short: "Integrate tree-sitter grammars to be used as parser in GitAegis",
	Long:  "Integrating local tree-sitter grammar into gitaegis (ideal if user use nvim/helix/zed)",
	Run: func(cmd *cobra.Command, args []string) {
		logging, _ := cmd.Flags().GetBool("logging")
		if err := Add(logging, args...); err != nil {
			log.Fatal(err)
		}
	},
}

func Init_cmd() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().Float64VarP(&global_entLimit, "ent_limit", "e", 5.0, "Entropy threshold for secret detection")
	scanCmd.Flags().Bool("logging", false, "log")
	rootCmd.AddCommand(gitignoreCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(obfuscateCmd)
}
