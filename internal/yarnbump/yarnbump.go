// Package yarnbump bumps a repo's pinned Yarn version to the latest via
// Corepack and runs an install.
package yarnbump

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Result struct {
	Path    string
	OK      bool
	Message string
}

// Bump updates the repo at path to the latest Yarn (via `corepack use
// yarn@latest`, which rewrites package.json's "packageManager" field) and
// then runs `yarn install`.
func Bump(path string) Result {
	pkgJSON := filepath.Join(path, "package.json")
	if _, err := os.Stat(pkgJSON); err != nil {
		return Result{Path: path, Message: "no package.json found at " + pkgJSON}
	}

	if _, err := exec.LookPath("corepack"); err != nil {
		return Result{Path: path, Message: "corepack not found on PATH (needs Node 16.10+, or run `npm install -g corepack`)"}
	}

	if out, err := run(path, "corepack", "use", "yarn@latest"); err != nil {
		return Result{Path: path, Message: fmt.Sprintf("corepack use yarn@latest failed: %v\n%s", err, out)}
	}

	if out, err := run(path, "yarn", "install"); err != nil {
		return Result{Path: path, Message: fmt.Sprintf("yarn install failed: %v\n%s", err, out)}
	}

	return Result{Path: path, OK: true, Message: "bumped to latest yarn and installed"}
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
