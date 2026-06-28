# binman

Terminal macOS app uninstaller — pick apps and remove them completely (`.app` + `~/Library` leftovers → **Trash**), CleanMyMac-style, in your terminal. No GUI app. Just run `binman`.

> Status: **Minimal MVP**. Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss).

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
binman             # list → select many → confirm → Trash → results
binman -n          # preview only (dry-run)
binman -y          # skip the confirm screen and delete

binman clean                 # dry-run report (caches + logs) — default safe
binman clean --apply         # run the default cleanup
binman clean --xcode --apply # add Xcode artifacts
binman clean --all --apply   # everything (xcode + pkg + downloads)
```

**Flow** (`binman` with no arguments):

1. **List** — all installed apps (`/Applications`, `~/Applications`), with sizes.
2. **Select many** — multi-select the apps to remove (type to filter).
3. **Confirm** — aggregate screen: N apps, leftover count, total size.
4. **Process** — progress bar; items moved to Trash.
5. **Results** — summary (items moved, space freed).

**Keys:** `↑↓`/`jk` move · `space` toggle · `a` toggle all (in current filter) ·
type to **filter** · `enter` confirm · `q`/`ctrl+c` cancel. Shared/group
containers are off by default; system apps (`/System`) are shown marked but skipped.

## Safety principles

- Deletion = move to **Trash** (undoable via "Put Back"), never bare `rm` on user data.
- `-n` previews; `-y` skips confirm. `clean` defaults to dry-run and needs `--apply`.
- Never touches SIP-protected paths (`/System`, `/usr`, `/bin`, `/sbin`, `/private/var/db`).
- Quits a running app before deleting its data; group containers off by default.
- `/Library` (system, needs sudo) leftovers are skipped in the MVP.

## Roadmap (out of MVP)

`startup` (login items / launch agents), `health` (purge RAM / flush DNS / thin snapshots), app updater (`mas`), large/old files, browser privacy, undo history, Homebrew tap.
