# binman

Terminal macOS app cleaner — uninstall apps completely (`.app` + `~/Library` leftovers → **Trash**) and clean system junk. CleanMyMac-style, in your terminal. No GUI app.

> Status: **Minimal MVP** — `uninstall` + `clean`. Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Install

Requires Go 1.25+ and macOS.

```bash
git clone <repo> binman && cd binman
make install        # installs to ~/go/bin (ensure it's on $PATH)
```

Optional (recommended for safe deletion): `brew install trash`

## Usage

```bash
binman uninstall <app>          # interactive TUI: review leftovers → confirm → Trash
binman uninstall Slack -y       # non-interactive (apply)
binman clean                    # dry-run report (default safe)
binman clean --apply            # run cleanup
binman clean --xcode --apply    # Xcode artifacts only
```

Global flag: `--dry-run/-n` — preview only, change nothing.

## Safety principles

- Deletion = move to **Trash** (undoable via "Put Back"), never bare `rm` on user data.
- `clean` defaults to **dry-run**; requires `--apply` to act.
- Never touches SIP-protected paths (`/System`, `/usr`, `/bin`, `/sbin`, `/private/var/db`).
- Excludes `com.apple.*` bundle IDs; warns before touching a running app's caches.
- `/Library` (system, needs sudo) leftovers are skipped with a hint in the MVP.

## Roadmap (out of MVP)

`list`/`info`, `startup` (login items / launch agents), `health` (purge RAM / flush DNS / thin snapshots), app updater (`mas`), large/old files, browser privacy, undo history, Homebrew tap.
