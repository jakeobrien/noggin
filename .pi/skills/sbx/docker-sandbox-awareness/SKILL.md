---
name: docker-sandbox-awareness
description: Explains how to recognize and correctly respond to Docker Sandbox (microVM) boundaries — restricted network/filesystem/tooling — versus genuine task errors. Use whenever a command fails with DNS errors, connection refused/timeout, proxy 403/407, TLS errors, missing tools, permission denied outside the workspace, or any other symptom that could be a sandbox policy restriction rather than a bug. Also use to record or look up prior sandbox policy decisions before re-raising a blocked request.
---

# Docker Sandbox Awareness

You may be running inside a **Docker Sandbox** — a microVM that isolates actions from
the host. Read this whenever you hit a wall that might be a sandbox boundary rather
than a genuine task/code error.

## The one rule that matters most

**The sandbox is conservative by default. A restriction almost never means the task
is impossible or forbidden — it usually means the resource hasn't been allow-listed
yet.**

When you hit a wall:

1. **Stop. Do not silently work around it.** Don't switch to an unverified mirror,
   disable TLS verification, vendor a random dependency copy, pipe a script from an
   unknown host, or re-architect the task to avoid the blocked resource.
2. **Name the constraint to the user**: what you tried, what blocked it, what you'd
   need.
3. **Ask for the specific unblock** — a domain to allow-list, a tool to
   pre-install, a credential, or a file to place in the workspace.
4. **Resume** once granted.
5. **Record the decision** once it's established (see [Recording Decisions](#recording-decisions) below).

The user can grant almost any of these selectively. Surfacing the constraint is the
helpful move; quietly routing around it hides risk.

## How the sandbox is built

You run in a **microVM with its own kernel**, with broad freedom inside it —
including `sudo`, package installation, and a private Docker Engine. The walls are
at the VM boundary, not inside it.

- **Workspace** — a single read-write directory, mounted live to the host. Treat it
  as production.
- **Network** — only **HTTP/HTTPS** to **allow-listed domains**, proxied through the
  host. Everything else is deny-by-default; raw TCP/UDP/ICMP is blocked.
- **Credentials** — API keys are injected into outbound headers by the host proxy.
  Raw secret values never enter the VM — you cannot read, echo, or copy them.
- **Docker** — you get your own Docker Engine, with no path to the host daemon.

**Blocked entirely:** host filesystem outside the workspace, host Docker daemon,
host network/`localhost`, other sandboxes, non-allow-listed domains.

**Persists across restarts:** installed packages, pulled images, other VM state.

### Policy is controlled by the user, not by you

Network and filesystem rules live in **local policy on the host**, editable only by
the user — you cannot inspect or change them from inside the sandbox. Infer the
boundary from failure symptoms (below) and ask the user to grant what's needed.

Default allow lists can include broad wildcards (e.g. `*.googleapis.com`), so
success against one endpoint doesn't confirm a narrow, intended grant. Flag
unexpectedly successful access to sensitive endpoints too.

## Common developer actions that collide with the sandbox

| What you try to do | What can block it | What to ask for |
|---|---|---|
| Install deps (`npm`/`pip`/`cargo`/`go mod`/`apt-get`/`gem`) | Registry/CDN domain not allow-listed | The exact registry/index/CDN domains |
| `git clone`/`fetch`/`push` | Host not allow-listed; SSH (`git@`) blocked (raw TCP denied) | Allow-list the Git host; use HTTPS, not SSH |
| `docker pull` | Registry domain not allow-listed | The registry domain (e.g. `ghcr.io`) |
| Web research / fetch docs / call an API | Domain not allow-listed | The specific domains, or local docs in the workspace |
| Use a missing CLI tool (`jq`, `gh`, a toolchain, linter, DB client) | Not installed in the image | Ask user to pre-install or approve install from an allowed source |
| Connect to a DB, queue, SSH host, or other non-HTTP service | Raw TCP/UDP blocked | An HTTP-based equivalent, or confirm reachability another way |
| Reach a service on the host/`localhost` | Host network not accessible from VM | Run the dependency inside the sandbox, or expose it over HTTP |
| Read/print an API key to verify it | Raw credentials never enter the VM | Confirm behavior via the actual proxied request instead |
| Read/write files outside the workspace | Only the workspace is mounted | Have the needed files placed in the workspace |

## Symptoms that mean "sandbox boundary," not a bug

Treat these as a likely policy/allow-list issue first — surface it rather than
retrying blindly or working around it:

- **DNS failures** — `could not resolve host`, `EAI_AGAIN`, `getaddrinfo ENOTFOUND`.
- **Connection refused/reset/timeout** to an external host, especially non-HTTP
  ports (DB, SSH/22, custom ports) — implies raw TCP/UDP is blocked.
- **Proxy-level rejections** — `403`/`407`, "forbidden/blocked by policy", or an
  HTML error page where you expected JSON/a tarball.
- **TLS/certificate errors** through the proxy — investigate the policy; **never
  disable certificate verification** to push past it.
- **Package manager errors** — network errors, "could not find a version," registry
  `404`/`503`.
- **`git` over SSH hanging or refused** — switch to HTTPS, then check the allow
  list.
- **`command not found`** for an expected tool — missing tooling, not a typo.
- **`permission denied`/`read-only file system`** outside the workspace, or failed
  attempts to reach the host Docker socket.
- **Empty, truncated, or placeholder API responses** — the request may have been
  silently rejected upstream.

## Quick checklist when blocked

1. Sandbox boundary (see symptoms above) or genuine task error?
2. Identify the specific blocked resource (domain, tool, file, service).
3. Tell the user: what you tried, the exact error, the resource needed.
4. Ask for the narrow grant (domain/tool/credential/file) — the user controls local
   policy.
5. **Never** silently work around the constraint.
6. Once resolved, record the decision (below).

## Recording Decisions

Every request for a sandbox config change — network policy additions, tool
installations, or any other environmental adjustment — must be recorded in
`dev-docs/sbx_policy_decisions.md` in the workspace (create it from the template
below if it doesn't exist yet).

**Declined requests must not be raised again in the same session** unless the user
re-opens the topic. Before asking for an unblock, check this file first — the
question may already have been answered.

Template for each entry (append to the file, one entry per decision):

```
YYYY-MM-DD

**Constraint:** <The action that was blocked by sandbox restrictions>

**Resolution:** <`GRANTED` | `DECLINED` | `PENDING`>

**Workaround:** <How to address (i.e. ask the user to do it, defer, work around)>

**Context:** (Optional) <Where this came up, affected code or process if helpful for future requests>
```

If `dev-docs/sbx_policy_decisions.md` doesn't exist, create it with this header:

```markdown
# SBX Policy Decisions

Records every agent request for a sandbox config change — network policy additions,
tool installations, or any other environmental adjustment. Declined requests must
**not** be raised again in the same session unless the user re-opens the topic.

---
```

---

*Reference: Docker Sandboxes [security model](https://docs.docker.com/ai/sandboxes/security/)
and [governance](https://docs.docker.com/ai/sandboxes/governance/) documentation.*
