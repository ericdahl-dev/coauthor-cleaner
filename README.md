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
coauthor-cleaner              # same as: coauthor-cleaner fix
coauthor-cleaner fix --push   # clean + push when safe
```

Already pushed to GitHub? `coauthor-cleaner fix --force --force-push`

```bash
coauthor-cleaner config init
coauthor-cleaner hook install
# set behavior.hook_mode: clean  → auto-fix on git commit
```

## Commands

| Command | Description |
|---------|-------------|
| `fix` | **Default** — auto-clean safe findings (`--push`, `--force`, `--check`) |
| `status` | Repo check + findings + next steps |
| `doctor` | Verify hooks, config, gh, git repo |
| `scan` | Report markers (`--file` / `--dir` work outside git) |
| `clean` | Remove markers (`--yes`; `--force` if commit was pushed) |
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
