package scan

import (
	"encoding/json"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
)

type JSONFinding struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	FilePath   string `json:"file_path,omitempty"`
	LineNumber int    `json:"line_number,omitempty"`
	Match      string `json:"match"`
	RuleName   string `json:"rule_name"`
	Confidence string `json:"confidence"`
}

type JSONReport struct {
	Count    int           `json:"count"`
	Findings []JSONFinding `json:"findings"`
}

func FormatJSON(findings []detect.Finding) ([]byte, error) {
	report := JSONReport{
		Count:    len(findings),
		Findings: make([]JSONFinding, 0, len(findings)),
	}
	for _, f := range findings {
		report.Findings = append(report.Findings, JSONFinding{
			ID:         f.ID,
			Source:     string(f.Source),
			FilePath:   f.FilePath,
			LineNumber: f.LineNumber,
			Match:      f.Match,
			RuleName:   f.RuleName,
			Confidence: string(f.Confidence),
		})
	}
	return json.MarshalIndent(report, "", "  ")
}
