// Package gitops commits pending changes in a repo to a freshly created branch.
package gitops

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

// Result describes the outcome of committing changes in a repo.
type Result struct {
	Branch    string
	Committed bool
}

// CommitChanges creates a new branch named after a short random sha and
// commits any pending changes in path onto it. If path isn't a git repo or
// has no pending changes, it returns a zero Result and no error.
func CommitChanges(path string) (Result, error) {
	if _, err := run(path, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return Result{}, nil
	}

	status, err := run(path, "git", "status", "--porcelain")
	if err != nil {
		return Result{}, fmt.Errorf("git status failed: %w", err)
	}
	if status == "" {
		return Result{}, nil
	}

	branch, err := randomSHA()
	if err != nil {
		return Result{}, err
	}

	if out, err := run(path, "git", "checkout", "-b", branch); err != nil {
		return Result{}, fmt.Errorf("git checkout -b %s failed: %w\n%s", branch, err, out)
	}

	if out, err := run(path, "git", "add", "-A"); err != nil {
		return Result{}, fmt.Errorf("git add -A failed: %w\n%s", err, out)
	}

	if out, err := run(path, "git", "commit", "-m", "winding: bump yarn to latest"); err != nil {
		return Result{}, fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return Result{Branch: branch, Committed: true}, nil
}

func randomSHA() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:7], nil
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}
