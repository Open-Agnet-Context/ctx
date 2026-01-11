# Devin Operating Procedure (repo-first, .agent contract)

## 0) Goal
Work with minimal user prompting. Use repo-local context/state so work is resumable, portable across agents, and consistent with architecture/standards.

## 1) Always read first (in this order)
1. `.agent/context.yaml`  (architecture, standards, quality_gates, constraints)
2. `.agent/state.yaml`    (active work item + handoff)
3. `.agent/workitems/`    (open the active WI file referenced by state)

## 2) What to work on
- If `state.active_work_item` is set, work on that WI.
- If it’s empty, create a new work item (or ask for a 1-line goal) and set it active by updating `.agent/state.yaml`.

## 3) How to execute work (defaults)
- Make incremental, reviewable commits.
- Follow architecture/standards from `context.yaml`.
- Respect `quality_gates` and `constraints`.
- Keep scope tight to the active WI acceptance criteria.
- Do not introduce network calls/telemetry into `ctx` itself unless explicitly requested.

## 4) Updating repo state (required)
After **any** meaningful progress (pause/stop/finish), update:
- The active work item file in `.agent/workitems/WI-XXX.md`:
  - Status: `In progress` / `Blocked` / `Done`
  - Notes: 3–6 bullets (what changed, where, why)
  - Next steps: 1–3 bullets
- `.agent/state.yaml`:
  - `active_work_item` (keep accurate)
  - `last_summary` (one short paragraph)
  - `branch_suggestion` (if relevant)
  - `health.status` + `health.issues` (short, no logs)

## 5) Handling architecture changes (strict rule)
Do **not** silently change architecture.
If you believe architecture must change:
1. Propose the change briefly (why + impact).
2. Create/use a dedicated ARCH work item (e.g., `WI-ARCH-00X`) and mark it active.
3. Only after approval, update `.agent/context.yaml`:
   - bump `architecture.version`
   - update `architecture.style` and `notes`
   - add any new standards/constraints

## 6) Working on existing projects (no .agent yet)
If `.agent/` is missing:
- Create `.agent/` and minimal files:
  - `context.yaml` with a reasonable first-pass architecture + standards
  - `state.yaml` with empty/unknown fields
  - a first WI capturing the user’s goal
- Then proceed normally and refine context as you discover repo conventions.

## 7) Minimal user interaction pattern
If the user says “fix this” with little detail:
- Infer scope from the repo and active WI.
- Ask at most ONE clarifying question only if truly blocking.
- Otherwise proceed with a safe, incremental fix.

## 8) Evidence
Evidence attachments are optional. If provided, store as files under `.agent/evidence/` and reference paths in the WI. Do not inline large logs.

Follow AGENTS.md. Read .agent/context.yaml and .agent/state.yaml. Fix the active work item and update WI + state on completion.