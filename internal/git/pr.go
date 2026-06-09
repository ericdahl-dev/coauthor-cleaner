package git

import (
	"strings"
)

type AddedLine struct {
	FilePath string
	Line     string
}

// DiffAddedLines returns lines added in base...head (three-dot diff).
func (r Runner) DiffAddedLines(base, head string) ([]AddedLine, error) {
	diff, err := r.run("diff", base+"..."+head)
	if err != nil {
		return nil, err
	}
	return parseAddedLines(diff), nil
}

func parseAddedLines(diff string) []AddedLine {
	var lines []AddedLine
	currentFile := ""

	for _, raw := range strings.Split(diff, "\n") {
		if strings.HasPrefix(raw, "+++ b/") {
			currentFile = strings.TrimPrefix(raw, "+++ b/")
			continue
		}
		if !strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "+++") {
			continue
		}
		if currentFile == "" {
			continue
		}
		line := strings.TrimPrefix(raw, "+")
		if line == "" {
			continue
		}
		lines = append(lines, AddedLine{
			FilePath: currentFile,
			Line:     line,
		})
	}

	return lines
}

// CommitMessagesInRange returns commit SHA -> message for commits reachable from head but not base.
func (r Runner) CommitMessagesInRange(base, head string) (map[string]string, error) {
	shas, err := r.run("rev-list", base+".."+head)
	if err != nil {
		return nil, err
	}
	msgs := make(map[string]string)
	for _, sha := range strings.Split(strings.TrimSpace(shas), "\n") {
		if sha == "" {
			continue
		}
		msg, err := r.run("log", "-1", "--format=%B", sha)
		if err != nil {
			return nil, err
		}
		msgs[sha] = msg
	}
	return msgs, nil
}
