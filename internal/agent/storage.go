package agent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	agentDir           = ".agent"
	contextFile        = "context.yaml"
	stateFile          = "state.yaml"
	promptProfilesFile = "prompt_profiles.yaml"
	templatesDir       = "templates"
	workitemsDir       = "workitems"
	evidenceDir        = "evidence"
	exportsDir         = "exports"
	currentPromptFile  = "current.prompt.md"
)

var (
	workItemPattern = regexp.MustCompile(`^WI-(\d{3,})\.md$`)
)

// AgentPath builds a path relative to the .agent directory.
func AgentPath(parts ...string) string {
	all := append([]string{agentDir}, parts...)
	return filepath.Join(all...)
}

// EnsureAgentLayout creates the agent directory structure and default files.
func EnsureAgentLayout(templateName string) error {
	if err := ensureFreshAgentLayout(); err != nil {
		return err
	}

	ctx, _, err := ResolveTemplate(templateName)
	if err != nil {
		return err
	}

	dirs := []string{
		AgentPath(),
		AgentPath(workitemsDir),
		AgentPath(evidenceDir),
		AgentPath(exportsDir),
		AgentPath(templatesDir),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	if err := SaveContext(ctx); err != nil {
		return err
	}
	if err := SaveState(DefaultState()); err != nil {
		return err
	}
	if err := SavePromptProfiles(DefaultPromptProfiles()); err != nil {
		return err
	}
	return nil
}

// ensureFreshAgentLayout avoids overwriting existing agent data.
func ensureFreshAgentLayout() error {
	info, err := os.Stat(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", agentDir)
	}
	protected := []string{
		AgentPath(contextFile),
		AgentPath(stateFile),
		AgentPath(promptProfilesFile),
	}
	for _, path := range protected {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf(".agent already exists, refusing to overwrite")
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// DefaultState constructs starter state.
func DefaultState() State {
	st := State{}
	st.Health.Status = "unknown"
	st.Health.Issues = []string{}
	return st
}

// DefaultPromptProfiles provides the required profiles.
func DefaultPromptProfiles() PromptProfileSet {
	return PromptProfileSet{
		Profiles: map[string]PromptProfile{
			"cheap": {
				Description:         "Summaries only; minimal context and references.",
				IncludeArchitecture: false,
				IncludeStandards:    false,
				Detail:              "summary",
			},
			"standard": {
				Description:         "Include architecture and standards for balanced prompts.",
				IncludeArchitecture: true,
				IncludeStandards:    true,
				Detail:              "balanced",
			},
			"deep": {
				Description:         "Full context disclosure; include architecture, standards, and constraints.",
				IncludeArchitecture: true,
				IncludeStandards:    true,
				Detail:              "full",
			},
		},
	}
}

// SaveContext writes context.yaml.
func SaveContext(ctx Context) error {
	return saveYAML(AgentPath(contextFile), ctx)
}

// LoadContext reads context.yaml.
func LoadContext() (Context, error) {
	var ctx Context
	if err := readYAML(AgentPath(contextFile), &ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// SaveState writes state.yaml.
func SaveState(st State) error {
	if st.Health.Status == "" && len(st.Health.Issues) == 0 {
		st.Health.Status = "unknown"
	}
	return saveYAML(AgentPath(stateFile), st)
}

// LoadState reads state.yaml.
func LoadState() (State, error) {
	var st State
	if err := readYAML(AgentPath(stateFile), &st); err != nil {
		return st, err
	}
	if st.Health.Status == "" && len(st.Health.Issues) == 0 {
		st.Health.Status = "unknown"
	}
	return st, nil
}

// SavePromptProfiles writes prompt_profiles.yaml.
func SavePromptProfiles(p PromptProfileSet) error {
	return saveYAML(AgentPath(promptProfilesFile), p)
}

// LoadPromptProfiles reads prompt_profiles.yaml.
func LoadPromptProfiles() (PromptProfileSet, error) {
	var p PromptProfileSet
	if err := readYAML(AgentPath(promptProfilesFile), &p); err != nil {
		return p, err
	}
	return p, nil
}

func saveYAML(path string, data any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func readYAML(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// NextWorkItemID computes the next sequential work item ID.
func NextWorkItemID() (string, error) {
	dir := AgentPath(workitemsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := workItemPattern.FindStringSubmatch(e.Name())
		if len(m) == 2 {
			var num int
			fmt.Sscanf(m[1], "%d", &num)
			if num > max {
				max = num
			}
		}
	}
	return fmt.Sprintf("WI-%03d", max+1), nil
}

// LoadWorkItem reads a work item file.
func LoadWorkItem(id string) (*WorkItemFile, error) {
	path := WorkItemPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseWorkItem(data)
}

// SaveWorkItem writes a work item back to disk.
func SaveWorkItem(wi *WorkItemFile) error {
	metaBytes, err := yaml.Marshal(wi.Meta)
	if err != nil {
		return err
	}

	var body string
	if strings.TrimSpace(wi.Body) == "" {
		body = defaultWorkItemBody(wi.Meta)
	} else {
		body = wi.Body
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(metaBytes)
	buf.WriteString("---\n")
	if !strings.HasPrefix(body, "\n") {
		buf.WriteString("\n")
	}
	buf.WriteString(body)

	return os.WriteFile(WorkItemPath(wi.Meta.ID), buf.Bytes(), 0o644)
}

// WorkItemPath returns the path to a work item file.
func WorkItemPath(id string) string {
	filename := fmt.Sprintf("%s.md", id)
	return AgentPath(workitemsDir, filename)
}

func parseWorkItem(data []byte) (*WorkItemFile, error) {
	str := string(data)
	if !strings.HasPrefix(str, "---") {
		return nil, errors.New("work item missing front matter")
	}
	parts := strings.SplitN(str, "---", 3)
	if len(parts) < 3 {
		return nil, errors.New("invalid work item format")
	}
	front := strings.TrimSpace(parts[1])
	body := strings.TrimLeft(parts[2], "\n")

	var meta WorkItem
	if err := yaml.Unmarshal([]byte(front), &meta); err != nil {
		return nil, err
	}

	return &WorkItemFile{
		Meta: meta,
		Body: body,
	}, nil
}

func defaultWorkItemBody(meta WorkItem) string {
	return fmt.Sprintf(`# Work Item %s

## Summary
%s

## Acceptance Criteria
- Add criteria as you work.

## Notes
- Capture decisions, scope, and dependencies here.
`, meta.ID, meta.Title)
}

// CopyEvidence copies a file into the evidence directory and returns the relative path used.
func CopyEvidence(srcPath string) (string, error) {
	if err := os.MkdirAll(AgentPath(evidenceDir), 0o755); err != nil {
		return "", err
	}
	base := filepath.Base(srcPath)
	dest := AgentPath(evidenceDir, base)

	if _, err := os.Stat(dest); err == nil {
		dest = uniquePath(dest)
	}

	if err := copyFile(srcPath, dest); err != nil {
		return "", err
	}
	rel, err := filepath.Rel(agentDir, dest)
	if err != nil {
		return "", err
	}
	return rel, nil
}

func uniquePath(path string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// EnsureAgentExists verifies that .agent is present.
func EnsureAgentExists() error {
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		return fmt.Errorf(".agent not found. Run ctx init <template> first")
	}
	return nil
}

// UpdateWorkItemStatus updates the status field for a work item.
func UpdateWorkItemStatus(id, status string) error {
	wi, err := LoadWorkItem(id)
	if err != nil {
		return err
	}
	wi.Meta.Status = status
	return SaveWorkItem(wi)
}

// ListWorkItems returns available work item IDs sorted ascending.
func ListWorkItems() ([]string, error) {
	entries, err := os.ReadDir(AgentPath(workitemsDir))
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if workItemPattern.MatchString(name) {
			ids = append(ids, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// TouchPromptFile ensures the exports directory exists.
func TouchPromptFile() (string, error) {
	if err := os.MkdirAll(AgentPath(exportsDir), 0o755); err != nil {
		return "", err
	}
	return AgentPath(exportsDir, currentPromptFile), nil
}

// NewWorkItemFile constructs a new work item with defaults.
func NewWorkItemFile(id, title string, intents []string) *WorkItemFile {
	if len(intents) == 0 {
		intents = []string{"general"}
	}
	return &WorkItemFile{
		Meta: WorkItem{
			ID:        id,
			Title:     title,
			Intent:    intents,
			Status:    "active",
			CreatedAt: time.Now().UTC(),
		},
	}
}
