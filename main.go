package main

import (
	"log"
)

func main() {
	Init_cmd()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
