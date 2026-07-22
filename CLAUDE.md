# CLAUDE.md

## Purpose

This file contains Claude Code execution rules for `WoL-tool-Claude`. Design intent, handoff policy, and Codex review belong in `AGENTS.md`.

## Read First

Before editing, read:

- `AGENTS.md` and this file.
- The supplied handoff or equivalent inline task scope permitted by `AGENTS.md`.
- The files listed for inspection and the existing nearby implementation.
- `README.md`, `Dockerfile`, `docker-compose.yml`, and relevant `docs/` notes when the task affects runtime, deployment, networking, authentication, or persistence.

## Instruction Handling

- Apply the instruction precedence defined in `AGENTS.md`.
- The handoff or equivalent inline scope is the approved task scope. It may narrow durable project constraints but may not weaken them.
- If instructions conflict, required files are outside the approved scope, or design remains unresolved, stop and return the issue to Codex instead of guessing.

## Project Shape

- `main.go` contains the Go 1.22 HTTP service, JSON store, Wake-on-LAN, ping, shutdown, authentication, and API handlers.
- `static/` contains the plain-JavaScript frontend.
- `Dockerfile` builds the dependency-free Go binary and the non-root Alpine runtime with iputils and samba-client.
- `docker-compose.yml` defines host networking, the `/data` bind mount, environment, restart behavior, and resource limits.
- The primary runtime is Raspberry Pi `linux/arm64`; publication also supports `linux/amd64`.

## Execution Rules

- Before editing, capture `git status --short`. After editing, compare the final status and diff with that baseline. Do not reset, clean, stage, or rewrite pre-existing changes.
- Implement only the current independently verifiable slice and wait for Codex review before starting another.
- Keep Go code standard-library-only and the frontend framework-free unless the approved task explicitly changes that constraint.
- Preserve `network_mode: host` and the UDP broadcast to `255.255.255.255:9`; do not replace host networking with port mapping without an approved network design.
- Keep mutable state outside the image. Preserve the `/data/devices.json` bind-mount boundary, non-root execution, host-directory protection, and file mode `0600`.
- Never return or log `shutdown_pass`. Preserve `has_shutdown_pass`, empty-password retention, explicit clearing, and stdin delivery to `net rpc`.
- Preserve the external-access boundary: internet exposure requires both `AUTH_USER`/`AUTH_PASS` Basic Auth and an upstream layer such as Cloudflare Access.
- Preserve the CSRF warning: no CORS response headers does not prevent cross-site simple form POST requests, and wake/shutdown CSRF protection is not implemented.
- Do not send Wake-on-LAN packets, ping or shut down real devices, or change remote Windows firewall/RPC settings during routine verification without explicit authorization.
- Return any proposed dependency, image, architecture, host-network, mount, persistence, authentication, deployment, CI/CD, registry, or external-exposure change outside the approved handoff to Codex.
- Subagents are optional and limited to clearly parallel mechanical work within the same files, scope, and constraints.
- On Windows, keep a delegated command line ASCII-only when its instructions contain non-ASCII text; put those instructions in a UTF-8 handoff file.

## Safety and Scope

- Preserve unrelated user and other-agent changes. Treat unexpected diffs as having unknown authorship and exclude them from the task unless confirmed.
- Do not inspect secrets, credentials, personal or device data, real `devices.json`, shutdown credentials, Basic Auth credentials, `.env`, production configuration, or runtime state unless their contents are strictly necessary for the approved task.
- Do not edit secrets, credentials, `.env`, local settings, device configuration, production data, runtime state, remote systems, or generated heavy artifacts unless explicitly required.
- Never reproduce secrets, credentials, personal or device data, private network values, or production configuration in prompts, handoffs, documentation examples, logs, reports, verification output, or external tools.
- Do not add dependencies or change build, publication, deployment, networking, persistence, authentication, or external exposure unless the approved task explicitly requires it.
- Do not commit, push, publish, or deploy unless explicitly requested.

## Verification

There is no test suite. Run the smallest relevant set of these verified commands:

- Documentation-only: `git diff --check -- AGENTS.md CLAUDE.md` and a focused reference scan.
- Go source: `gofmt -l .`, `go vet ./...`, and `go build ./...`.
- Docker or network behavior: only the focused check authorized by the task, using isolated device configuration.

`gofmt -l .` may report the known existing `main.go` formatting baseline. Do not change unrelated formatting; report baseline output separately. Report any unavailable Docker, target-host, or remote-device verification and why.

## Report

Return:

- Changed files and a concise summary.
- Verification commands and results, including blocked checks.
- Pre-existing changes preserved and any partial edits left in the worktree.
- Subagent usage.
- Design questions for Codex.
- If acceptance criteria are unmet, report `status=interrupted`, usable partial results, remaining scope, and the exact resume condition.

Report reusable discoveries to Codex. Update durable documentation only when it is inside the approved scope.
