# git-swap

Quickly switch between multiple git accounts across multiple hubs (GitHub, GitLab, Azure DevOps, or any custom host). Personal, work, client — swap identity, SSH key or HTTPS credential, and commit signing in one command.

## What it does

Each account is a **profile**. Switching a profile coordinates four concerns:

1. **Identity** — `user.name` / `user.email`
2. **Signing** — `user.signingkey` / `commit.gpgsign` (optional)
3. **Auth** — either an SSH key (via a managed `~/.ssh/config` block) or an HTTPS Personal Access Token (via git's credential helper)
4. **Remote** — optionally rewrite `origin` to the right URL for the account (`--rewrite-remote`)

Switching **identity** is separate from switching **credentials**. On a single hub (e.g. two GitHub accounts) the credential piece is what actually makes the second account work — that's the managed SSH block (ssh mode) or the credential username + seeded token (https mode).

## Install

```bash
go install github.com/jokot/git-swap@latest
```

Or build from source:

```bash
git clone https://github.com/jokot/git-swap && cd git-swap
go build -o git-swap .
```

Config lives at `~/.config/git-swap/config.yaml` (honors `XDG_CONFIG_HOME`). PATs are **never** stored there — only a reference (`token_env` / `token_file`) is.

## Auth modes

git-swap supports two auth modes per account via `--auth`:

- **`ssh`** (default) — authenticates with a private key. Two accounts on the same host would collide, so git-swap writes per-account `Host` aliases into a managed block in `~/.ssh/config` (e.g. `github.com-work` → `IdentityFile ~/.ssh/id_work`).
- **`https`** — authenticates with a Personal Access Token via git's credential helper. git-swap sets `credential.https://<host>.username` and seeds the token into the helper with `git credential approve`. The token is resolved from an env var or file **at switch time** and handed to git on stdin (never on the command line, never persisted by git-swap).

## Quickstart

```bash
# SSH account (default)
git-swap add personal --hub github --name "Jane Doe" --email jane@personal.com --ssh-key ~/.ssh/id_personal

# A second SSH account on the SAME hub (with commit signing)
git-swap add work --hub github --name "Jane Doe" --email jane@corp.com \
  --ssh-key ~/.ssh/id_work --sign --signing-key ~/.ssh/id_work.pub

# HTTPS account — token pulled from an env var at switch time
export GH_WORK_PAT=ghp_xxxxxxxx
git-swap add work-https --hub github --auth https \
  --name "Jane Doe" --email jane@corp.com \
  --username jane-work --token-env GH_WORK_PAT

# ...or from a file
git-swap add gl --hub gitlab --auth https --email jane@personal.com \
  --username jane --token-file ~/.secrets/gitlab_pat

# See what you have
git-swap list

# Switch machine-wide identity + auth
git-swap use work

# Switch for the current repo only
git-swap use personal --local

# Switch AND rewrite origin to the right URL for the mode
git-swap use work       --local --rewrite-remote acme/api  # git@github.com-work:acme/api.git
git-swap use work-https --local --rewrite-remote acme/api  # https://jane-work@github.com/acme/api.git

# Inspect / diagnose
git-swap current            # active identity + matching profile
git-swap current --local    # ...for the current repo
git-swap doctor             # check ssh keys exist / tokens resolve

# Remove a profile
git-swap remove work-https
```

## Commands

| Command | Purpose |
|---------|---------|
| `add <name>`    | Create or update a profile |
| `list`          | List profiles (name, hub, auth, host, email, cred, sign) |
| `use <name>`    | Switch identity + auth. `--local` for current repo, `--rewrite-remote <owner/repo>` to fix origin |
| `current`       | Show active identity and which profile matches (`--local` for repo scope) |
| `doctor`        | Report missing SSH keys / unresolvable HTTPS tokens |
| `remove <name>` | Delete a profile |

## How multi-account-on-one-hub works

**SSH:** git-swap owns a marker-delimited block in `~/.ssh/config` mapping host aliases to keys. Your remote must point at the alias (`git@github.com-work:acme/api.git`) — clone with it, or let `--rewrite-remote` set it. `IdentitiesOnly yes` ensures the right key is used deterministically.

**HTTPS:** git can't tell two `github.com` accounts apart by URL alone, so git-swap sets a host-scoped credential username and the remote embeds it (`https://jane-work@github.com/...`) so git selects the matching stored credential.

## Security & requirements

- **PATs are never stored by git-swap.** Only `token_env` / `token_file` is saved. The secret is resolved at switch time and piped to `git credential approve` on stdin.
- **HTTPS needs a configured credential helper** (`git config --global credential.helper osxkeychain|store|manager`). Without one, git won't remember the token. Note the plain `store` helper writes `~/.git-credentials` in cleartext; prefer `osxkeychain` on macOS.
- SSH signing users may also want `git config --global gpg.format ssh`.
