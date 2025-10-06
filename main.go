package main

import (
	"log"
	"fronte"
)

func main() {
	frontend.Init_cmd()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}