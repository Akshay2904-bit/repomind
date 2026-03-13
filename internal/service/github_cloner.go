package service

import (
	"os"
	"fmt"
	"strings"
    git "github.com/go-git/go-git/v5")

type CloneResult struct {
	Dir string
	Cleanup func()
}

func CloneGitHubRepo(RepoURL string) (*CloneResult, error) {
	if !strings.HasPrefix(RepoURL, "https://github.com/") &&
		!strings.HasPrefix(RepoURL, "https://www.github.com/") {
	return nil, fmt.Errorf("not a GitHub URL: %s", RepoURL)
}
	tmpDir, err := os.MkdirTemp("", "repomind-clone-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	_,err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: RepoURL,
		Depth: 1,
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to clone repo: %w",RepoURL, err)
	}
	return &CloneResult{
		Dir: tmpDir,
		Cleanup: func() {
			os.RemoveAll(tmpDir)
		},
	}, nil
}

func IsGitHubURL(s string) bool {
    return strings.HasPrefix(s, "https://github.com/") ||
        strings.HasPrefix(s, "http://github.com/")
}


	