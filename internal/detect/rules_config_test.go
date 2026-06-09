package detect

import (
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/config"
)

func TestSelectRules_RespectsProviders(t *testing.T) {
	cfg := config.Default()
	cfg.Providers.Claude = false
	cfg.Providers.ChatGPT = false
	cfg.Providers.Copilot = false
	cfg.Providers.Cursor = false
	cfg.Providers.GenericAI = false
	rules := SelectRules(cfg, false, false)
	if len(rules) != 0 {
		t.Fatalf("got %d rules", len(rules))
	}
}

func TestListRules(t *testing.T) {
	if len(ListRules(config.Default())) < 5 {
		t.Fatal()
	}
}
