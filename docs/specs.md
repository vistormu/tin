# `tin` — tiny personal CLI toolbox

## Overview

`tin` is a small Go CLI for high-friction local file tasks. The current baseline includes media helpers, archive helpers, safe trashing, directory flattening, and size inspection.

The public binary name is `tin`. The local Go module name is also `tin`.

## Current progress

Implemented commands:

- `compress`
- `flatten`
- `resize`
- `size`
- `trash`
- `zip`
- `unzip`

Implemented infrastructure:

- Cobra root command and per-command help.
- Custom Cobra help template for richer command docs.
- Styled terminal output through Lip Gloss.
- Standardized top-level error rendering.
- Runtime platform detection for macOS/Linux support.
- Runtime install hints for external tools.
- Shared filesystem, confirmation, size, help-text, and output helpers.
- Cross-build workflow for macOS/Linux amd64/arm64.

## Platform support

v1 supports macOS and Linux only. Unsupported platforms, including Windows, return clear errors.

Platform selection happens at runtime with `runtime.GOOS` inside internal platform/trash selection code. Command implementations should use internal interfaces instead of branching directly on the operating system. v1 does not use Go build tags.

## Tech stack

- **Language**: Go
- **CLI framework**: Cobra
- **Go dependencies**: Cobra and Lip Gloss
- **Runtime tools**: `ghostscript`, `magick`, and `cwebp` when the related command path needs them

## Project structure

```text
tin/
├── main.go
├── go.mod
├── Justfile
├── cmd/
│   ├── root.go
│   ├── compress.go
│   ├── flatten.go
│   ├── resize.go
│   ├── size.go
│   ├── trash.go
│   ├── unzip.go
│   └── zip.go
└── internal/
    ├── archive/
    ├── compress/
    ├── cwebp/
    ├── flatten/
    ├── ghostscript/
    ├── installhint/
    ├── magick/
    ├── platform/
    ├── resize/
    ├── shared/
    │   ├── cobrahelp/
    │   ├── confirm/
    │   ├── fileutil/
    │   ├── fsx/
    │   ├── helptext/
    │   ├── size/
    │   └── ui/
    └── trash/
```

## Safety policy

Commands that can remove or move user data must show the exact target list before acting. Destructive or hard-to-undo actions confirm by default and support a dry-run mode when useful.

`tin trash` and `tin flatten` confirm by default. `--dry-run` prints planned changes only. `--yes` skips confirmation.

## Terminal output

User-facing command output is rendered through `internal/shared/ui`. The UI layer uses Lip Gloss for terminal-aware colors and simple boxes while keeping command packages independent from styling details.

Color policy:

- `NO_COLOR` disables styled color output.
- `CLICOLOR=0` disables styled color output unless forced.
- `CLICOLOR_FORCE=1` forces styled color output.
- Otherwise, color support is detected from the output writer.

Top-level command errors are rendered consistently as `Error: <message>` without printing usage for runtime or validation errors.

## Install hints

External command wrappers expose `Check()` methods. Missing-tool messages are improved by detecting package managers from `PATH`, not by assuming from the operating system.

Package manager candidates:

- `brew`
- `apt`
- `dnf`
- `pacman`
- `apk`
- `zypper`

If none is found, `tin` shows a generic package-manager hint.

## Commands

### `tin compress <file> [-q quality] [-o output]`

Compresses a supported file by extension.

Supported input types:

- `.pdf` uses `ghostscript`
- `.jpg`, `.jpeg`, and `.png` use `magick`
- `.webp` uses `cwebp`

Default output is `<name>-compressed.<ext>`.

The command prints:

- before size
- after size
- percent reduction
- output path

Unsupported file types and missing tools return clear errors.

Run `tin compress qualities` or `tin compress --help` to see supported quality values. PDF quality supports `screen`, `ebook`, `printer`, `print`, `prepress`, and `default`. Image quality supports integers from `1` to `100`.

### `tin resize <file> -w width [-H height] [-o output]`

Resizes a supported image file using ImageMagick.

Supported input types:

- `.jpg`
- `.jpeg`
- `.png`
- `.webp`

Default output is `<name>-resized.<ext>`.

The command prints:

- before size
- after size
- output path

Width must be greater than `0`. Height is optional; when omitted, aspect ratio is preserved by ImageMagick.

### `tin flatten <dir> [--dry-run] [--yes]`

Moves all nested files into the given directory root and removes empty subdirectories.

Behavior:

- Requires a directory root.
- Plans moves before changing anything.
- Shows exact source-to-destination moves.
- Confirms by default.
- `--dry-run` prints the move/remove plan only.
- `--yes` skips confirmation.
- Leaves files already in the root directory in place.
- Handles regular files and symlinks.
- Rejects unsupported file types.
- Avoids overwriting root files or planned destinations.
- Uses extension-preserving collision names like `file.1.txt`.
- Removes empty subdirectories deepest-first after moving files.

The command prints planned move count, planned directory removal count, and final moved/removed counts.

### `tin size <path>`

Shows a human-readable size for a file or directory.

Implementation uses pure Go recursive filesystem walking, not `du`.

### `tin trash <file|glob>... [--dry-run] [--yes]`

Moves files to trash.

Behavior:

- Expands quoted shell-style glob patterns.
- Shows exact targets before trashing.
- Confirms by default.
- `--dry-run` prints targets only.
- `--yes` skips confirmation.
- Avoids overwriting existing trash entries by choosing unique names like `file.1`.
- Handles regular files, directories, and symlinks.

macOS moves files into the user's `~/.Trash` directory without requiring the external `trash` CLI.

Linux implements XDG Trash under:

- `$XDG_DATA_HOME/Trash`
- fallback `~/.local/share/Trash`

Linux writes `.trashinfo` metadata so file managers can restore trashed files.

### `tin zip <path>... [-o archive.zip]`

Creates a `.zip` archive using Go's standard library.

Behavior:

- Accepts files and directories.
- Recursively archives directory contents.
- Default output is `<single-input-name>.zip` for one input or `archive.zip` for multiple inputs.
- Refuses to overwrite an existing output archive.
- Refuses unsupported file types such as symlinks.

The command prints file count, uncompressed byte count, and output path.

### `tin unzip <archive.zip> [-o dir]`

Extracts a `.zip` archive using Go's standard library.

Behavior:

- Default output directory is the archive name without `.zip`.
- Refuses non-zip archive extensions.
- Refuses unsafe archive paths such as `../file`.
- Refuses to overwrite existing destination files.
- Refuses unsupported entries such as symlinks.

The command prints file count, extracted byte count, and output directory.

## Deferred commands

The following broader toolbox ideas are deferred:

- convert
- merge
- split
- find
- dedup
- diff
- count
- disks
- mount
- unmount
- eject
- open
- ip
- kill
- usage
- git subcommands

## Installation

```sh
just install
```

This builds and copies `tin` to `~/.local/bin/tin` unless `PREFIX` is overridden.

## Builds

```sh
just build
```

Builds release binaries into `dist/` for:

- macOS amd64
- macOS arm64
- Linux amd64
- Linux arm64

```sh
just build-local
```

Builds the current host binary into `dist/tin`.
