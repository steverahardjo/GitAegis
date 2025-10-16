package main

import (
	"fmt"
	runner "github.com/steverahardjo/GitAegis/frontend"
)

func main() {
	rv := runner.NewRuntimeConfig()

	// Initialize filters (regex + basic + entropy)
	rv.SetFilters(map[string]string{
		"password": `(?i)password\s*[:=]`,
		"apikey":   `(?i)api[_-]?key\s*[:=]`,
	})

	// Now scan
	res, err := rv.Scan("/home/holyknight101/Documents/Projects/Personal/exp_site")
	if err != nil {
		fmt.Println("Scan error:", err)
		return
	}
	print(res)

}
