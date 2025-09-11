package main

import (
	"fmt"
)

func main() {
	result, err := iterFolder(".") // scan current folder
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for file, lines := range result {
		fmt.Printf("File: %s (total %d lines)\n", file, len(lines))
		for _, l := range lines {
			fmt.Printf("  %d: %s\n", l.index, l.line)
		}
	}
}
