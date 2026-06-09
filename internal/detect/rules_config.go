package detect

import "github.com/Skeyelab/coauthor-cleaner/internal/config"

func SelectRules(cfg config.Config, strict, aggressive bool) []Rule {
	var rules []Rule
	if aggressive {
		rules = AggressiveRules()
	} else if strict {
		rules = StrictRules()
	} else {
		rules = DefaultRules()
	}

	var filtered []Rule
	for _, r := range rules {
		if cfg.ProviderEnabled(r.Name) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func ListRules(cfg config.Config) []Rule {
	return SelectRules(cfg, false, false)
}
