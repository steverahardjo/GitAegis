package core

import (
	git "github.com/go-git/go-git/v5"
	"log"
)

func GitAdd(repoPath string, paths ...string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	for _, p := range paths {
		_, err := w.Add(p)
		if err != nil {
			return err
		}
	}
	return nil
}


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

