package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var builtInTemplates = map[string]Context{
	"default": {
		Project: struct {
			Name     string `yaml:"name"`
			Summary  string `yaml:"summary"`
			Template string `yaml:"template,omitempty"`
		}{
			Name:     "default",
			Summary:  "Project context placeholder. Capture architecture, standards, and risks here.",
			Template: "default",
		},
		Architecture: Architecture{
			Style:   "general",
			Version: "v1",
			Notes:   "Update this section with your system's architecture overview.",
		},
		Standards: map[string][]string{
			"process": {
				"Keep prompts token-cheap; prefer summaries over full dumps.",
				"Reference evidence paths; do not inline logs.",
			},
			"code": {
				"Maintain compatibility across supported platforms.",
			},
		},
		Constraints: []string{
			"Keep prompts token-cheap; expand context only when profile requests.",
			"Maintain portable state inside the repo for agent switching and parallel work.",
		},
		QualityGates: []string{
			"All tests pass and lint is clean.",
			"No breaking API changes.",
			"Prompt written to .agent/exports/current.prompt.md.",
		},
	},
	"react-spring": {
		Project: struct {
			Name     string `yaml:"name"`
			Summary  string `yaml:"summary"`
			Template string `yaml:"template,omitempty"`
		}{
			Name:     "react-spring",
			Summary:  "Full-stack React + Spring Boot application.",
			Template: "react-spring",
		},
		Architecture: Architecture{
			Style:   "layered",
			Version: "v1",
			Notes:   "React + TypeScript frontend talking to Spring Boot REST APIs; separate client/server modules with shared contracts.",
		},
		Standards: map[string][]string{
			"frontend": {
				"React + TypeScript with functional components and hooks.",
				"Feature-oriented folder structure with co-located tests and styles.",
				"Use lint/format defaults; keep API clients typed and surface errors to users.",
			},
			"backend": {
				"Spring Boot REST controllers -> services -> repositories with constructor injection.",
				"DTOs decoupled from persistence models; validate inputs at boundaries.",
				"JUnit/Mockito tests for services/controllers; consistent API error responses.",
			},
			"shared": {
				"Document API contracts and align client/server versions.",
			},
		},
		Constraints: []string{
			"Keep prompts token-cheap; expand context only when profile requests.",
			"Maintain portable state inside the repo for agent switching and parallel work.",
		},
		QualityGates: []string{
			"All tests pass and lint is clean.",
			"No breaking API changes.",
			"Prompt written to .agent/exports/current.prompt.md.",
		},
	},
}

// BuiltInTemplateNames returns built-in template names sorted.
func BuiltInTemplateNames() []string {
	names := make([]string, 0, len(builtInTemplates))
	for name := range builtInTemplates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ResolveTemplate loads a template by name, preferring repo templates, then built-ins, falling back to default.
func ResolveTemplate(templateName string) (Context, string, error) {
	if templateName == "" {
		templateName = "default"
	}
	if ctx, err := loadRepoTemplate(templateName); err == nil {
		return finalizeTemplateMetadata(ctx, templateName, templateName), templateName, nil
	} else if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, os.ErrNotExist) {
		return Context{}, "", err
	}
	if ctx, ok := builtInTemplate(templateName); ok {
		return finalizeTemplateMetadata(ctx, templateName, templateName), templateName, nil
	}
	fallback, ok := builtInTemplate("default")
	if !ok {
		return Context{}, "", fmt.Errorf("default template not available")
	}
	return finalizeTemplateMetadata(fallback, "default", templateName), "default", nil
}

// ListRepoTemplates returns YAML template files under .agent/templates.
func ListRepoTemplates() ([]string, error) {
	dir := AgentPath(templatesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
	}
	sort.Strings(names)
	return names, nil
}

// RepoTemplatePath resolves the path to a repo-local template.
func RepoTemplatePath(name string) string {
	filename := fmt.Sprintf("%s.yaml", name)
	return AgentPath(templatesDir, filename)
}

// InstallTemplate writes a built-in template into .agent/templates/<name>.yaml.
func InstallTemplate(name string, force bool) (string, error) {
	ctx, ok := builtInTemplate(name)
	if !ok {
		return "", fmt.Errorf("built-in template %q not found", name)
	}
	if err := os.MkdirAll(AgentPath(templatesDir), 0o755); err != nil {
		return "", err
	}
	dest := RepoTemplatePath(name)
	if !force {
		if _, err := os.Stat(dest); err == nil {
			return "", fmt.Errorf("template %q already exists at %s (use --force to overwrite)", name, dest)
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	ctx = finalizeTemplateMetadata(ctx, name, name)
	data, err := yaml.Marshal(ctx)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

func loadRepoTemplate(name string) (Context, error) {
	var ctx Context
	path := RepoTemplatePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return ctx, err
	}
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func builtInTemplate(name string) (Context, bool) {
	ctx, ok := builtInTemplates[name]
	if !ok {
		return Context{}, false
	}
	return cloneContext(ctx), true
}

func cloneContext(ctx Context) Context {
	out := ctx
	out.Constraints = append([]string(nil), ctx.Constraints...)
	out.QualityGates = append([]string(nil), ctx.QualityGates...)
	if ctx.Standards != nil {
		out.Standards = make(map[string][]string, len(ctx.Standards))
		for k, v := range ctx.Standards {
			out.Standards[k] = append([]string(nil), v...)
		}
	}
	return out
}

func finalizeTemplateMetadata(ctx Context, resolvedTemplateName, requestedName string) Context {
	if requestedName != "" && (ctx.Project.Name == "" || resolvedTemplateName != requestedName) {
		ctx.Project.Name = requestedName
	}
	if ctx.Project.Template == "" {
		ctx.Project.Template = resolvedTemplateName
	}
	return ctx
}
