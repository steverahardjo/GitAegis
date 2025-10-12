package frontend

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
	"log"
	"os"
	"path/filepath"
	intro "github.com/steverahardjo/GitAegis/intro"
	cobra "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitaegis",
	Short: "API key scanner in Go",
	Long:  "Lightweight API key scanner using entropy and tree-sitter in Golang",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rv = NewRuntimeConfig()
		rv.GlobalResult.Init()
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
		rv.LoggingEnabled, _ = cmd.Flags().GetBool("logging")

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			log.Fatal("Unable to resolve absolute path:", err)
		}

		LazyInitConfig()

		fmt.Println("START SCANNING...")
		fmt.Println("Target path:", absPath)

		found, err := rv.Scan(absPath)
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
		//if err := rv.runObfuscate(); err != nil {
			//log.Fatal(err)
		//}
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Scan before a git add",
	Long:  "Couple GitAegis with git add to block commits containing secrets.",
	Run: func(cmd *cobra.Command, args []string) {
		rv.LoggingEnabled, _ = cmd.Flags().GetBool("logging")
		if err := rv.Add(args...); err != nil {
			log.Fatal(err)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Integrate gitaegis to pre-hook or to bashrc",
	Long:  "pre-hook require the correct root you need, bashrc require it existss",
	Run: func(cmd *cobra.Command, args []string) {
		root, err := os.Getwd()
		if err != nil{
			log.Println(err)
		}
		preHook, _ := cmd.Flags().GetBool("prehook")
		bash, _ := cmd.Flags().GetBool("bash")
		if preHook && bash {
			log.Fatal("Flags --prehook and --bash cannot be used together")
		}else if bash{
			intro.AttachShellConfig()
		}else{
			intro.GitPreHookInit(root)
		}

	},
}

func Init_cmd() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().Float64VarP(&rv.EntropyLimit, "ent_limit", "e", 5.0, "Entropy threshold for secret detection")
	scanCmd.Flags().Bool("logging", false, "log")
	rootCmd.AddCommand(gitignoreCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(obfuscateCmd)
	initCmd.Flags().Bool("prehook", false, "Integrate GitAegis as a git pre-hook")
	initCmd.Flags().Bool("bash", false, "Integrate GitAegis into bashrc")
	rootCmd.AddCommand(initCmd)
}
