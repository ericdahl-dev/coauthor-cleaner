package detect

type SourceType string

const (
	SourceCommitMessage SourceType = "commit_message"
	SourceCommitTrailer SourceType = "commit_trailer"
	SourceStagedDiff    SourceType = "staged_diff"
	SourcePRDiff        SourceType = "pr_diff"
	SourcePRBody        SourceType = "pr_body"
	SourceFileHeader    SourceType = "file_header"
)

type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

type Finding struct {
	ID          string
	Source      SourceType
	FilePath    string
	LineNumber  int
	Match       string
	Replacement string
	Confidence  Confidence
	RuleName    string
	Selected    bool
}
