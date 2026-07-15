# CLAUDE.md

## Purpose

This file contains Claude Code execution rules for `WoL-tool-Claude`. Design intent, handoff policy, and Codex review belong in `AGENTS.md`.

## Read First

Before editing, read:

- `AGENTS.md` and this file.
- The supplied handoff, when present.
- The files listed for inspection and the existing nearby implementation.
- `README.md`, `Dockerfile`, `docker-compose.yml`, and relevant `docs/` notes when the task affects runtime, deployment, networking, authentication, or persistence.

If the instructions conflict, required files are outside the approved scope, or design remains unresolved, stop and return the issue to Codex.

## Project Shape

- `main.go` contains the Go 1.22 HTTP service, JSON store, Wake-on-LAN, ping, shutdown, authentication, and API handlers.
- `static/` contains the plain-JavaScript frontend.
- `Dockerfile` builds the dependency-free Go binary and the non-root Alpine runtime with iputils and samba-client.
- `docker-compose.yml` defines host networking, the `/data` bind mount, environment, restart behavior, and resource limits.
- The primary runtime is Raspberry Pi `linux/arm64`; publication also supports `linux/amd64`.

## Execution Rules

- Implement only the current independently verifiable slice and wait for Codex review before starting another.
- Keep Go code standard-library-only and the frontend framework-free unless the approved task explicitly changes that constraint.
- Preserve `network_mode: host` and the UDP broadcast to `255.255.255.255:9`; do not replace host networking with port mapping without an approved network design.
- Keep mutable state outside the image. Preserve the `/data/devices.json` bind-mount boundary, non-root execution, host-directory protection, and file mode `0600`.
- Never return or log `shutdown_pass`. Preserve `has_shutdown_pass`, empty-password retention, explicit clearing, and stdin delivery to `net rpc`.
- Preserve the external-access boundary: internet exposure requires both `AUTH_USER`/`AUTH_PASS` Basic Auth and an upstream layer such as Cloudflare Access.
- Preserve the CSRF warning: no CORS response headers does not prevent cross-site simple form POST requests, and wake/shutdown CSRF protection is not implemented.
- Return any proposed dependency, image, architecture, host-network, mount, persistence, authentication, deployment, CI/CD, registry, or external-exposure change outside the approved handoff to Codex.
- On Windows, keep a delegated command line ASCII-only when its instructions contain non-ASCII text; put those instructions in a UTF-8 handoff file.

## Safety and Scope

- Preserve unrelated user and other-agent changes. Treat unexpected diffs as having unknown authorship and exclude them from the task.
- Do not read, edit, copy, or expose real `devices.json`, shutdown credentials, Basic Auth credentials, `.env`, production configuration, runtime state, or generated artifacts unless explicitly required.
- Do not put passwords in API responses, logs, command arguments, documentation examples, or verification output.
- Do not add dependencies or change build, publication, deployment, networking, persistence, authentication, or external exposure unless the approved task explicitly requires it.
- Do not commit, push, publish, or deploy unless explicitly requested.

## Verification

There is no test suite. Run the smallest relevant set of these verified commands:

- `gofmt -l .`
- `go vet ./...`
- `go build ./...`
- `git diff --check`

`gofmt -l .` may report the known existing `main.go` formatting baseline. Do not change unrelated formatting; report baseline output separately. Report any unavailable Docker or target-host verification and why.

## Report

Return:

- Changed files.
- Concise summary.
- Verification commands and results.
- Blocked checks.
- Subagent usage.
- Design questions for Codex.

Report reusable discoveries to Codex. Update durable documentation only when it is inside the approved scope.
