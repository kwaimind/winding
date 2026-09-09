// Package gitops commits pending changes in a repo to a freshly created branch.
package gitops

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result describes the outcome of committing changes in a repo.
type Result struct {
	Branch    string
	Committed bool
}

// yarnPaths are the files a Yarn bump can touch. Only these are inspected
// and staged, so pre-existing unrelated changes in the working tree are
// never swept into the bump commit.
var yarnPaths = []string{"package.json", "yarn.lock", ".yarn", ".yarnrc.yml"}

// CommitChanges creates a new branch named after a short random sha and
// commits pending Yarn-bump changes (package.json, yarn.lock, .yarn/,
// .yarnrc.yml) in path onto it. If path isn't a git repo, or the bump left
// none of those paths changed, it returns a zero Result and no error —
// even if the working tree has unrelated pending changes.
func CommitChanges(path string) (Result, error) {
	if _, err := run(path, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return Result{}, nil
	}

	paths := existingPaths(path, yarnPaths)
	if len(paths) == 0 {
		return Result{}, nil
	}

	status, err := run(path, "git", append([]string{"status", "--porcelain", "--"}, paths...)...)
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

	if out, err := run(path, "git", append([]string{"add", "-A", "--"}, paths...)...); err != nil {
		return Result{}, fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	if out, err := run(path, "git", "commit", "-m", "winding: bump yarn to latest"); err != nil {
		return Result{}, fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return Result{Branch: branch, Committed: true}, nil
}

// existingPaths returns the subset of candidates (relative to repo) that
// currently exist on disk. A pathspec git doesn't recognize on disk causes
// `git add`/`git status` to fail outright, so untouched paths (e.g. no
// .yarnrc.yml in a non-Berry repo) must be filtered out first.
func existingPaths(repo string, candidates []string) []string {
	var found []string
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(repo, c)); err == nil {
			found = append(found, c)
		}
	}
	return found
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
