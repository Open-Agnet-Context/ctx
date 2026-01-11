# ctx

Offline, repo-local CLI for managing developer context, intent, and prompts for coding agents.

## Features
- Creates and maintains `.agent/` with YAML/Markdown artifacts only.
- Rule-based intent classification to keep prompting cheap and deterministic.
- Work item lifecycle: issue creation, active switching, handoff summaries.
- Evidence ingestion without embedding logs (paths only).
- Profile-driven prompt assembly written to `.agent/exports/current.prompt.md` with global quality gates and task-level acceptance.

## Requirements
- Go 1.21+ to build the static binary.
- No runtime network access or external services are required.

## Install
```bash
go build -trimpath -ldflags="-s -w" -o ctx .
```
For fully offline environments, vendor dependencies before distribution:
```bash
go mod tidy
go mod vendor
```

## Usage
- `ctx init <template>`: create `.agent/` with starter context/state/prompt profiles. No prompts.
- `ctx template list`: show built-in templates and repo-local overrides.
- `ctx template install <name> [--force]`: copy a built-in template into `.agent/templates/`.
- `ctx context apply <template>`: overwrite `.agent/context.yaml` with a template (after init).
- `ctx issue "<text>"`: create a new work item, classify intent, set it active.
- `ctx work start <WI-XXX>`: mark a work item active and suggest a branch name.
- `ctx work stop`: prompt for a one-line handoff summary and pause the active item.
- `ctx evidence add <file>`: copy evidence into `.agent/evidence/` and link it to the active item.
- `ctx prompt --profile <cheap|standard|deep>`: generate the prompt at `.agent/exports/current.prompt.md`.

## Templates
- Repo templates live in `.agent/templates/<name>.yaml` and follow the same structure as `.agent/context.yaml`.
- `ctx init <template>` resolves templates in this order: repo template override, built-in template, built-in `default` fallback.
- `ctx context apply <template>` overwrites `.agent/context.yaml` using the same resolution order so you can switch templates after init without touching state or prompt profiles.
- `ctx template list` shows built-in templates and any repo templates.
- `ctx template install <name> [--force]` copies a built-in template into `.agent/templates/` so you can edit it without rebuilding the binary.

### Smoke Test
1. In a new folder: `ctx init default`.
2. `ctx template list` shows built-ins plus repo templates (if present).
3. `ctx template install react-spring` creates `.agent/templates/react-spring.yaml`.
4. Edit the installed template (for example, adjust architecture notes) to confirm overrides.
5. `ctx init react-spring` writes `.agent/context.yaml` using the repo template values.

## Repository Contract
```
.agent/
  context.yaml
  state.yaml
  prompt_profiles.yaml
  templates/
    <template>.yaml
  workitems/
    WI-001.md
  evidence/
    sample.log
  exports/
    current.prompt.md
```

## Security & Posture
- No network calls, telemetry, or background services.
- No agent SDKs or custom DSLs; YAML + Markdown only.
- Everything stored inside the repo for portability and auditability.
- Evidence is referenced by path only; contents are never embedded in prompts.
