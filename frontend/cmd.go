package frontend

import (
	"fmt"
	"os"
	"path/filepath"

	core "github.com/steverahardjo/gitaegis/core"
	intro "github.com/steverahardjo/gitaegis/intro"

	cobra "github.com/spf13/cobra"
)

var rv *RuntimeValue

// Root command
var rootCmd = &cobra.Command{
	Use:   "gitaegis",
	Short: "gitaegis CLI tool",
	Long:  "Scan recursively searches files for potential secrets like API keys, tokens, and credentials using entropy analysis and pattern matching.",

	Run: func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println("gitaegis version 1.0")
			return
		}
		fmt.Println("Run 'gitaegis --help' for usage.")
	},
}

var scanCmd = &cobra.Command{
	Use:     "scan [path]",
	Short:   "Scan a directory or file for secrets",
	Long:    `Scan recursively searches files for potential secrets like API keys, tokens, and credentials using entropy analysis and pattern matching.`,
	Example: "gitaegis scan -l -e 4.0 ./myapp",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rv == nil {
			rv = NewRuntimeConfig()
		}

		var targetPath string
		if len(args) > 0 {
			targetPath = args[0]
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("unable to get current working directory: %w", err)
			}
			targetPath = wd
		}

		rv.LoggingEnabled, _ = cmd.Flags().GetBool("logging")
		LazyInitConfig()

		absPath, _ := filepath.Abs(targetPath)
		fmt.Println("START SCANNING...")
		fmt.Println("Target path:", absPath)

		found, err := rv.Scan(absPath)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if found {
			fmt.Println("\nSecrets detected!")
			os.Exit(1)
		}
		fmt.Println("\nNo secrets found.")
		return nil
	},
}

var gitignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Generate/update .gitignore from previous scan.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rv == nil {
			rv = NewRuntimeConfig()
		}
		blob, err := core.LoadFilenameMap(".")
		if err != nil {
			return fmt.Errorf("unable to load scan results: %w", err)
		}
		if err := core.UpdateGitignore(blob); err != nil {
			return fmt.Errorf("failed to update .gitignore: %w", err)
		}
		return nil
	},
}

//var obfuscateCmd = &cobra.Command{
//Use:   "obfuscate",
//Short: "Obfuscate detected secrets in the codebase",
//Run: func(cmd *cobra.Command, args []string) {
//if rv == nil { rv = NewRuntimeConfig() }
// Uncomment when implemented
// if err := rv.runObfuscate(); err != nil { log.Fatal(err) }
//},
//}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Scan before git add.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rv == nil {
			rv = NewRuntimeConfig()
		}
		rv.LoggingEnabled, _ = cmd.Flags().GetBool("logging")
		gitPath, _ := os.Getwd()
		if err := rv.Add(gitPath, args...); err != nil {
			return fmt.Errorf("git add failed: %w", err)
		}
		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Shortcut to uninstall GitAegis from your shell.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return UninstallSelf()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Integrate gitaegis to pre-hook or bashrc",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		preHook, _ := cmd.Flags().GetBool("prehook")
		bash, _ := cmd.Flags().GetBool("bash")
		if preHook && bash {
			return fmt.Errorf("flags --prehook and --bash cannot be used together")
		} else if bash {
			intro.AttachShellConfig()
		} else {
			if err := intro.GitPreHookInit(root); err != nil {
				return fmt.Errorf("failed to init pre-hook: %w", err)
			}
		}
		return nil
	},
}

// Init_cmd registers commands and flags
func Init_cmd() *cobra.Command {
	rv = NewRuntimeConfig()

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	scanCmd.Flags().Float64VarP(&rv.EntropyLimit, "ent_limit", "e", rv.EntropyLimit, "Entropy threshold for secret detection")
	scanCmd.Flags().BoolP("logging", "l", false, "Enable logging")

	initCmd.Flags().Bool("prehook", false, "Integrate gitaegis as git pre-hook")
	initCmd.Flags().Bool("bash", false, "Integrate gitaegis into bashrc")

	rootCmd.AddCommand(scanCmd, gitignoreCmd, addCmd, initCmd, uninstallCmd)

	return rootCmd
}
