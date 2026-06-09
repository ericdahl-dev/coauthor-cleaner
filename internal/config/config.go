package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = ".coauthor-cleaner.yml"

type Config struct {
	Providers       Providers `yaml:"providers"`
	Targets         Targets   `yaml:"targets"`
	Behavior        Behavior  `yaml:"behavior"`
	AllowedTrailers []string  `yaml:"allowed_trailers"`
}

type Providers struct {
	Claude    bool `yaml:"claude"`
	ChatGPT   bool `yaml:"chatgpt"`
	Copilot   bool `yaml:"copilot"`
	Cursor    bool `yaml:"cursor"`
	GenericAI bool `yaml:"generic_ai"`
}

type Targets struct {
	StagedDiff      bool `yaml:"staged_diff"`
	CommitMessages  bool `yaml:"commit_messages"`
	PRBody          bool `yaml:"pr_body"`
	FileHeaders     bool `yaml:"file_headers"`
}

type Behavior struct {
	DefaultAction          string `yaml:"default_action"`
	HookMode               string `yaml:"hook_mode"`
	PreserveHumanCoauthors bool   `yaml:"preserve_human_coauthors"`
}

func Default() Config {
	return Config{
		Providers: Providers{
			Claude: true, ChatGPT: true, Copilot: true, Cursor: true, GenericAI: true,
		},
		Targets: Targets{
			StagedDiff: true, CommitMessages: true, PRBody: true, FileHeaders: true,
		},
		Behavior: Behavior{
			DefaultAction: "review", HookMode: "warn", PreserveHumanCoauthors: true,
		},
		AllowedTrailers: []string{"Signed-off-by:", "Reviewed-by:"},
	}
}

func DefaultYAML() string {
	return `# Coauthor Cleaner configuration
providers:
  claude: true
  chatgpt: true
  copilot: true
  cursor: true
  generic_ai: true

targets:
  staged_diff: true
  commit_messages: true
  pr_body: true
  file_headers: true

behavior:
  default_action: review
  hook_mode: warn
  preserve_human_coauthors: true

allowed_trailers:
  - "Signed-off-by:"
  - "Reviewed-by:"
  # - "Co-authored-by: Your Name"
`
}

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func LoadFromRepo(repoRoot string) (Config, string, error) {
	path := filepath.Join(repoRoot, FileName)
	cfg, err := Load(path)
	return cfg, path, err
}

func IsAllowedTrailer(line string, allowed []string) bool {
	trimmed := strings.TrimSpace(line)
	for _, a := range allowed {
		if a == "" {
			continue
		}
		if strings.HasPrefix(trimmed, a) || strings.EqualFold(trimmed, strings.TrimSpace(a)) {
			return true
		}
	}
	return false
}

func (c Config) ProviderEnabled(ruleName string) bool {
	switch ruleName {
	case "ai-coauthor-trailer":
		return c.Providers.Claude || c.Providers.ChatGPT || c.Providers.Copilot || c.Providers.Cursor || c.Providers.GenericAI
	case "claude-generated-with":
		return c.Providers.Claude
	case "chatgpt-generated":
		return c.Providers.ChatGPT
	case "cursor-generated":
		return c.Providers.Cursor
	case "copilot-generated":
		return c.Providers.Copilot
	case "ai-emoji-generated":
		return c.Providers.GenericAI || c.Providers.ChatGPT
	case "generic-ai-assisted":
		return c.Providers.GenericAI
	default:
		return true
	}
}
