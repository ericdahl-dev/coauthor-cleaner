package github

import (
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/config"
)

func TestScanPR_Body(t *testing.T) {
	cfg := config.Default()
	pr := PR{Body: "Summary\n\n🤖 Generated with ChatGPT\n"}
	findings := ScanPR(cfg, pr, false, false)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestCleanPR(t *testing.T) {
	pr := PR{Body: "Summary\n\n🤖 Generated with ChatGPT\n\nDetails here.\n"}
	findings := ScanPR(config.Default(), pr, false, false)
	_, body := CleanPR(pr, findings)
	if body == pr.Body {
		t.Error("expected cleaned body")
	}
}
