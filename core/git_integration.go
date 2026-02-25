package core

import (
	"fmt"
	"log"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

type DiffMetadata struct{
	filename string
	start int
	end int
}

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

func GitWorkingStatus(repoPath string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	var files []string
	for file := range status {
		files = append(files, file)
	}
	fmt.Println(files)
	return files, nil
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
		absPath, err := filepath.Abs(filepath.Join(path, file))
		if err != nil {
			log.Printf("[go-git] failed to get absolute path for %s: %v", file, err)
			continue
		}

		if s.Worktree == git.Untracked || s.Worktree == git.Modified {
			fmt.Println("File in Status", file)
			files = append(files, absPath)
		}
	}

	return files
}
