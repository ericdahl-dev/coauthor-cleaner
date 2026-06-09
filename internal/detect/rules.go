package detect

import "regexp"

type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
	SourceTypes []SourceType
	Confidence  Confidence
	Trailer     bool // matches commit trailer lines
}

func baseRules() []Rule {
	return []Rule{
		{
			Name:        "claude-generated-with",
			Pattern:     regexp.MustCompile(`(?i)generated with claude( code)?`),
			Replacement: "",
			SourceTypes: []SourceType{SourceStagedDiff, SourcePRDiff, SourceFileHeader, SourcePRBody},
			Confidence:  ConfidenceHigh,
		},
		{
			Name:        "chatgpt-generated",
			Pattern:     regexp.MustCompile(`(?i)generated (by|with) chatgpt`),
			Replacement: "",
			SourceTypes: []SourceType{SourceStagedDiff, SourcePRDiff, SourceFileHeader, SourcePRBody, SourceCommitMessage},
			Confidence:  ConfidenceHigh,
		},
		{
			Name:        "cursor-generated",
			Pattern:     regexp.MustCompile(`(?i)generated with cursor`),
			Replacement: "",
			SourceTypes: []SourceType{SourceStagedDiff, SourcePRDiff, SourceFileHeader, SourcePRBody},
			Confidence:  ConfidenceHigh,
		},
		{
			Name:        "copilot-generated",
			Pattern:     regexp.MustCompile(`(?i)generated (by|with) (github )?copilot`),
			Replacement: "",
			SourceTypes: []SourceType{SourceStagedDiff, SourcePRDiff, SourceFileHeader, SourcePRBody},
			Confidence:  ConfidenceHigh,
		},
		{
			Name:        "ai-coauthor-trailer",
			Pattern:     regexp.MustCompile(`(?i)^co-authored-by:\s*(chatgpt|claude|copilot|cursor|ai assistant|github copilot).*$`),
			Replacement: "",
			SourceTypes: []SourceType{SourceCommitMessage},
			Confidence:  ConfidenceHigh,
			Trailer:     true,
		},
		{
			Name:        "ai-emoji-generated",
			Pattern:     regexp.MustCompile(`(?i)🤖\s*generated with`),
			Replacement: "",
			SourceTypes: []SourceType{SourcePRBody, SourceStagedDiff, SourcePRDiff},
			Confidence:  ConfidenceHigh,
		},
	}
}

func aggressiveRules() []Rule {
	rules := baseRules()
	rules = append(rules, Rule{
		Name:        "generic-ai-assisted",
		Pattern:     regexp.MustCompile(`(?i)(assisted|written|created)\s+by\s+(an?\s+)?ai`),
		Replacement: "",
		SourceTypes: []SourceType{SourceStagedDiff, SourcePRDiff, SourceFileHeader, SourcePRBody},
		Confidence:  ConfidenceLow,
	})
	return rules
}

func DefaultRules() []Rule  { return baseRules() }
func StrictRules() []Rule    { return baseRules() }
func AggressiveRules() []Rule { return aggressiveRules() }

func ruleApplies(rule Rule, source SourceType) bool {
	for _, s := range rule.SourceTypes {
		if s == source {
			return true
		}
	}
	return false
}
