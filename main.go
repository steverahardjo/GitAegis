package main

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
)

func main() {
	files := core.GetUntrackedFile(".")
	for _, f := range files {
		fmt.Println("Untracked:", f)
	}
}

