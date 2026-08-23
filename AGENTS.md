# AGENTS.md

## Purpose

This is the Codex-side working agreement for `WoL-tool-Claude`, a lightweight web tool for Wake-on-LAN, ping monitoring, and remote Windows shutdown.

`AGENTS.md` owns design intent, model and handoff policy, Codex review, and documentation lifecycle. `CLAUDE.md` provides compatibility boundaries for implementation, verification, and reporting.

## Project Facts

- Runtime: a Go 1.22 single binary in Docker, primarily on Raspberry Pi `linux/arm64`.
- Frontend: plain JavaScript and static assets under `static/`.
- Go dependencies: standard library only; `go.mod` has no third-party modules.
- Runtime commands: `ping` from iputils and `net rpc shutdown` from samba-client.
- Container definitions: `Dockerfile` and `docker-compose.yml`; the image is published for `linux/amd64` and `linux/arm64`.
- `network_mode: host` is required for the UDP Wake-on-LAN broadcast to `255.255.255.255:9`.
- Persistent state is `/data/devices.json`, supplied through the established host bind mount.

## Instruction Precedence

When instructions conflict, apply them in this order:

1. Runtime, tool, organization, and safety policy.
2. Explicit user instructions that change project policy.
3. Durable project instructions.
4. Other instructions for the current user task and the approved task scope.

The active handoff or equivalent inline prompt is the approved task scope. Verified project facts override shared-source defaults. Only an explicit user instruction to change project policy may revise a durable project rule; other task instructions and approved scopes may narrow durable rules but may not weaken them. Report unresolved conflicts instead of guessing.

- Prefer the smallest correct change and reuse existing capabilities before adding dependencies or parallel policy.
- Approvals and completion require concise, evidence-backed scope, verification, and residual-risk/blocked-check reporting.

## Model and Role Policy

- Before implementation, classify the initial route from acceptance evidence: `small-primary` for small or transfer-negative work, `bounded` for settled multi-step work with one verifiable writer, `adaptive` when unresolved native, platform, runtime, or cross-subsystem behavior is material, or `non-implementation` for analysis, design, review, or operations. This classification does not force delegation; reclassify only after a material scope change or contract reset.
- The runtime-selected primary owns requirements, design, synthesis, and approval-sensitive decisions.
- Use native Codex roles: `bounded_implementer` is the cohesive default for settled work; choose `adaptive_implementer` directly when acceptance depends on unresolved native, platform, or cross-layer lifecycle behavior.
- Use `bounded_explorer` only for genuinely independent read-only questions and `bounded_reviewer` only when concrete correctness, security, compatibility, or verification risk warrants it. One active writer owns overlapping files or behavior.
- The writer's stable self-review gate is a dispatch barrier. If the writer changes the candidate after review starts, acceptance must be re-established; request a fresh final review only when material risk still warrants it. A second correction round, or two blocked/partial returns, requires a contract reset before continuing. If a selected role is unavailable or unobservable, use an observable equivalent or keep the work in the primary context.
- Name the concrete material risk in any reviewer handoff. Use a fresh task boundary for an independent phase with its own acceptance and verification; reintegrate delegated work from the stable diff and evidence instead of repeating its discovery.
- Claude Code is not an approved route unless an explicit policy change says so.
## Durable Project Rules

- Keep the service and frontend lightweight. Prefer the Go standard library and plain JavaScript; do not add dependencies or a framework without explicit approval.
- Preserve `network_mode: host` and UDP broadcast behavior unless the approved design explicitly changes the network model.
- Keep mutable state outside the image. Preserve the `/data` bind mount, non-root container execution, host-directory protection, and `devices.json` mode `0600`.
- `shutdown_pass` is stored in plaintext in `devices.json`. Never expose it through an API response, log, process argument, documentation example, or error. Responses expose only `has_shutdown_pass`; remote shutdown passes the password to `net rpc` through stdin.
- Preserve update semantics: an empty `shutdown_pass` retains the existing value, while `clear_shutdown_pass: true` explicitly removes it.
- The service assumes a trusted LAN. Internet exposure requires both configured `AUTH_USER`/`AUTH_PASS` Basic Auth and an upstream access boundary such as Cloudflare Access.
- Do not treat the absence of CORS headers as CSRF protection. Cross-site simple form POST requests can still reach wake and shutdown endpoints; CSRF protection is not implemented.
- Treat configured device addresses, remote Windows state, and network actions as operational data. Do not send Wake-on-LAN packets, ping or shut down real devices, or change remote firewall/RPC settings during routine verification without explicit authorization.
- Do not change the GHCR image, supported architectures, GitHub Actions publication, Portainer/Compose deployment, host networking, persistent mounts, authentication, or external exposure unless explicitly requested.

## Safety and Scope

- Preserve unrelated user and other-agent changes. Treat unexpected diffs as having unknown authorship and keep them outside the current task unless confirmed.
- Do not inspect secrets, credentials, personal or device data, real `devices.json`, shutdown credentials, Basic Auth credentials, `.env`, production configuration, or runtime state unless their contents are strictly necessary for the approved task.
- Do not edit secrets, credentials, `.env`, local settings, device configuration, production data, runtime state, remote systems, or generated heavy artifacts unless the approved task explicitly requires the change.
- Never reproduce secrets, credentials, personal or device data, private network values, or production configuration in prompts, handoffs, documentation examples, logs, reports, verification output, or external tools.
- Do not add dependencies or change build tooling, packaging, CI/CD, deployment, networking, persistence, authentication, or external exposure outside the approved task scope.
- Do not commit, push, publish, or deploy unless explicitly requested.

## Handoff Workflow

- Keep policy, design, review, read-only investigation, and small documentation corrections in Codex.
- Delegate only after the goal, files, constraints, non-goals, data sources, acceptance criteria, and verification are clear and material design choices are resolved.
- One handoff covers one cohesive, independently verifiable change and its direct verification. Run unresolved discovery as a separate read-only slice.
- State the goal, files to inspect and edit, constraints, non-goals, concrete data sources, acceptance criteria, verification, and expected report.
- Treat a delegation that ends before meeting its acceptance criteria as interrupted even when its process exits normally. Record usable partial results, verification, remaining scope, and the resume condition; narrow an over-broad handoff before rerunning it.
- The implementer follows the approved slice and returns design questions to the primary. The primary reviews the report and diff before starting another slice.
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
