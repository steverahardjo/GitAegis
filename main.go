package main

import (
    "fmt"
    "os"
    frontend "github.com/steverahardjo/GitAegis/frontend"
)

func main() {
    frontend.Init_cmd() // register commands & flags

    // Execute the CLI
    if err := frontend.RootCmd().Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
