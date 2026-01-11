package agent

import "time"

// Architecture describes the project's structural style.
type Architecture struct {
	Style   string `yaml:"style"`
	Version string `yaml:"version,omitempty"`
	Notes   string `yaml:"notes,omitempty"`
}

// HealthSnapshot captures lightweight operational health.
type HealthSnapshot struct {
	Status string   `yaml:"status,omitempty"`
	Issues []string `yaml:"issues,omitempty"`
}

// Context represents slow-changing project context shared across work items.
type Context struct {
	Project struct {
		Name     string `yaml:"name"`
		Summary  string `yaml:"summary"`
		Template string `yaml:"template,omitempty"`
	} `yaml:"project"`
	Architecture Architecture       `yaml:"architecture"`
	Standards     map[string][]string `yaml:"standards,omitempty"`
	Constraints   []string            `yaml:"constraints,omitempty"`
	QualityGates  []string            `yaml:"quality_gates,omitempty"`
}

// State represents fast-changing state that is easy to resume.
type State struct {
	ActiveWorkItem   string `yaml:"active_work_item"`
	LastSummary      string `yaml:"last_summary,omitempty"`
	BranchSuggestion string `yaml:"branch_suggestion,omitempty"`
	Health           HealthSnapshot `yaml:"health,omitempty"`
}

// PromptProfile controls how much context is expanded when building a prompt.
type PromptProfile struct {
	Description         string `yaml:"description"`
	IncludeArchitecture bool   `yaml:"include_architecture"`
	IncludeStandards    bool   `yaml:"include_standards"`
	Detail              string `yaml:"detail,omitempty"`
}

// PromptProfileSet wraps configured profiles.
type PromptProfileSet struct {
	Profiles map[string]PromptProfile `yaml:"profiles"`
}

// WorkItem metadata is stored in front matter, while Body preserves user edits.
type WorkItem struct {
	ID                  string    `yaml:"id"`
	Title               string    `yaml:"title"`
	Intent              []string  `yaml:"intent,omitempty"`
	Status              string    `yaml:"status"`
	CreatedAt           time.Time `yaml:"created_at"`
	Evidence            []string  `yaml:"evidence,omitempty"`
	LastSummary         string    `yaml:"last_summary,omitempty"`
	AcceptanceCriteria  []string  `yaml:"acceptance_criteria,omitempty"`
	BranchSuggestion    string    `yaml:"branch_suggestion,omitempty"`
}

// WorkItemFile combines metadata with free-form body text.
type WorkItemFile struct {
	Meta WorkItem
	Body string
}
