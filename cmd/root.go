// Package cmd implements winding's CLI commands.
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kwaimind/winding/internal/config"
	"github.com/kwaimind/winding/internal/gitops"
	"github.com/kwaimind/winding/internal/spinner"
	"github.com/kwaimind/winding/internal/yarnbump"
	"github.com/spf13/cobra"
)

var (
	gitFlag      bool
	parallelFlag int
	repoFlag     string
)

var rootCmd = &cobra.Command{
	Use:   "winding",
	Short: "Bump every tracked repo to the latest Yarn and install",
	Long: fmt.Sprintf(`winding bumps every tracked repo to the latest Yarn (via Corepack) and
runs an install.

Config file: %s (edit by hand anytime, or use the subcommands below)`, mustConfigPath()),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAll()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&gitFlag, "git", false, "commit each bumped repo's changes to a new branch")
	rootCmd.Flags().IntVarP(&parallelFlag, "parallel", "j", 4, "number of repos to bump concurrently")
	rootCmd.Flags().StringVar(&repoFlag, "repo", "", "only bump this repo (path or directory name)")
}

// Execute runs the root command, exiting non-zero on failure.
func Execute() error {
	return rootCmd.Execute()
}

// selectRepo resolves want to one of the tracked repos. An exact path or
// directory-name match is used as-is. Otherwise it falls back to a
// case-insensitive substring match against each repo's directory name (so
// `--repo ssr` finds `apoteket-ssr`) and asks for confirmation before
// proceeding, since that match was inferred rather than typed exactly.
func selectRepo(repos []string, want string) (string, error) {
	if abs, err := filepath.Abs(want); err == nil {
		for _, repo := range repos {
			if repo == abs {
				return repo, nil
			}
		}
	}

	for _, repo := range repos {
		if filepath.Base(repo) == want {
			return repo, nil
		}
	}

	wantLower := strings.ToLower(want)
	var matches []string
	for _, repo := range repos {
		if strings.Contains(strings.ToLower(filepath.Base(repo)), wantLower) {
			matches = append(matches, repo)
		}
	}
	switch len(matches) {
	case 1:
		ok, err := confirm(fmt.Sprintf("bump %s?", matches[0]))
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("cancelled")
		}
		return matches[0], nil
	case 0:
		return "", fmt.Errorf("no tracked repo matches %q (run `winding list` to see tracked repos)", want)
	default:
		return "", fmt.Errorf("%q matches multiple tracked repos: %v (be more specific)", want, matches)
	}
}

// confirm asks the user a yes/no question on stdin, defaulting to no.
func confirm(question string) (bool, error) {
	fmt.Printf("%s [y/N] ", question)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func mustConfigPath() string {
	path, err := config.Path()
	if err != nil {
		return "~/.config/winding/config.yaml"
	}
	return path
}

func runAll() error {
	repos, err := config.Repos()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("no repos tracked yet — run `winding add <path>` to add one")
		return nil
	}

	if repoFlag != "" {
		repo, err := selectRepo(repos, repoFlag)
		if err != nil {
			return err
		}
		repos = []string{repo}
	}

	concurrency := parallelFlag
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(repos) {
		concurrency = len(repos)
	}

	outputs := make([]string, len(repos))
	failed := make([]bool, len(repos))

	message := fmt.Sprintf("bumping %d repo(s)...", len(repos))
	if concurrency > 1 {
		message = fmt.Sprintf("bumping %d repos (%d at a time)...", len(repos), concurrency)
	}
	sp := spinner.Start(message)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for i, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, repo string) {
			defer wg.Done()
			defer func() { <-sem }()
			outputs[i], failed[i] = bumpRepo(repo)
		}(i, repo)
	}
	wg.Wait()
	sp.Stop()

	failures := 0
	for i, out := range outputs {
		fmt.Print(out)
		if failed[i] {
			failures++
		}
	}

	fmt.Printf("\n%d/%d repos bumped and installed successfully\n", len(repos)-failures, len(repos))
	if failures > 0 {
		return fmt.Errorf("%d repo(s) failed", failures)
	}
	return nil
}

// bumpRepo bumps a single repo's Yarn version, optionally commits the
// result, and returns its full report and whether it failed. Output is
// buffered rather than printed directly so concurrent runs don't interleave
// on stdout.
func bumpRepo(repo string) (string, bool) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "==> %s\n", repo)

	result := yarnbump.Bump(repo)
	if !result.OK {
		fmt.Fprintf(&buf, "    FAILED: %s\n", result.Message)
		return buf.String(), true
	}
	fmt.Fprintf(&buf, "    ok: %s\n", result.Message)

	if gitFlag {
		commitResult, err := gitops.CommitChanges(repo)
		switch {
		case err != nil:
			fmt.Fprintf(&buf, "    git: FAILED: %v\n", err)
		case commitResult.Committed:
			fmt.Fprintf(&buf, "    git: committed to branch %s\n", commitResult.Branch)
		default:
			fmt.Fprintln(&buf, "    git: nothing to commit")
		}
	}

	return buf.String(), false
}
