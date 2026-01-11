Task: Add command ctx context apply <template> to overwrite only .agent/context.yaml (allow switching templates after init); keep ctx init bootstrap-only (WI-001)
Intent: general
Status: active
Health: unknown
Last Summary: Not provided.

Constraints:
- Keep prompts token-cheap; expand context only when profile requests.
- Maintain portable state inside the repo for agent switching and parallel work.
- Preserve backward compatibility of .agent/ artifacts and CLI flags.
- No network access; offline-only CLI.
- Do not embed logs; reference evidence paths.
- Keep prompts token-cheap; expand only by profile.

Quality Gates:
- All tests pass and lint is clean.
- No breaking API changes.
- Prompt written to .agent/exports/current.prompt.md.

Evidence (paths only):
- None

Likely Files:
- cmd/
- internal/
- pkg/

Task Acceptance:
- Work item completes without expanding scope.

Project Context:
- Offline, repo-local CLI for managing developer context, intent, and prompts without network access.

Architecture:
- cli-offline v1 — Single static Go binary (Cobra) that reads/writes .agent/ data in-repo; deterministic, no telemetry or network calls.
Standards:
- code: Go 1.21+; cobra commands kept small and composable.; No runtime network calls or telemetry; all I/O is repo-local.; Keep persisted YAML/Markdown human-editable and backward compatible.
- process: Keep prompts token-cheap; prefer summaries over full dumps.; Reference evidence paths; do not inline logs.; Default to offline operation; avoid introducing external services.
