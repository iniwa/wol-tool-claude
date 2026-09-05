# AGENTS.md

This entry governs the Wake-on-LAN, ping, and remote Windows shutdown tool.

## Verified project facts and protected behavior
- Go 1.22 standard-library service with plain JavaScript frontend; Docker targets linux/arm64 and linux/amd64.
- Preserve network_mode: host and UDP broadcast 255.255.255.255:9; mutable state is /data/devices.json on the established bind mount, non-root, mode 0600.
- shutdown_pass remains plaintext in devices.json but is never exposed in APIs, logs, arguments, docs, or errors; send it to net rpc through stdin. Empty update retains it; clear_shutdown_pass explicitly removes it. Responses expose only has_shutdown_pass.
- Internet exposure requires configured Basic Auth plus an upstream boundary such as Cloudflare Access. No CORS is not CSRF protection. Never contact real devices or alter remote firewall/RPC settings during routine checks.

## Authority and scope
Apply runtime, tool, organization, and safety policy, then explicit user policy, then this entry and the approved task. Verified repository facts replace defaults; they do not grant authorization. Preserve unrelated work and stop on an overlap that requires guessing.

## Execution
Choose the smallest correct change. The user selects the runtime model and effort; role configuration owns model, effort, and role instructions. Use one bounded writer for settled work, adaptive implementation only for material native/platform uncertainty, and read-only exploration or review only when independently useful. Keep one writer for overlapping files. A changed candidate after review must be restabilized; after a second correction or two blocked returns, reset the contract before continuing. Persisted handoffs are for named cross-session, interruption-sensitive, risky, or separately executed work; otherwise use the approved inline scope. Optional cheap direct regression tests are appropriate when they materially support changed behavior; do not require a new harness or full suite by default.

## Safety
Do not inspect or edit secrets, credentials, local settings, runtime or production state, generated heavy artifacts, dependencies, CI/CD, deployment, publication, or external exposure unless explicitly in scope. Never reproduce private values. Do not commit, push, or publish unless explicitly requested. Report source readiness separately from unavailable runtime verification.

## Completion
Review the stable diff against every criterion and protected behavior, verify affected references and Markdown fences, run the smallest relevant checks plus git diff --check, and report changed files, evidence, blocked checks, partial edits, and unresolved questions.

## Checks
The repository has no test suite. Use gofmt -l ., go vet ./..., go build ./..., and git diff --check; report the known existing main.go formatting baseline separately.
