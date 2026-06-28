# binman

Terminal macOS app cleaner — uninstall apps completely (`.app` + `~/Library` leftovers → **Trash**) and clean system junk. CleanMyMac-style, in your terminal. No GUI app.

> Status: **Minimal MVP** — `uninstall` + `clean`. Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Install

Requires Go 1.25+ and macOS.

```bash
git clone <repo> binman && cd binman
make install        # builds and installs to ~/go/bin (ensure it's on $PATH)
# or just: make build && ./binman
```

Safe deletion uses the macOS `trash` tool — on macOS 15.2+ it ships built-in
(`/usr/bin/trash`); on older systems run `brew install trash`. If absent, binman
falls back to AppleScript (Finder → Trash).

## Usage

```bash
binman list                     # list installed apps with sizes
binman uninstall                # no app arg + terminal: pick an app from a list, then uninstall
binman uninstall <app>          # interactive TUI: review leftovers → confirm → Trash
binman uninstall Slack -y       # non-interactive: delete without prompting
binman uninstall Slack -n       # preview only (dry-run)

binman clean                    # dry-run report (caches + logs) — default safe
binman clean --apply            # run the default cleanup
binman clean --xcode --apply    # add Xcode artifacts
binman clean --pkg --apply      # add brew/npm/pnpm/pip/docker cleanup
binman clean --downloads --apply# add old installers in ~/Downloads
binman clean --all --apply      # everything
```

Global flag: `--dry-run/-n` — preview only, change nothing.

**`uninstall` TUI keys:** `↑↓`/`jk` move · `space` toggle · `a` toggle all ·
`enter` confirm · `q`/`esc` cancel. Group/Shared containers are unchecked by
default to avoid removing shared data.

**App picker** (`binman uninstall` with no arg): type to filter · `↑↓` move ·
`enter` select · `q`/`ctrl+c` cancel. Lists user-installed apps in
`/Applications` (system apps in `/System` are excluded).

## Safety principles

- Deletion = move to **Trash** (undoable via "Put Back"), never bare `rm` on user data.
- `clean` defaults to **dry-run**; requires `--apply` to act.
- Never touches SIP-protected paths (`/System`, `/usr`, `/bin`, `/sbin`, `/private/var/db`).
- Quits a running app before deleting its data; group containers off by default.
- `/Library` (system, needs sudo) leftovers are reported but skipped in the MVP.

## Roadmap (out of MVP)

`info`, `startup` (login items / launch agents), `health` (purge RAM / flush DNS / thin snapshots), app updater (`mas`), large/old files, browser privacy, undo history, Homebrew tap. (`list` + app picker shipped.)
