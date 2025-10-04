package main

import (
	"GitAegis/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	cobra "github.com/spf13/cobra"
)

var (
	result   *core.ScanResult
	entLimit float64
)

var rootCmd = &cobra.Command{
	Use:   "gitaegis",
	Short: "API key scanner in Go",
	Long:  "Lightweight API key scanner using entropy and tree-sitter in Golang",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		result = &core.ScanResult{}
		result.Init()
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
        logging, _ := cmd.Flags().GetBool("logging")

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			log.Fatal("Unable to resolve absolute path:", err)
		}

		fmt.Println("START SCANNING...")
		fmt.Println("Target path:", absPath)

		found, err := Scan(entLimit, logging, absPath)
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

func Add(logging bool, paths ...string) error {
	secretsFound, err := Scan(5.0, logging, paths...)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	if secretsFound {
		return fmt.Errorf("secrets detected! aborting add")
	}
	for _, f := range paths {
		if err := GitAdd(f); err != nil {
			return err
		}
	}
	return nil
}

func Scan(entropyLimit float64, logging bool, projectPaths ...string) (bool, error) {
	if len(projectPaths) == 0 {
		projectPaths = []string{"."}
	}

	result = &core.ScanResult{}
	result.Init()

	fmt.Println("Scanning paths:", projectPaths)
	time.Sleep(1 * time.Second)

	filters := core.AllFilters(
		core.BasicFilter(),
		core.EntropyFilter(entropyLimit),
	)

	foundSecrets := false

	for _, path := range projectPaths {
		err := result.IterFolder(path, filters)
		if err != nil {
			return foundSecrets, fmt.Errorf("scan failed for %s: %w", path, err)
		}
	}

	if result.IsFilenameMapEmpty() {
		return false, nil
	}

	result.PrettyPrintResults()

	saveRoot, err := filepath.Abs(".")
	if err != nil {
		return true, fmt.Errorf("failed to resolve save path: %w", err)
	}
	if logging == true{
	if err := core.SaveFilenameMap(saveRoot, result.GetFilenameMap()); err != nil {
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

	fmt.Println("✅ Secrets obfuscated successfully.")
	return nil
}

func Init_cmd() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().Float64VarP(&entLimit, "ent_limit", "e", 5.0, "Entropy threshold for secret detection")
	scanCmd.Flags().Bool("logging", false, "log")
	rootCmd.AddCommand(gitignoreCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(obfuscateCmd)
}
