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

### macOS / Linux (Homebrew)
*Coming soon! (Requires a separate homebrew-tap repository).*

### Direct Download (macOS / Linux / Windows)
Head to the [Releases page](https://github.com/jokot/git-swap/releases) and download the `.tar.gz` (Mac/Linux) or `.zip` (Windows) for your OS and architecture. 
Extract it and put the `git-swap` executable somewhere in your `PATH` (like `/usr/local/bin`).

### From Source (Requires Go)

```bash
go install github.com/jokot/git-swap@latest
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

```
═══════════════════════════════════════════════
STEP 0 — Decide the account's auth type
═══════════════════════════════════════════════
Look at how you clone/push for this account:
    • remote looks like  git@github.com:owner/repo.git   → SSH
    • remote looks like  https://github.com/owner/repo.git → HTTPS (needs a PAT)

If unsure, SSH is the simpler default and matches your current setup.

═══════════════════════════════════════════════
PATH A — SSH ACCOUNT
═══════════════════════════════════════════════

A1. Find (or create) the SSH key for this account
    List existing private keys:
    ls -1 ~/.ssh/id_* | grep -v '.pub$'
    Check which account a key belongs to (reads the email comment):
    ssh-keygen -lf ~/.ssh/id_ed25519_jokot.pub
    Need a new key? Create one:
    ssh-keygen -t ed25519 -C "you@example.com" -f ~/.ssh/id_ed25519_myacct
    Then add its .pub to GitHub → Settings → SSH and GPG keys → Authentication key.

A2. Add the profile (signing OFF — the simple case)
    git-swap add <name> \
        --hub github \
        --name "Your Name" \
        --email you@example.com \
        --ssh-key ~/.ssh/id_ed25519_myacct

    (--hub can be github | gitlab | azure | custom.
        For a non-default host add:  --host git.company.com)

A3. (Optional) Turn on SSH commit signing
    Re-add the same profile WITH signing:
    git-swap add <name> --hub github --name "Your Name" \
        --email you@example.com --ssh-key ~/.ssh/id_ed25519_myacct \
        --sign --signing-key ~/.ssh/id_ed25519_myacct.pub
    Set the signing format ONCE (global, never per-switch):
    git config --global gpg.format ssh
    Add the same .pub to GitHub as a Signing key (Settings → SSH and GPG keys).
    Skip this whole step if you don't want signing.

A4. Verify it's registered
    git-swap list

═══════════════════════════════════════════════
PATH B — HTTPS ACCOUNT (uses a PAT, not a key)
═══════════════════════════════════════════════

B1. Get a Personal Access Token
    GitHub → Settings → Developer settings → Personal access tokens.
    Copy it — but do NOT paste it into any chat or config file.

B2. Put the token in your shell, by reference (never inline)
    export MYACCT_PAT=ghp_xxxxxxxx
    (Put that line in ~/.zshrc if you want it to persist.)

B3. Add the profile pointing at the env var
    git-swap add <name> \
        --hub github --auth https \
        --name "Your Name" \
        --email you@example.com \
        --username <github-login> \
        --token-env MYACCT_PAT
    (Or read from a file instead:  --token-file ~/.secrets/pat)

B4. Ensure git has a credential helper (needed for HTTPS)
    git config --global credential.helper osxkeychain    # macOS, secure

B5. Verify
    git-swap list
    git-swap doctor        # confirms the token is resolvable

═══════════════════════════════════════════════
STEP 1 — Switch to the account (applies the config)
═══════════════════════════════════════════════
Adding only SAVES a profile; nothing changes until you switch:

    git-swap use <name>              # machine-wide (global)
    git-swap use <name> --local      # current repo only

When using an SSH account, git-swap will prompt you to automatically set an `insteadOf` alias. This intercepts Git URLs (like `git@github.com:`) and routes them through the correct SSH key alias (`github.com-<name>`) so your pushes just work without manually changing remotes:

    Update remote alias (insteadOf) for this account?
      [l] Local repo only (default)
      [g] Global (all repos)
      [s] Skip

(Passing `--local` skips the prompt and applies the alias locally automatically).

Confirm what's active:

    git-swap current                 # identity + matching profile
    git config --global user.email   # sanity check

═══════════════════════════════════════════════
TURNING SIGNING OFF (if you enabled it earlier)
═══════════════════════════════════════════════
Re-add WITHOUT --sign, then clear the leftover global flags
(git-swap sets these but never unsets them):

    git-swap add <name> --hub github --name "Your Name" \
        --email you@example.com --ssh-key ~/.ssh/id_ed25519_myacct
    git config --global --unset commit.gpgsign     # no output = success
    git config --global --unset user.signingkey    # no output = success

═══════════════════════════════════════════════
GOOD TO KNOW
═══════════════════════════════════════════════
    • Identity ≠ login. --name / --email only stamp commits; they need
      not match the GitHub/GitLab account.
    • add is idempotent — re-running it with the same <name> overwrites
      that profile. That's how you edit one.
    • use is what applies changes; add alone touches nothing.
    • Auth (SSH vs HTTPS) is about push/pull. Signing (SSH vs GPG) is a
      separate concern — an HTTPS account can still sign with an SSH key.
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
