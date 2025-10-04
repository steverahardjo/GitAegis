package main

import (
	//git "github.com/go-git/go-git/v6"
	"os/exec"
	"fmt"
)

func GitAdd(paths ...string) error {
	args := append([]string{"add"}, paths...)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput() // captures stdout + stderr
	if err != nil {
		return fmt.Errorf("git add failed: %v\n%s", err, string(out))
	}
	return nil
}


