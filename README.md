# winding

A CLI that bumps every tracked repo to the latest Yarn (via Corepack) and runs an install.

## What it does

`winding` keeps a list of repos you care about. Running it walks that list and, for each repo:

1. Runs `corepack use yarn@latest`, which rewrites `package.json`'s `packageManager` field to the latest Yarn release.
2. Runs `yarn install`.
3. Optionally commits the resulting changes to a new branch.

Repos are bumped concurrently, with a bounded worker pool, and a spinner shows progress while it runs.

## Requirements

- Go 1.27+ (to build)
- [Corepack](https://nodejs.org/api/corepack.html) on `PATH` (ships with Node 16.10+, or `npm install -g corepack`)
- Yarn and Git available in each tracked repo as applicable

## Install

```sh
go install github.com/kwaimind/winding@latest
```

Or build from source:

```sh
go build -o winding .
```

## Usage

### Track a repo

```sh
winding add <path>
```

The path must contain a `package.json`.

### List tracked repos

```sh
winding list   # alias: ls
```

Flags a tracked repo if its `package.json` has gone missing.

### Stop tracking a repo

```sh
winding remove <path>   # alias: rm
```

### Bump all tracked repos

```sh
winding
```

Flags:

| Flag | Default | Description |
| --- | --- | --- |
| `--git` | `false` | Commit each bumped repo's changes to a new branch |
| `--parallel`, `-j` | `4` | Number of repos to bump concurrently |
| `--repo` | | Only bump this repo (path or directory name) |

When `--git` is set, each repo with pending changes is committed to a new branch named after a short random SHA, with the message `winding: bump yarn to latest`. Repos with no changes, or that aren't git repos, are left alone.

### Bump a single repo

```sh
winding --repo <name>
```

`<name>` can be the full tracked path, a directory name, or a fuzzy substring of one — e.g. `winding --repo ssr` matches a tracked `/path/to/apoteket-ssr`. An exact path or directory-name match runs immediately; a fuzzy substring match asks for confirmation first, and an ambiguous substring (matching more than one tracked repo) lists the matches and asks you to be more specific. Combine with `--git` to also commit the change.

## Configuration

Tracked repos are stored in a human-editable YAML file:

- `$XDG_CONFIG_HOME/winding/config.yaml` if `XDG_CONFIG_HOME` is set
- `~/.config/winding/config.yaml` otherwise (including on macOS)

You can edit this file by hand, or use `winding add`/`winding remove`.

## Project layout

```
cmd/              cobra CLI commands (root, add, list, remove)
internal/config/  tracked-repo list, persisted as YAML
internal/yarnbump/ bumps a single repo's Yarn version and installs
internal/gitops/  commits a repo's changes to a new branch
internal/spinner/ terminal spinner used while bumping
```
