package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Resolve the full path for safety (tilde doesn't expand automatically in Go)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to get home directory:", err)
	}
	projectPath := filepath.Join(home, "Documents", "Projects", "Personal", "patent-analyser-fyp")
	var entLimit float64 = 5.0
	found, err := Scan(entLimit, projectPath)
	if err != nil {
		log.Fatal(err)
	}
	println(found)
	fmt.Println("Scan complete!")
}