package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultWhenMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Providers.Claude {
		t.Error("expected default providers")
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	content := `providers:
  claude: false
allowed_trailers:
  - "Co-authored-by: Eric Dahl"
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Providers.Claude {
		t.Error("claude should be disabled")
	}
	if len(cfg.AllowedTrailers) != 1 {
		t.Fatalf("trailers = %v", cfg.AllowedTrailers)
	}
}

func TestIsAllowedTrailer(t *testing.T) {
	allowed := []string{"Co-authored-by: Eric Dahl", "Signed-off-by:"}
	if !IsAllowedTrailer("Co-authored-by: Eric Dahl <eric@example.com>", allowed) {
		t.Error("expected allowed")
	}
	if IsAllowedTrailer("Co-authored-by: Claude <noreply@anthropic.com>", allowed) {
		t.Error("expected not allowed")
	}
}

func TestProviderEnabled(t *testing.T) {
	cfg := Default()
	cfg.Providers.Claude = false
	if cfg.ProviderEnabled("claude-generated-with") {
		t.Error("claude off should disable claude rule")
	}
	cfg.Providers.Claude = true
	if !cfg.ProviderEnabled("claude-generated-with") {
		t.Error("claude on should enable claude rule")
	}
}
