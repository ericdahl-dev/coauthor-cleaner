package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
)

type PR struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Client struct {
	Dir string
}

func (c Client) run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

func (c Client) Available() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func (c Client) ViewPR() (PR, error) {
	out, err := c.run("pr", "view", "--json", "title,body")
	if err != nil {
		return PR{}, err
	}
	var pr PR
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return PR{}, err
	}
	return pr, nil
}

func (c Client) EditBody(body string) error {
	_, err := c.run("pr", "edit", "--body", body)
	return err
}

func (c Client) EditTitle(title string) error {
	_, err := c.run("pr", "edit", "--title", title)
	return err
}

func ScanPR(cfg config.Config, pr PR, strict, aggressive bool) []detect.Finding {
	rules := detect.SelectRules(cfg, strict, aggressive)
	opts := detect.ScanOpts{AllowedTrailers: cfg.AllowedTrailers}
	var findings []detect.Finding
	if pr.Title != "" {
		findings = append(findings, detect.ScanLines(pr.Title+"\n", detect.SourcePRBody, "pr_title", rules, opts)...)
	}
	if pr.Body != "" {
		findings = append(findings, detect.ScanLines(pr.Body, detect.SourcePRBody, "pull_request", rules, opts)...)
	}
	return findings
}

func CleanPR(pr PR, findings []detect.Finding) (string, string) {
	title := pr.Title
	body := pr.Body
	if len(findings) == 0 {
		return title, body
	}
	var titleFindings, bodyFindings []detect.Finding
	for _, f := range findings {
		if f.FilePath == "pr_title" {
			titleFindings = append(titleFindings, f)
		} else {
			bodyFindings = append(bodyFindings, f)
		}
	}
	if len(titleFindings) > 0 {
		title = strings.TrimSpace(detect.CleanText(pr.Title+"\n", titleFindings))
	}
	if len(bodyFindings) > 0 {
		body = detect.CleanText(pr.Body, bodyFindings)
	}
	return title, body
}
