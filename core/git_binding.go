package core

import (
	"log"

	"github.com/go-git/go-git/v5"
)

func Init(path string) *git.Repository {
	repo, err := git.PlainOpen(path)
	if err != nil {
		log.Fatalf("[go-git] failed to open repo: %v", err)
	}
	return repo
}

func GetUntrackedFile(path string) []string {
	repo := Init(path)
	var files []string

	w, err := repo.Worktree()
	if err != nil {
		log.Printf("[go-git] worktree error: %v", err)
		return files
	}

	status, err := w.Status()
	if err != nil {
		log.Printf("[go-git] status error: %v", err)
		return files
	}

	for file, s := range status {
		if s.Worktree == git.Untracked {
			files = append(files, file)
		}
		if s.Worktree == git.Modified{
			files = append(files, file)
		}
	}

	return files
}