// file created to string all core func and everything together as a cli cmd
package main

import (
	"fmt"
	"log"
	"os"

	"GitAegis/core"

	cobra "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitaegis",
	Short: "API key scanner in Go",
	Long:  "Lightweight API key scanner using entropy and tree-sitter in Golang",
}

var entLimit float64

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the current directory for secrets",
	Long:  "Scan the current directory for secrets using entropy and regex for api key",
	Run: func(cmd *cobra.Command, args []string) {
		found, err := Scan(entLimit)
		if err != nil {
			log.Fatal(err)
		}

		if found {
			fmt.Println("\n Secrets detected! You may run `gitaegis obfuscate` to mask them.")
		} else {
			fmt.Println("\n No secrets found. Nothing to obfuscate.")
		}
	},
}

var gitignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Generate or update .gitignore from previous scan run",
	Long:  "Generate or update .gitignore from previous scan run",
	Run: func(cmd *cobra.Command, args []string) {
		if err := core.UpdateGitignore(); err != nil {
			log.Fatal(err)
		}
	},
}

var obfuscateCmd = &cobra.Command{
	Use:   "obfuscate",
	Short: "Obfuscate detected secrets in the codebase",
	Long:  "Obfuscate detected secrets in the codebase by replacing them with placeholders",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runObfuscate(); err != nil {
			log.Fatal(err)
		}
	},
}

var ExemptAdditor = &cobra.Command{
	Use:   "add_exempt [files...]",
	Short: "Add exemptions to the program",
	Long:  "Add exemptions to the program so certain files will be ignored from scanning",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("\n No files specified. Usage: add_exempt <file1> <file2> ...")
			return
		}

		for _, file := range args {
			core.AddExempt(file)
		}

		fmt.Println("\n Current exemption list:", core.Exempt)
	},
}

// Scan runs the secret detection on the current working directory.
// It returns true if secrets were found, otherwise false.
func Scan(entrophy_limit float64) (bool, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get working directory: %w", err)
	}
	filters := core.AllFilters(
		core.EntropyFilter(entrophy_limit),
		core.RegexFilter(),
	)

	// Run folder iteration
	results, err := core.IterFolder(projectPath, filters)
	if err != nil {
		return false, fmt.Errorf("scan failed: %w", err)
	}

	if core.IsFilenameMapEmpty(results) {
		log.Println("No secrets found")
		return false, nil
	}

	core.PrettyPrintResults(results)
	if err := core.SaveFilenameMap(projectPath, results); err != nil {
		return true, fmt.Errorf("failed to save scan results: %w", err)
	}
	return true, nil
}

// runObfuscate replaces secret lines with placeholder text.
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

func init() {
	// add scanCmd to root
	rootCmd.AddCommand(scanCmd)
	// add --ent_limit (default 5.0)
	scanCmd.Flags().Float64VarP(&entLimit, "ent_limit", "e", 5.0, "Entropy threshold for secret detection")
	//add gitignoreCmd to root
	rootCmd.AddCommand(gitignoreCmd)
	//add obfuscateCmd to root
	rootCmd.AddCommand(obfuscateCmd)
	//add ExemptAdditor to root
	rootCmd.AddCommand(ExemptAdditor)
}


