# AGENTS.md

## Purpose

This is the Codex-side working agreement for `WoL-tool-Claude`, a lightweight web tool for Wake-on-LAN, ping monitoring, and remote Windows shutdown.

`AGENTS.md` owns design intent, model and handoff policy, Codex review, and documentation lifecycle. `CLAUDE.md` owns implementation, verification, and reporting rules.

## Project Facts

- Runtime: a Go 1.22 single binary in Docker, primarily on Raspberry Pi `linux/arm64`.
- Frontend: plain JavaScript and static assets under `static/`.
- Go dependencies: standard library only; `go.mod` has no third-party modules.
- Runtime commands: `ping` from iputils and `net rpc shutdown` from samba-client.
- Container definitions: `Dockerfile` and `docker-compose.yml`; the image is published for `linux/amd64` and `linux/arm64`.
- `network_mode: host` is required for the UDP Wake-on-LAN broadcast to `255.255.255.255:9`.
- Persistent state is `/data/devices.json`, supplied through the established host bind mount.

## Model and Role Policy

- Use GPT-5.3-Codex-Spark (`gpt-5.3-codex-spark`) proactively, when available, for low-risk, well-scoped, independently verifiable supporting work that requires no material design judgment or source-code implementation.
- GPT-5.6 Terra (`gpt-5.6-terra`) or Sol (`gpt-5.6-sol`) owns requirements and design. Whenever Terra is used, set its reasoning level to `high`. Prefer Sol for substantial ambiguity, risk, or cross-boundary reasoning.
- After design is fixed, delegate source-code implementation first to Claude Code Sonnet 5 at effort medium from the repository root: `claude -p --model sonnet --permission-mode auto "<handoff/task prompt>"`.
- Only when Sonnet 5 is unavailable because of usage limits or service availability, use GPT-5.6 Luna (`gpt-5.6-luna`) with reasoning level `max` for the same implementation slice.
- Implementation failure, failed verification, or a design question is not model unavailability; return it to Codex.
- Apply this policy to every coordinating Codex model and its subagents. Do not create coordinator-specific exceptions.
- Codex may keep requirements, design, read-only investigation, review, synthesis, and small documentation-consistency changes in one context.
- Claude Code subagents are optional and limited to clearly parallel mechanical work inside the approved handoff.

## Durable Project Rules

- Keep the service and frontend lightweight. Prefer the Go standard library and plain JavaScript; do not add dependencies or a framework without explicit approval.
- Preserve `network_mode: host` and UDP broadcast behavior unless the approved design explicitly changes the network model.
- Keep mutable state outside the image. Preserve the `/data` bind mount, non-root container execution, host-directory protection, and `devices.json` mode `0600`.
- `shutdown_pass` is stored in plaintext in `devices.json`. Never expose it through an API response, log, process argument, documentation example, or error. Responses expose only `has_shutdown_pass`; remote shutdown passes the password to `net rpc` through stdin.
- Preserve update semantics: an empty `shutdown_pass` retains the existing value, while `clear_shutdown_pass: true` explicitly removes it.
- The service assumes a trusted LAN. Internet exposure requires both configured `AUTH_USER`/`AUTH_PASS` Basic Auth and an upstream access boundary such as Cloudflare Access.
- Do not treat the absence of CORS headers as CSRF protection. Cross-site simple form POST requests can still reach wake and shutdown endpoints; CSRF protection is not implemented.
- Do not change the GHCR image, supported architectures, GitHub Actions publication, Portainer/Compose deployment, host networking, persistent mounts, authentication, or external exposure unless explicitly requested.
- Do not read or edit real `devices.json`, credentials, `.env`, runtime state, or generated artifacts unless explicitly required.
- Preserve unrelated changes. Do not commit, push, publish, or deploy unless explicitly requested.

## Handoff Workflow

- Keep policy, design, review, read-only investigation, and small documentation corrections in Codex.
- One handoff covers one cohesive, independently verifiable change and its direct verification. Run unresolved discovery as a separate read-only slice.
- State the goal, files to inspect and edit, constraints, non-goals, concrete data sources, verification, and expected report.
- If a handoff times out before its intended edit, do not rerun it unchanged. Narrow the behavior, files, and verification first.
- The implementer works only on the current slice and returns design questions to Codex. Codex reviews the report and diff before starting another slice.
- Keep active or blocked handoffs in `docs/handoffs/`. Move a handoff to `docs/handoffs/archive/` only after implementation, verification, review, required runtime work, and follow-up are complete.

## Verification and Review

The repository has no test suite. Use the focused applicable checks:

- `gofmt -l .`
- `go vet ./...`
- `go build ./...`
- `git diff --check`

`gofmt -l .` may report the known existing `main.go` formatting baseline; do not format unrelated code merely to clear that output. Report it distinctly from new formatting regressions.

During review, confirm that the diff stayed in scope, preserved host-network and persistent-data boundaries, did not expose passwords or weaken external access, introduced no unapproved dependency or deployment change, and reported blocked verification explicitly.

## Documentation Lifecycle

- Keep this file limited to short, current, durable rules and links.
- Put detailed decisions and evidence in `docs/decisions/`.
- Keep current decision guidance active; archive it only when fully implemented and no longer needed.
- Put reusable procedures and troubleshooting information in the appropriate `docs/` location.
- Do not rewrite completed handoffs or archived decisions merely to match a newer shared policy.
