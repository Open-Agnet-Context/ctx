package agent

import (
	"regexp"
	"strings"
)

var intentRules = map[string][]string{
	"bugfix":    {"fix", "error", "broken", "failure", "bug", "regression", "crash"},
	"frontend":  {"ui", "react", "component", "console", "browser", "css", "html"},
	"backend":   {"api", "timeout", "service", "database", "db", "server", "latency", "test"},
	"design":    {"architecture", "refactor", "design", "pattern", "structure"},
}

// ClassifyIntent applies rule-based, deterministic intent tags.
func ClassifyIntent(text string) []string {
	lower := strings.ToLower(text)
	var intents []string
	for intent, keywords := range intentRules {
		for _, kw := range keywords {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(kw) + `\b`)
			if re.MatchString(lower) {
				intents = append(intents, intent)
				break
			}
		}
	}
	if len(intents) == 0 {
		intents = append(intents, "general")
	}
	return intents
}
