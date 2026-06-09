package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (r Runner) Push() error {
	return r.push(false)
}

func (r Runner) PushForceWithLease() error {
	return r.push(true)
}

func (r Runner) push(force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force-with-lease")
	}
	cmd := exec.Command("git", args...)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return nil
}
