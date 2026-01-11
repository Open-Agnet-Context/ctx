package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// PromptData collects data for template execution.
type PromptData struct {
	Profile        string
	WorkItem       WorkItem
	State          State
	Context        Context
	Constraints    []string
	LikelyFiles    []string
	Evidence       []string
	QualityGates   []string
	TaskAcceptance []string
	HealthStatus   string
	HealthIssues   []string
}

var promptTemplate = `Task: {{.WorkItem.Title}} ({{.WorkItem.ID}})
Intent: {{join .WorkItem.Intent ", "}}
Status: {{.WorkItem.Status}}
Health: {{healthLine .HealthStatus}}
Last Summary: {{summaryLine .State.LastSummary .WorkItem.LastSummary}}

Constraints:
{{bulletList .Constraints}}

Quality Gates:
{{bulletList .QualityGates}}

Evidence (paths only):
{{bulletList .Evidence}}

Likely Files:
{{bulletList .LikelyFiles}}

Task Acceptance:
{{bulletList .TaskAcceptance}}

{{if .Context.Project.Summary}}Project Context:
- {{.Context.Project.Summary}}
{{end}}{{if includeArch .Profile}}
Architecture:
- {{archSummary .Context.Architecture}}{{end}}{{if includeStandards .Profile}}
Standards:
{{scopedList .Context.Standards}}{{end}}{{if healthIssuesPresent .HealthIssues}}

Health Issues:
{{bulletList .HealthIssues}}{{end}}
`

// BuildPrompt assembles the prompt and writes exports/current.prompt.md.
func BuildPrompt(profileName string) (string, error) {
	if profileName == "" {
		profileName = "cheap"
	}
	profiles, err := LoadPromptProfiles()
	if err != nil {
		return "", err
	}
	_, ok := profiles.Profiles[profileName]
	if !ok {
		return "", fmt.Errorf("prompt profile %q not found", profileName)
	}

	state, err := LoadState()
	if err != nil {
		return "", err
	}
	if state.ActiveWorkItem == "" {
		return "", fmt.Errorf("no active work item; start one with ctx work start <WI-XXX>")
	}

	wiFile, err := LoadWorkItem(state.ActiveWorkItem)
	if err != nil {
		return "", err
	}
	context, err := LoadContext()
	if err != nil {
		return "", err
	}

	constraints := mergeUnique(context.Constraints, []string{
		"No network access; offline-only CLI.",
		"Do not embed logs; reference evidence paths.",
		"Keep prompts token-cheap; expand only by profile.",
	})
	taskAcceptance := wiFile.Meta.AcceptanceCriteria
	if len(taskAcceptance) == 0 {
		taskAcceptance = []string{"Work item completes without expanding scope."}
	}
	qualityGates := context.QualityGates
	if len(qualityGates) == 0 {
		qualityGates = []string{"All tests pass.", "No breaking API changes."}
	}

	data := PromptData{
		Profile:        profileName,
		WorkItem:       wiFile.Meta,
		State:          state,
		Context:        context,
		Constraints:    constraints,
		LikelyFiles:    likelyFiles(wiFile.Meta),
		Evidence:       evidenceList(wiFile.Meta),
		QualityGates:   qualityGates,
		TaskAcceptance: taskAcceptance,
		HealthStatus:   state.Health.Status,
		HealthIssues:   state.Health.Issues,
	}

	tpl := template.Must(template.New("prompt").Funcs(template.FuncMap{
		"join": func(items []string, sep string) string {
			return strings.Join(items, sep)
		},
		"bulletList": bulletList,
		"includeArch": func(profile string) bool {
			p := profiles.Profiles[profile]
			return p.IncludeArchitecture
		},
		"includeStandards": func(profile string) bool {
			p := profiles.Profiles[profile]
			return p.IncludeStandards
		},
		"summaryLine": func(parts ...string) string {
			for _, p := range parts {
				if strings.TrimSpace(p) != "" {
					return p
				}
			}
			return "Not provided."
		},
		"archSummary": archSummary,
		"scopedList":  scopedList,
		"healthLine": func(status string) string {
			if strings.TrimSpace(status) == "" {
				return "unknown"
			}
			return status
		},
		"healthIssuesPresent": func(issues []string) bool {
			return len(issues) > 0
		},
	}).Parse(promptTemplate))

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	dest, err := TouchPromptFile()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

func bulletList(items []string) string {
	if len(items) == 0 {
		return "- None"
	}
	var b strings.Builder
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return "- None"
	}
	return strings.TrimRight(b.String(), "\n")
}

func mergeUnique(primary []string, extras []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range [][]string{primary, extras} {
		for _, item := range list {
			item = strings.TrimSpace(item)
			if item == "" || seen[item] {
				continue
			}
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func evidenceList(w WorkItem) []string {
	var items []string
	for _, e := range w.Evidence {
		if strings.TrimSpace(e) != "" {
			items = append(items, filepath.ToSlash(e))
		}
	}
	return items
}

func likelyFiles(w WorkItem) []string {
	text := strings.ToLower(w.Title + " " + strings.Join(w.Intent, " "))
	var files []string

	if containsAny(text, []string{"ui", "react", "component", "css", "html", "frontend"}) {
		files = append(files, "src/ui/", "web/", "frontend/", "components/")
	}
	if containsAny(text, []string{"api", "server", "backend", "service", "timeout", "latency"}) {
		files = append(files, "cmd/", "internal/", "api/", "server/")
	}
	if containsAny(text, []string{"test", "bug", "fix", "regression"}) {
		files = append(files, "tests/", "internal/", "cmd/")
	}
	if len(files) == 0 {
		files = append(files, "cmd/", "internal/", "pkg/")
	}
	return dedupe(files)
}

func containsAny(text string, terms []string) bool {
	for _, t := range terms {
		if strings.Contains(text, t) {
			return true
		}
	}
	return false
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func archSummary(a Architecture) string {
	var parts []string
	if strings.TrimSpace(a.Style) != "" {
		parts = append(parts, a.Style)
	}
	if strings.TrimSpace(a.Version) != "" {
		parts = append(parts, a.Version)
	}
	summary := strings.Join(parts, " ")
	if strings.TrimSpace(a.Notes) != "" {
		if summary != "" {
			summary += " — " + a.Notes
		} else {
			summary = a.Notes
		}
	}
	if strings.TrimSpace(summary) == "" {
		return "Not documented."
	}
	return summary
}

func scopedList(scopes map[string][]string) string {
	if len(scopes) == 0 {
		return "- None"
	}
	keys := make([]string, 0, len(scopes))
	for k := range scopes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, scope := range keys {
		items := scopes[scope]
		b.WriteString("- ")
		b.WriteString(scope)
		b.WriteString(": ")
		if len(items) == 0 {
			b.WriteString("None")
		} else {
			b.WriteString(strings.Join(items, "; "))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
