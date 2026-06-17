# Coauthor Cleaner

Remove unwanted AI coauthor trailers, generated-by comments, and attribution boilerplate from commits, staged changes, and pull requests.

> Clean commits. Quiet bots.

## Install

```bash
go install github.com/ericdahl-dev/coauthor-cleaner/cmd/coauthor-cleaner@latest
```

Or build from source:

```bash
CGO_ENABLED=0 go build -o coauthor-cleaner ./cmd/coauthor-cleaner
```

## Quick start

```bash
coauthor-cleaner    # opens the TUI — scan, review, clean, and push
```

Lazygit-style split-panel TUI: findings list (left), live preview (right), status bar, contextual footer. Full workflow: toggle, clean, push (`--force-with-lease` when needed). Press `?` for help.

Non-interactive: `coauthor-cleaner fix --push`

## Protection Setup

Prevent AI attribution markers from being committed in the first place:

**Locally (git hooks):**
```bash
cp .coauthor-cleaner.protect.yml .coauthor-cleaner.yml
coauthor-cleaner hook install
# Commits with AI markers will now be blocked
```

**In CI (GitHub Actions):**
The workflow `.github/workflows/protect.yml` automatically blocks PRs with AI markers. No setup needed—it runs on every PR.

Or add to your workflow:
```yaml
- uses: ericdahl-dev/coauthor-cleaner/action.yml@v0.1.3
  with:
    mode: block   # Fail if any AI markers found
```

## Commands

| Command | Description |
|---------|-------------|
| *(default)* | **TUI** — scan, preview, clean, push (`review` / `tui` alias) |
| `fix` | Non-interactive auto-clean (`--push`, `--force`, `--check`) |
| `status` | Repo check + findings + next steps |
| `doctor` | Verify hooks, config, gh, git repo |
| `scan` | Report markers (`--file` / `--dir` work outside git) |
| `clean` | Remove markers (`--yes`; `--force` if commit was pushed) |
| `ci` | PR range scan for GitHub Actions |
| `pr scan` / `pr clean` | Scan/clean current PR via `gh` |
| `hook install` | Install pre-commit + commit-msg hooks |
| `config init` | Create `.coauthor-cleaner.yml` |
| `rules list` | Show active detection rules |

## Configuration

`.coauthor-cleaner.yml` in your repo root. See `.coauthor-cleaner.yml.example` or `.coauthor-cleaner.protect.yml` (strict mode).

**Hook modes:**
- `warn` — Alert on findings; commit allowed (default)
- `block` — Fail commit if markers found; user must clean first
- `clean` — Auto-remove markers; commit proceeds

```yaml
providers:
  claude: true
  chatgpt: true

behavior:
  hook_mode: block   # warn | block | clean

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
- uses: ericdahl-dev/coauthor-cleaner/action.yml@v0.1.3
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
