//file created to string all core func and everything together as a cli cmd
package main

import(
	cobra "github.com/spf13/cobra"
	"os"
	"log"
	"GitAegis/core"
)

var helloCmd = &cobra.Command{
	Use: "gitaegis",
	Short: "API key scannerin go",
	Long: "Lightweight API key scanner using entrophy and tree-sitter in golang"
}

func Scan(){

	projectPath, err := os.Getwd()
	if err != nil{
		log.Fatal("Unable to detect the current path project is in")
	}
	var DEF_ENTR:float = 5.0
	filters := core.AllFilters(
		core.EntropyFilter(DEF_ENTR),
		core.RegexFilter()
	)
	// Run folder iteration
	results, err := core.IterFolder(projectPath, filters)
	// Pretty print the results
	core.PrettyPrintResults(results)
}
