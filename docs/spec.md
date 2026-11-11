# `tin` — tiny personal CLI toolbox

## Overview

`tin` is a small Go CLI for high-friction local file tasks. The current baseline includes media helpers, archive helpers, safe trashing, directory flattening, and size inspection.

The public binary name is `tin`. The local Go module name is also `tin`.

## Current progress

Implemented commands:

- `compress`
- `convert`
- `count`
- `dedup`
- `find`
- `flatten`
- `git undo`
- `git obliterate`
- `ip`
- `merge`
- `new`
- `open`
- `rename`
- `replace`
- `resize`
- `size`
- `split`
- `trash`
- `zip`
- `unzip`

Implemented infrastructure:

- Cobra root command and per-command help.
- Custom Cobra help template for richer command docs.
- Styled terminal output through Lip Gloss.
- Standardized top-level error rendering.
- Shared shell-style glob expansion for commands that accept multiple inputs.
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
- **Runtime tools**: `ghostscript`, `magick`, `cwebp`, `ffmpeg`, `pdftocairo`/`pdfinfo` (poppler), `git`, and `git-filter-repo` when the related command path needs them

## Project structure

```text
tin/
├── main.go
├── go.mod
├── justfile
├── cmd/
│   ├── root.go
│   ├── compress.go
│   ├── convert.go
│   ├── count.go
│   ├── dedup.go
│   ├── find.go
│   ├── flatten.go
│   ├── git.go
│   ├── ip.go
│   ├── merge.go
│   ├── new.go
│   ├── open.go
│   ├── rename.go
│   ├── replace.go
│   ├── resize.go
│   ├── size.go
│   ├── split.go
│   ├── trash.go
│   ├── unzip.go
│   └── zip.go
└── internal/
    ├── archive/
    ├── compress/
    ├── convert/
    ├── count/
    ├── cwebp/
    ├── dedup/
    ├── ffmpeg/
    ├── find/
    ├── flatten/
    ├── git/
    ├── ghostscript/
    ├── installhint/
    ├── ip/
    ├── magick/
    ├── merge/
    ├── opener/
    ├── platform/
    ├── poppler/
    ├── rename/
    ├── replace/
    ├── resize/
    ├── shared/
    │   ├── cobrahelp/
    │   ├── confirm/
    │   ├── fileutil/
    │   ├── fsx/
    │   ├── helptext/
    │   ├── size/
    │   └── ui/
    ├── split/
    └── trash/
```

## Safety policy

Commands that can remove or move user data must show the exact target list before acting. Destructive or hard-to-undo actions confirm by default and support a dry-run mode when useful.

`tin trash`, `tin flatten`, `tin rename`, `tin replace`, `tin dedup`, `tin git undo --hard`, and `tin git obliterate` confirm destructive changes by default. `--dry-run` prints planned changes only where supported. `--yes` skips confirmation.

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

## Input expansion

Commands that accept one or more inputs use a shared shell-style glob expander.

- Quoted glob patterns like `'*.pdf'` expand against the working directory.
- Literal paths must exist.
- A pattern that matches no files is an error.
- Duplicate paths are collapsed.
- Commands that produce a per-input output file disallow `-o` when more than one input is resolved.

## Commands

### `tin compress <file|glob>... [-i intensity] [-o output] [--in-place]`

Compresses one or more supported files by extension. Accepts file paths and shell-style glob patterns.

Supported input types:

- `.pdf` uses `ghostscript`
- `.jpg`, `.jpeg`, and `.png` use `magick`
- `.webp` uses `cwebp`
- `.mp4`, `.mov`, and `.mkv` use `ffmpeg` (H.264/CRF, AAC 128k audio)

Default output is `<name>-compressed.<ext>`.

The command prints:

- before size
- after size
- percent reduction
- output path

Unsupported file types and missing tools return clear errors. `-o` is only allowed with a single input; for multiple inputs each result is written to the default location. `--in-place` replaces each input file with its compressed output via an atomic temp-file write and rename in the same directory; it cannot be combined with `-o`.

Run `tin compress intensities` or `tin compress --help` to see supported intensity values. Intensity uses a single uniform vocabulary for PDFs, images, and videos:

| Intensity   | PDF preset | Image quality | Video CRF |
|-------------|------------|---------------|-----------|
| `low`       | `prepress` | 90            | 20        |
| `medium`    | `printer`  | 80            | 23        |
| `high`      | `ebook`    | 65            | 28        |
| `veryhigh`  | `screen`   | 45            | 32        |

Default intensity is `medium`.

### `tin count <path|glob>... [-e ext]`

Walks one or more files or directories and reports file, line, and byte totals. Accepts paths and shell-style glob patterns.

Behavior:

- Recurses into directories, skipping symlinks and non-regular files.
- Counts newlines as lines, matching `wc -l` semantics (a file without a trailing newline contributes 0 lines).
- `-e` / `--ext` accepts a comma-separated list of extensions (`go,md`) to restrict the count to.

The command prints file count, line count, and total bytes.

### `tin dedup <dir> [--dry-run] [--yes]`

Detects duplicate files inside a directory by content hash and moves duplicates to the system trash.

Behavior:

- Walks the directory recursively. Skips symlinks and non-regular files.
- Groups files by size first; only candidates with matching sizes are hashed (SHA-256).
- Within each duplicate group, keeps the alphabetically-first path and marks the rest for trashing.
- Confirms by default.
- `--dry-run` prints the duplicate plan only.
- `--yes` skips confirmation.
- Uses the same trash backend as `tin trash`, so removed files can be restored.

The command prints the number of duplicate groups, removable file count, recoverable bytes, and per-group keep/trash listings.

### `tin find [dir] [-e ext] [-n glob] [--type f|d|all]`

Recursively walks a directory and prints matching paths, one per line. Defaults to the current directory.

Behavior:

- `-e` / `--ext` accepts a comma-separated list of extensions (`go,md`), matched case-insensitively.
- `-n` / `--name` matches the basename against a shell-style glob pattern.
- `--type` selects entries to print: `f` for regular files (default), `d` for directories, `all` for both.
- All filters combine with AND semantics.

### `tin convert <file|glob>... [-t target] [-o output]`

Converts one or more supported files to another format. The router picks a backend from the (input, target) pair: ImageMagick for image conversions and pdf↔image, poppler (`pdftocairo`, `pdfinfo`) for pdf→svg, ffmpeg for video, audio, and video→audio. Accepts file paths and shell-style glob patterns. Target type is taken from `-t` when provided, otherwise inferred from the `-o` output extension. At least one of `-t` or `-o` must be given.

Supported formats by family:

- **image**: `.jpg`, `.jpeg`, `.png`, `.webp`, `.heic`, `.tiff`, `.tif`
- **vector**: `.svg`
- **pdf**: `.pdf`
- **video**: `.mp4`, `.mov`, `.mkv`, `.webm`, `.avi`
- **audio**: `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`, `.opus`

Supported routes:

| From → To | Backend | Notes |
|-----------|---------|-------|
| image → image | ImageMagick | One output per input. |
| image → pdf (1 input) | ImageMagick | Writes `<name>.pdf` next to the input. |
| image → pdf (N inputs) | ImageMagick | Combines inputs into a single multi-page PDF. Default output is `<first-input-dir>/combined.pdf`. `-o` is allowed. |
| pdf → image | ImageMagick | One image per page, written as `<name>-001.<ext>`, `<name>-002.<ext>`, ... next to the input. `-o` is not allowed. |
| pdf → svg | poppler | Vector output, one SVG per page. Single-page PDFs write `<name>.svg` and accept `-o`; multi-page PDFs write `<name>-001.svg`, `<name>-002.svg`, ... next to the input and reject `-o`. |
| video → video | ffmpeg | Re-encodes using container defaults. |
| video → audio | ffmpeg | Strips the video stream. |
| audio → audio | ffmpeg | Re-encodes using container defaults. |

`-o` is only allowed with a single input, except for the multi-image→pdf combine case. Refuses to overwrite existing outputs and refuses no-op conversions where the input is already the target format; `.jpeg`/`.jpg` and `.tif`/`.tiff` are treated as the same format for that check. Inputs in a single invocation must share a family.

The command prints one box per operation with before size, after size, and the produced output path or page summary.

### `tin ip`

Shows network address information for the host.

Behavior:

- Lists active local network interfaces with their IPv4 and IPv6 addresses, skipping interfaces that are down or expose only link-local addresses.
- Fetches the current public IP from `https://api.ipify.org` with a 5-second timeout.
- Reports the public lookup error inline instead of failing the whole command.

The command prints one box per scope: a `local` box with per-interface addresses and a `public` box with the resolved address.

### `tin new <path>...`

Creates one or more files or directories.

Behavior:

- A trailing `/` marks a directory; everything else is treated as a file.
- Missing parent directories are created automatically with mode `0755`.
- Files are created empty with mode `0644`.
- Pre-validates the full target list: if any target already exists or a duplicate is supplied, the command aborts before creating anything.

The command prints created file and directory counts.

### `tin open <target>...`

Opens one or more targets with the platform's default handler.

Behavior:

- macOS uses `open`. Linux uses `xdg-open` (install hint shown if missing).
- Arguments containing `://` are treated as URLs and passed through without local checks.
- Other targets must point to an existing local path.
- Returns silently on success and prints `error: ...` on failure.

### `tin merge <file|glob>... [-o output.pdf]`

Merges two or more PDF files into a single PDF using ghostscript. Accepts file paths and shell-style glob patterns.

Behavior:

- Requires at least two `.pdf` inputs after expansion.
- Default output is `merged.pdf`.
- Refuses to overwrite an existing output file.
- Refuses non-`.pdf` inputs or outputs.

The command prints input file count, combined input size, output size, and output path.

### `tin split <file.pdf> [-o dir]`

Splits a PDF file into one PDF per page using ghostscript.

Behavior:

- Requires a `.pdf` input.
- Default output directory is `<name>-pages/`.
- Creates the output directory if it does not exist.
- Refuses to write into an existing non-empty output directory.
- Page files are named `page-001.pdf`, `page-002.pdf`, ...

The command prints page count and output directory.

### `tin rename <file|glob>... [-s style] [--dry-run] [--yes]`

Renames files so each basename follows a chosen case style. Accepts file paths and shell-style glob patterns. Normalization is Unicode-aware: letters and digits from any script (including Hangul, CJK, Cyrillic, accented Latin) are preserved.

Supported styles for `-s` / `--style`:

| Style       | Example for `My Photo Album.JPG` |
|-------------|----------------------------------|
| `kebab`     | `my-photo-album.jpg` (default)   |
| `snake`     | `my_photo_album.jpg`             |
| `camel`     | `myPhotoAlbum.jpg`               |
| `pascal`    | `MyPhotoAlbum.jpg`               |
| `screaming` | `MY_PHOTO_ALBUM.jpg`             |

Tokenization splits on non-letter/digit characters and on case transitions (`camelCase` → `camel` + `Case`; `HTMLParser` → `HTML` + `Parser`).

Behavior:

- Splits the basename stem into tokens, then re-joins them in the chosen style.
- Combining marks (e.g. macOS NFD-encoded accented characters) are treated as part of the preceding letter, not as separators. `Información.TXT` stays one word.
- Preserves any leading dots so dotfiles like `.gitignore` keep their prefix.
- Lowercases the trailing extension regardless of style.
- Leaves already-normalized files in place and reports them as skipped.
- Confirms by default.
- `--dry-run` prints the rename plan only.
- `--yes` skips confirmation.
- Avoids collisions with existing files and with other planned destinations using `name.1.ext` style suffixes.
- Treats case-only renames on case-insensitive filesystems as a single rename rather than a collision.
- Rejects names whose stem contains no letters or digits.

The command prints planned rename count, skipped count, and final moved count.

### `tin replace <old> <new> [dir] [-e ext] [-n glob] [--dry-run] [--yes]`

Recursively replaces every occurrence of a literal string `<old>` with `<new>` across files under `dir` (current directory by default). This is the safe equivalent of `find . -exec sed -i ...` without the BSD/GNU sed differences and escaping pitfalls.

Behavior:

- Matching is literal, not regex; no escaping is needed for the search or replacement text.
- Walks the directory recursively. `-e` / `--ext` limits to comma-separated extensions (e.g. `go,md`); `-n` / `--name` limits to files whose basename matches a glob.
- Binary files (detected by a NUL byte in the first 8 KB) are skipped so they are never corrupted.
- The `.git` directory is always skipped.
- File permissions are preserved; each file is rewritten via an atomic temp-file write and rename in the same directory.
- Confirms by default, showing the per-file occurrence counts first.
- `--dry-run` prints the plan only. `--yes` skips confirmation.

The command prints the affected file count and total occurrence count, both in the plan and after applying.

### `tin resize <file|glob>... -w width [-H height] [-o output] [--in-place]`

Resizes one or more supported image files using ImageMagick. Accepts file paths and shell-style glob patterns. `-o` is only allowed with a single input. `--in-place` replaces each input file with its resized output via an atomic temp-file write and rename in the same directory; it cannot be combined with `-o`.

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

### `tin git undo [--hard] [--yes]`

Undoes the latest commit by resetting the current branch to its first parent.

Behavior:

- Defaults to `git reset --soft HEAD^`, leaving the undone commit's changes staged.
- `--hard` uses `git reset --hard HEAD^`, discarding the undone commit plus all staged and unstaged tracked changes.
- Untracked files are never removed, including with `--hard`.
- Shows the commit hash, subject, repository root, and reset mode before acting.
- `--hard` confirms by default. `--yes` skips that confirmation.
- Refuses to undo a root commit because it has no parent.

The command prints the commit and reset mode before and after applying the reset.

### `tin git obliterate <path> [--dry-run] [--force] [--yes]`

Permanently removes an exact file or directory path from all local Git history using `git-filter-repo`.

Behavior:

- Requires both `git` and `git-filter-repo`.
- Requires a clean working tree, including no untracked files.
- Accepts paths relative to the current directory or absolute paths, then normalizes them to a repository-root-relative path.
- Refuses paths outside the repository, the repository root itself, and `.git` metadata.
- Verifies that the path appears in history before running the rewrite.
- Runs `git filter-repo --path <path> --invert-paths`.
- Preserves `git-filter-repo`'s fresh-clone safety check by default. `--force` explicitly forwards `--force` to override it.
- Confirms by default. `--dry-run` prints the plan only and `--yes` skips confirmation.
- Path filtering does not follow renames. Every historical path used by a renamed file must be obliterated separately.
- Rewrites commit hashes across local branches and tags. `git-filter-repo` may remove the `origin` remote to prevent accidental pushes of rewritten history.

Publishing the rewrite requires coordinating with other repository users and force-pushing the rewritten refs. Removing a committed secret from history does not revoke it; exposed credentials must also be rotated.

### `tin size <path|glob>...`

Shows a human-readable size for one or more files or directories. Accepts paths and shell-style glob patterns.

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

### `tin zip <path|glob>... [-o archive.zip]`

Creates a `.zip` archive using Go's standard library. Accepts paths and shell-style glob patterns.

Behavior:

- Accepts files and directories.
- Recursively archives directory contents.
- Default output is `<single-input-name>.zip` for one input or `archive.zip` for multiple inputs.
- Refuses to overwrite an existing output archive.
- Skips symlinks and non-regular files (devices, sockets, FIFOs) instead of failing; they are left out of the archive and counted as skipped.

The command prints file count, uncompressed byte count, skipped count (when any), and output path.

### `tin unzip <archive.zip|glob>... [-o dir]`

Extracts one or more `.zip` archives using Go's standard library. Accepts paths and shell-style glob patterns.

Behavior:

- Each archive is extracted into its own directory named after the archive without `.zip`, next to the archive.
- `-o` sets the parent directory those per-archive directories are created in, so `tin unzip '*.zip' -o out` produces `out/<archive>/` for each match.
- Prints one result box per archive.
- A failing archive in a batch is reported as a warning on stderr and the remaining archives are still extracted; the command exits non-zero with a summary of how many succeeded. A single archive reports its own error verbatim.
- Directories are created writable rather than with their stored mode; zips written on Windows record the DOS read-only attribute as mode `0555`, which would make every entry inside them fail with `permission denied`. File modes are preserved.
- Entry names that are not valid UTF-8 are decoded as CP437, the encoding the zip spec assigns to names stored without the UTF-8 flag. Writing such bytes to the filesystem raw fails on macOS with `illegal byte sequence`.
- Refuses non-zip archive extensions.
- Refuses unsafe archive paths such as `../file`.
- Refuses to overwrite existing destination files.
- Refuses unsupported entries such as symlinks.

The command prints file count, extracted byte count, and output directory for each archive.

## Deferred commands

The following broader toolbox ideas are deferred:

- diff
- disks
- mount
- unmount
- eject
- kill
- usage

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
