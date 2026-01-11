package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// SuggestBranchName proposes a branch name for a work item.
func SuggestBranchName(w WorkItem) string {
	base := strings.ToLower(w.Title)
	base = sanitizeBranch(base)
	return fmt.Sprintf("%s-%s", strings.ToLower(w.ID), base)
}

func sanitizeBranch(in string) string {
	in = strings.ReplaceAll(in, "&", "and")
	re := regexp.MustCompile(`[^a-z0-9]+`)
	in = re.ReplaceAllString(in, "-")
	in = strings.Trim(in, "-")
	if in == "" {
		in = "work"
	}
	if len(in) > 32 {
		in = in[:32]
	}
	return in
}
