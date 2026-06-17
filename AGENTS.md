# AGENTS.md

## Purpose
This file is the Codex-side working agreement for `WoL-tool-Claude`.

Codex uses this file to preserve design intent, decide whether work should stay in Codex or be handed off to Claude Code, and review implementation results.
Claude Code uses `CLAUDE.md` for execution rules.

## Project Summary
- Project name: `WoL-tool-Claude`
- Purpose: Wake-on-LAN, ping monitoring, and remote shutdown web tool for LAN PCs.
- Summary from project docs: Wake-on-LAN web tool for Raspberry Pi Docker deployment.
- Runtime target: Raspberry Pi Docker linux/arm64, Go single binary
- Repository path: `D:\Git\WoL-tool-Claude`
- Stack: Go, plain JavaScript, Docker, host networking

## Base References
- Codex base: `D:/Git/CLAUDEmdStrage/_base/AGENTS.md`
- Claude Code base for Windows/local projects: `D:/Git/CLAUDEmdStrage/_base/CLAUDE_windows.md`
- Claude Code base for Raspberry Pi Docker projects: `D:/Git/CLAUDEmdStrage/_base/CLAUDE_docker.md`

## Role Split
Codex is responsible for:
- clarifying requirements, non-goals, and success criteria
- identifying change type and design risk
- preserving responsibility boundaries and design intent
- preparing scoped Claude Code handoffs when execution is clear
- reviewing Claude Code output against this file and the handoff
- recording durable decisions in `AGENTS.md` or `docs/*.md`

Claude Code is responsible for:
- following the current Codex handoff and `CLAUDE.md`
- editing only allowed files unless it explains why more files are required
- running requested verification where possible
- reporting changed files, summary, verification results, blocked checks, and design questions

## Claude Code Model Orchestration
Claude Code should normally run with Opus as the primary coordinator.

Opus is responsible for:
- reading `AGENTS.md`, `CLAUDE.md`, handoff files, and relevant project context
- interpreting requirements, constraints, non-goals, and verification expectations
- deciding the implementation plan and whether subagents are appropriate
- giving Sonnet subagents narrow implementation or investigation tasks
- reviewing subagent output before final reporting
- making design-sensitive decisions only when they are already allowed by the handoff

Sonnet subagents are responsible for:
- mechanical code edits
- repetitive refactors inside an explicit file scope
- localized tests, verification, and log/code inspection
- implementation tasks where goal, files, constraints, and non-goals are already clear

Sonnet subagents must not:
- change documented design intent on their own
- expand the edit scope beyond the handoff without Opus review
- introduce dependencies, build tooling, packaging, CI/CD, deployment, or external exposure changes unless explicitly listed
- touch secrets, credentials, `.env`, or local settings
- make final architectural decisions without returning the question to Opus/Codex

For very small edits, Opus may implement directly instead of creating unnecessary subagent overhead.
If the requested Claude Code environment cannot use subagents or the intended model split, Claude Code should continue with the available model and report that limitation.

## Decision Rule
Keep work in Codex when:
- requirements are ambiguous
- design intent or responsibility boundaries may change
- the task is small enough to edit and review in one context
- the main value is planning, review, or documentation consistency

Hand off to Claude Code when:
- goal, files, constraints, non-goals, and verification are clear
- the task is mostly implementation or mechanical editing
- the allowed edit scope can be stated explicitly
- Claude Code tooling or iteration speed is useful

## Project-Specific Guidance
- Use Raspberry Pi / Docker guidance from `D:/Git/CLAUDEmdStrage/_base`.
- Preserve `linux/arm64` compatibility unless the project explicitly supports more architectures.
- Do not change deployment, image naming, Portainer, or external exposure behavior without explicit approval.

## Files To Inspect First
- CLAUDE.md
- CLAUDE_ja.md
- README.md
- main.go
- Dockerfile
- docker-compose.yml

## Files Claude Code May Edit In Scoped Tasks
- main.go
- static/
- Dockerfile
- docker-compose.yml

## Constraints
- Keep Go dependencies minimal; prefer standard library.
- Preserve `network_mode: host` unless the network design is explicitly changed.
- Treat shutdown credentials and `devices.json` as sensitive.
- Do not weaken authentication or external exposure safeguards.
- Do not commit automatically unless explicitly requested.
- Do not revert user or other-agent changes unless explicitly requested.
- Do not edit secrets, credentials, `.env`, local runtime data, or generated heavy artifacts unless explicitly requested.

## Handoff Template
When Codex hands work to Claude Code, create `docs/handoffs/YYYY-MM-DD-<short-task>.md`. Create the `docs/handoffs/` directory if it does not exist. Use this format in that file.

```md
Read AGENTS.md, CLAUDE.md, and this handoff file before implementation.
If implementation would violate constraints or require files outside this handoff, stop and ask before editing.

## Goal
...

## Background
...

## Files To Inspect
- ...

## Files To Edit
- ...

## Constraints
- ...

## Non Goals
- ...

## Verification
- ...

## Expected Report
- Changed files
- Summary
- Verification results
- Blocked checks
- Design questions for Codex
```

## Codex Review Checklist
After Claude Code returns, review:
- Did the diff stay inside the handoff?
- Did any file outside `Files To Edit` change? If yes, was it necessary?
- Did the implementation preserve stated constraints and non-goals?
- Did it introduce dependencies, build tooling, packaging, CI/CD, deployment changes, or external exposure changes unexpectedly?
- Did it touch secrets, credentials, `.env`, local settings, or runtime data?
- Did verification run, and are blocked checks explained?
- Does any discovery need to become a new `AGENTS.md` or `docs/*.md` decision?

## Knowledge Persistence
- Use `AGENTS.md` for durable workflow and design decisions.
- Use `docs/*.md` for reusable technical notes, architecture details, procedures, and project-specific knowledge.
- Before meaningful work, check relevant existing docs.
- Do not silently encode durable design decisions only in code.

## Design Record Scope
Keep `AGENTS.md` focused on short, durable rules that future Codex and Claude Code sessions must follow.

Do not add `Alternatives Considered` as a default Decision Log heading. When rejected options or longer background matter, summarize only the durable rule in `AGENTS.md` and put the detail under `docs/decisions/`.
## Decision Log

### YYYY-MM-DD: Decision title

Context:
- What problem or requirement caused this decision?

Decision:
- What did we decide?

Reason:
- Why is this the right tradeoff now?

Constraints Introduced:
- What should future implementation preserve?

Do Not Change Casually:
- What would cause design drift if changed without review?
