# Coauthor Cleaner

Remove unwanted AI coauthor trailers, generated-by comments, and attribution boilerplate from commits, staged changes, and pull requests.

> Clean commits. Quiet bots.

## Install

```bash
go install github.com/Skeyelab/coauthor-cleaner/cmd/coauthor-cleaner@latest
```

Or build from source:

```bash
CGO_ENABLED=0 go build -o coauthor-cleaner ./cmd/coauthor-cleaner
```

## Quick start

```bash
# Scan staged changes and HEAD commit
coauthor-cleaner scan

# Interactive review TUI
coauthor-cleaner review

# Install local git hooks (warn mode by default)
coauthor-cleaner hook install

# Create repo config
coauthor-cleaner config init
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | Report attribution markers (non-destructive) |
| `clean` | Remove markers (`--yes` to apply) |
| `review` | Interactive Bubble Tea TUI |
| `ci` | PR range scan for GitHub Actions |
| `pr scan` / `pr clean` | Scan/clean current PR via `gh` |
| `hook install` | Install pre-commit + commit-msg hooks |
| `config init` | Create `.coauthor-cleaner.yml` |
| `rules list` | Show active detection rules |

## Configuration

`.coauthor-cleaner.yml` in your repo root:

```yaml
providers:
  claude: true
  chatgpt: true

behavior:
  hook_mode: warn   # warn | block | clean

allowed_trailers:
  - "Co-authored-by: Your Name"
```

## GitHub Actions

Use the reusable action from this repo:

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
- uses: actions/setup-go@v5
  with:
    go-version-file: go.mod
- uses: Skeyelab/coauthor-cleaner/action.yml@v0.1.0
  with:
    mode: block   # or warn
```

Set repository variable `COAUTHOR_CLEANER_MODE` to `warn` or `block`.

## Development

```bash
go test ./...
CGO_ENABLED=0 go build -o coauthor-cleaner ./cmd/coauthor-cleaner
```

Release builds: `goreleaser release --snapshot --clean`
