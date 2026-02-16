package frontend

import (
	"fmt"
	"log"
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
	Run: func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println("gitaegis version 1.0")
			return
		}
		fmt.Println("Run 'gitaegis --help' for usage.")
	},
}

// Scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a directory or file for secrets",
	Long:  "Scan a specified directory or file for secrets using entropy and regex for API keys. Defaults to current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		if rv == nil {
			rv = NewRuntimeConfig()
		}

		var targetPath string
		if len(args) > 0 {
			targetPath = args[0]
		} else {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal("Unable to get current working directory:", err)
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
			log.Fatal(err)
		}
		if found {
			fmt.Println("\nSecrets detected! Run `gitaegis obfuscate` to mask them.")
		} else {
			fmt.Println("\nNo secrets found.")
		}
	},
}

// Other commands (gitignore, add, obfuscate, init)
var gitignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Generate/update .gitignore from previous scan",
	Run: func(cmd *cobra.Command, args []string) {
		if rv == nil {
			rv = NewRuntimeConfig()
		}
		blob, err := core.LoadFilenameMap(".")
		if err != nil {
			log.Printf("[mod] unable to load JSON log files")
		}
		if err := core.UpdateGitignore(blob); err != nil {
			log.Fatal(err)
		}
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
	Short: "Scan before git add",
	Run: func(cmd *cobra.Command, args []string) {
		if rv == nil {
			rv = NewRuntimeConfig()
		}
		rv.LoggingEnabled, _ = cmd.Flags().GetBool("logging")
		gitPath, _ := os.Getwd()
		if err := rv.Add(gitPath, args...); err != nil {
			log.Fatal(err)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Integrate gitaegis to pre-hook or bashrc",
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := os.Getwd()
		preHook, _ := cmd.Flags().GetBool("prehook")
		bash, _ := cmd.Flags().GetBool("bash")
		if preHook && bash {
			fmt.Println("Flags --prehook and --bash cannot be used together")
			return
		} else if bash {
			intro.AttachShellConfig()
		} else {
			intro.GitPreHookInit(root)
		}
	},
}

// Init_cmd registers commands and flags
func Init_cmd() *cobra.Command {
	rv = NewRuntimeConfig()

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	scanCmd.Flags().Float64VarP(&rv.EntropyLimit, "ent_limit", "e", rv.EntropyLimit, "Entropy threshold for secret detection")
	scanCmd.Flags().Bool("logging", false, "Enable logging")

	initCmd.Flags().Bool("prehook", false, "Integrate gitaegis as git pre-hook")
	initCmd.Flags().Bool("bash", false, "Integrate gitaegis into bashrc")

	rootCmd.AddCommand(scanCmd, gitignoreCmd, addCmd, initCmd)

	return rootCmd
}
