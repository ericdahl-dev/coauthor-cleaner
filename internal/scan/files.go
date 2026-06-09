package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/config"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
}

func ScanPath(path string, cfg config.Config, strict, aggressive bool) ([]detect.Finding, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	rules := detect.SelectRules(cfg, strict, aggressive)
	opts := detect.ScanOpts{AllowedTrailers: cfg.AllowedTrailers}

	if !info.IsDir() {
		return scanFile(path, rules, opts)
	}
	return scanDir(path, rules, opts)
}

func scanFile(path string, rules []detect.Rule, opts detect.ScanOpts) ([]detect.Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return detect.ScanLines(string(data), detect.SourceFileHeader, path, rules, opts), nil
}

func scanDir(root string, rules []detect.Rule, opts detect.ScanOpts) ([]detect.Finding, error) {
	var findings []detect.Finding
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if isBinaryExt(path) {
			return nil
		}
		found, err := scanFile(path, rules, opts)
		if err != nil {
			return nil // skip unreadable files
		}
		findings = append(findings, found...)
		return nil
	})
	return findings, err
}

func isBinaryExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".zip", ".tar", ".gz", ".exe", ".dll", ".so", ".dylib", ".woff", ".woff2":
		return true
	}
	return false
}
