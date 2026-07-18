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

## Adding a git account — complete walkthrough

This is the full flow for registering an account and switching to it, start to finish.

### Before you start: gather the account's details

You need:
- **A name for the profile** — anything you like (`work`, `personal`, `client-x`).
- **git identity** — the `user.name` and `user.email` to stamp on commits. This is independent of the GitHub/GitLab login; it's just what shows on your commits.
- **The credential**, which depends on auth mode:
  - **SSH** (default): the path to the private key for this account, e.g. `~/.ssh/id_personal`.
  - **HTTPS**: a username and a Personal Access Token (PAT). The PAT goes in an env var or file — **never** typed into the command directly.

Not sure which SSH key belongs to which account? These help:

```bash
# What each public key was labelled with (usually an email)
for k in ~/.ssh/id_*.pub; do printf "%s: " "$k"; cut -d' ' -f3- "$k"; done

# Which remote host uses which key today
cat ~/.ssh/config

# Which account a key actually authenticates as (GitHub replies "Hi <user>!")
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_personal -T git@github.com
```

### Option A — Add an SSH account (most common)

**1. Add the profile.** `add` creates a new profile, or updates an existing one with the same name.

```bash
git-swap add personal \
  --hub github \
  --name "Jane Doe" \
  --email jane@personal.com \
  --ssh-key ~/.ssh/id_personal
```

**2. Switch to it.** This writes your identity and refreshes git-swap's managed block in `~/.ssh/config`.

```bash
git-swap use personal            # machine-wide (global)
# or, for just the current repo:
git-swap use personal --local
```

**3. Verify.**

```bash
git-swap current                 # active identity + matching profile
git config --global user.email   # should print jane@personal.com
```

Done. For a repo to use this account's SSH identity, its `origin` must point at the alias host — either clone with it (`git clone git@github.com-personal:owner/repo.git`) or let git-swap rewrite it: `git-swap use personal --local --rewrite-remote owner/repo`.

### Option B — Add an HTTPS account (PAT)

**1. Put the PAT in an env var** (or a file) — keep it out of your shell history and the command line:

```bash
export GH_PERSONAL_PAT=ghp_xxxxxxxx      # or store it in a file and use --token-file
```

**2. Add the profile**, referencing the env var (git-swap stores the *reference*, never the token):

```bash
git-swap add personal-https \
  --hub github --auth https \
  --name "Jane Doe" --email jane@personal.com \
  --username jane \
  --token-env GH_PERSONAL_PAT
```

**3. Make sure git has a credential helper** (one-time, so git remembers the token):

```bash
git config --global credential.helper osxkeychain   # macOS (recommended)
# git config --global credential.helper store        # Linux — note: cleartext file
```

**4. Switch to it** (the PAT must be resolvable — env var set or file present):

```bash
export GH_PERSONAL_PAT=ghp_xxxxxxxx
git-swap use personal-https
```

**5. Verify.**

```bash
git config --global credential.https://github.com.username   # should print jane
git-swap doctor                                               # confirms the token resolves
```

### Optional — Enable commit signing (SSH signing)

Signing cryptographically proves a commit came from you (GitHub shows a green **Verified** badge). To sign with the SSH key you already have:

**1. Add (or re-add) the profile with `--sign` and the *public* key as the signing key:**

```bash
git-swap add personal \
  --hub github --name "Jane Doe" --email jane@personal.com \
  --ssh-key ~/.ssh/id_personal \
  --sign --signing-key ~/.ssh/id_personal.pub
```

**2. Tell git to sign using SSH format — this is a ONE-TIME setting** (git defaults to GPG otherwise, which would make every commit fail against an SSH key):

```bash
git config --global gpg.format ssh
```

You only ever run this once. It stays in your global git config and applies to every SSH-signing account. You do **not** repeat it per switch. (The only time you'd revisit it is if you also had a GPG-signing account and needed to flip `gpg.format` back to `openpgp`.)

**3. Switch, then verify:**

```bash
git-swap use personal
git config --global --get user.signingkey    # the .pub path
git config --global --get commit.gpgsign     # true
git config --global --get gpg.format         # ssh
```

**4. For GitHub's Verified badge**, add the *same* public key at <https://github.com/settings/keys> as a **Signing Key** (this is separate from an Authentication key — one key can be registered as both).

### Turning signing OFF

`git-swap add` sets `commit.gpgsign=true` only when you pass `--sign`. To turn signing off, re-add the profile **without** `--sign`, then clear the leftover global settings (git-swap sets these flags but never unsets them):

```bash
git-swap add personal --hub github --name "Jane Doe" --email jane@personal.com --ssh-key ~/.ssh/id_personal
git config --global --unset commit.gpgsign    # no output = success
git config --global --unset user.signingkey   # no output = success
```

Empty output from `--unset` / `--get` means the setting is gone — that's the success case. Commits will now go through unsigned.

### Good to know

- **Identity ≠ login.** `--name` / `--email` only affect what's stamped on commits. They don't have to match the GitHub/GitLab account name.
- **`add` is idempotent.** Re-running `add` with an existing profile name overwrites it — that's how you edit a profile.
- **`use` is what applies changes.** `add` only saves the profile; nothing touches your git or ssh config until you run `use`.
- **HTTPS vs SSH is about auth (push/pull); GPG vs SSH signing is about signing.** They're independent — an HTTPS account can still sign with an SSH key.

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
