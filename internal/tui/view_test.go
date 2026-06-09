package tui

import (
	"strings"
	"testing"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
)

func TestPreviewDiff(t *testing.T) {
	out := PreviewDiff(detect.Finding{Match: "Co-authored-by: Claude", RuleName: "ai-coauthor-trailer", Confidence: detect.ConfidenceHigh})
	if !strings.Contains(out, "- Co-authored-by") || !strings.Contains(out, "ai-coauthor-trailer") {
		t.Fatal(out)
	}
}

func TestContextKeys(t *testing.T) {
	m := Model{screen: screenMain}
	if !strings.Contains(m.contextKeys(), "clean") {
		t.Fatal(m.contextKeys())
	}
	m.screen = screenPush
	if !strings.Contains(m.contextKeys(), "force-with-lease") {
		t.Fatal(m.contextKeys())
	}
}

func TestRenderStatusBar(t *testing.T) {
	m := Model{
		layout: computeLayout(80, 24),
		repoState: git.RepoState{Branch: "main", Upstream: "origin/main", UpstreamExists: true},
		findings: []detect.Finding{{}},
	}
	out := m.renderStatusBar()
	if !strings.Contains(out, "main") || !strings.Contains(out, "findings") {
		t.Fatal(out)
	}
}

func TestComputeLayout(t *testing.T) {
	ly := computeLayout(100, 30)
	if ly.leftWidth < 28 || ly.rightWidth < 24 {
		t.Fatalf("%+v", ly)
	}
}
