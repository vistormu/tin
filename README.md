<p align="center">
  <img src="docs/tin.webp" width="200" alt="tin logo" />
</p>

<h1 align="center">tin</h1>

<p align="center">
  <strong>a compendium of (t)erminal b(in)aries under one command</strong>
</p>

---

`tin` is a small go cli for high-friction local file tasks: media compression and conversion, pdf helpers, archives, directory cleanup, safe trashing, file counting, and size inspection.

the public binary is `tin`. the local go module is also `tin`.

## status

`tin` currently supports macos and linux only. unsupported platforms, including windows, fail with a clear error.

the cli uses runtime platform detection, not go build tags. commands that move or remove user data print the exact plan first, confirm by default, and support `--dry-run` where useful.

## install

build and install to `~/.local/bin/tin`:

```sh
just install
```

override the install prefix if needed:

```sh
prefix=/usr/local just install
```

build a local binary:

```sh
just build-local
```

build release binaries for macos and linux on amd64 and arm64:

```sh
just build
```

run tests:

```sh
just test
```

## runtime dependencies

most commands are pure go. media and pdf commands call external tools only when needed:

| tool | used by |
|------|---------|
| `ghostscript` / `gs` | `compress` for pdfs, `merge`, `split` |
| `magick` | `compress` for jpg/jpeg/png, `convert`, `resize` |
| `cwebp` | `compress` for webp |

when a required tool is missing, `tin` detects package managers from `path` and prints an install hint.

## input globs

commands that accept multiple inputs support quoted shell-style globs:

```sh
tin size "docs/*.md"
tin compress "*.pdf" -i high
tin zip "photos/*.jpg" -o photos.zip
```

literal paths must exist. empty glob matches are errors. duplicate expanded paths are collapsed.

## commands

### `tin compress <file|glob>... [-i intensity] [-o output] [--in-place]`

compresses pdfs and images.

supported inputs:

- `.pdf`
- `.jpg`
- `.jpeg`
- `.png`
- `.webp`

examples:

```sh
tin compress report.pdf
tin compress photo.jpg -i high
tin compress "*.png" -i medium
tin compress photo.jpg --in-place
```

default output is `<name>-compressed.<ext>`. `-o` is allowed only with one input. `--in-place` replaces each input through a temporary file in the same directory and cannot be combined with `-o`.

compression intensities:

| intensity | pdf preset | image quality |
|-----------|------------|---------------|
| `low` | `prepress` | 90 |
| `medium` | `printer` | 80 |
| `high` | `ebook` | 65 |
| `veryhigh` | `screen` | 45 |

default intensity is `medium`. run `tin compress intensities` for the same list in the terminal.

errors on unsupported file types, missing tools, invalid intensity, multiple inputs with `-o`, or `--in-place` with `-o`.

### `tin convert <file|glob>... [-t target] [-o output]`

converts images with imagemagick.

supported input and target types:

- `.jpg`
- `.jpeg`
- `.png`
- `.webp`

examples:

```sh
tin convert photo.png -t webp
tin convert photo.png -o photo.jpg
tin convert "*.png" -t webp
```

at least one of `-t` or `-o` is required. if both are provided, the output extension must match the target type. default output is `<name>.<target>`.

errors on unsupported formats, no-op conversions, existing output files, multiple inputs with `-o`, missing imagemagick, or mismatched `-t` / `-o`.

### `tin count <path|glob>... [-e ext]`

counts files, lines, and bytes across files or directories.

examples:

```sh
tin count .
tin count ./cmd ./internal -e go
tin count "*.md"
```

directories are walked recursively. symlinks and non-regular files are skipped. line counts follow `wc -l` semantics, so a file without a trailing newline contributes zero lines.

errors on missing paths or glob patterns with no matches.

### `tin dedup <dir> [--dry-run] [--yes]`

finds duplicate files under a directory by content hash and moves duplicate copies to the system trash.

examples:

```sh
tin dedup ./downloads --dry-run
tin dedup ./photos --yes
```

files are grouped by size first, then sha-256 hash. the alphabetically-first path in each duplicate group is kept. the rest are trashed through the same backend as `tin trash`.

errors when the target is not a directory or when trashing fails.

### `tin find [dir] [-e ext] [-n glob] [--type f|d|all]`

finds paths under a directory.

examples:

```sh
tin find . -e go
tin find ./src -e go,md
tin find . -n "*test*"
tin find . --type d -n "*tmp*"
```

defaults to the current directory and regular files. filters combine with and semantics.

errors on invalid type filters, invalid name glob patterns, missing roots, or filesystem walk errors.

### `tin flatten <dir> [--dry-run] [--yes]`

moves nested files into a directory root and removes empty subdirectories.

examples:

```sh
tin flatten ./downloads --dry-run
tin flatten ./photos --yes
```

root-level files stay in place. nested regular files and symlinks are moved to the root. collisions use extension-preserving names like `file.1.txt`.

errors on non-directory roots, unsupported file types, or failed moves/removals.

### `tin merge <file|glob>... [-o output.pdf]`

merges two or more pdfs into one pdf with ghostscript.

examples:

```sh
tin merge a.pdf b.pdf
tin merge "*.pdf" -o combined.pdf
```

default output is `merged.pdf`.

errors on fewer than two inputs, non-pdf inputs or outputs, existing output files, missing ghostscript, or ghostscript failures.

### `tin rename <file|glob>... [-s style] [--dry-run] [--yes]`

normalizes filenames to a chosen case style.

examples:

```sh
tin rename "my photo.jpg"
tin rename "*.jpg" --dry-run
tin rename report.pdf --style snake
tin rename "my file.txt" --style camel --yes
```

supported styles:

| style | example for `my photo album.jpg` |
|-------|----------------------------------|
| `kebab` | `my-photo-album.jpg` |
| `snake` | `my_photo_album.jpg` |
| `camel` | `myphotoalbum.jpg` |
| `pascal` | `myphotoalbum.jpg` |
| `screaming` | `my_photo_album.jpg` |

default style is `kebab`. tokenization is unicode-aware, preserves leading dots, lowercases extensions, and avoids collisions with `name.1.ext` suffixes.

errors on unsupported styles, missing paths, or names whose stem contains no letters or digits.

### `tin resize <file|glob>... -w width [-h height] [-o output] [--in-place]`

resizes images with imagemagick.

supported inputs:

- `.jpg`
- `.jpeg`
- `.png`
- `.webp`

examples:

```sh
tin resize photo.jpg -w 1200
tin resize photo.png -w 1200 -h 800 -o photo-small.png
tin resize "*.jpg" -w 1600
tin resize photo.jpg -w 1200 --in-place
```

default output is `<name>-resized.<ext>`. height is optional; when omitted, imagemagick preserves the aspect ratio. `-o` is allowed only with one input. `--in-place` cannot be combined with `-o`.

errors on invalid dimensions, unsupported file types, multiple inputs with `-o`, missing imagemagick, or imagemagick failures.

### `tin size <path|glob>...`

prints human-readable sizes for files or directories.

examples:

```sh
tin size readme.md
tin size docs internal
tin size "docs/*.md"
```

directory sizes are computed with pure go recursive walking, not `du`.

errors on missing paths, empty glob matches, or filesystem walk errors.

### `tin split <file.pdf> [-o dir]`

splits a pdf into one pdf per page with ghostscript.

examples:

```sh
tin split report.pdf
tin split report.pdf -o report-pages
```

default output directory is `<name>-pages`. page files are named `page-001.pdf`, `page-002.pdf`, and so on.

errors on non-pdf inputs, existing non-empty output directories, output paths that are files, missing ghostscript, or ghostscript failures.

### `tin trash <file|glob>... [--dry-run] [--yes]`

moves files to the system trash.

examples:

```sh
tin trash old.txt
tin trash "tmp/*.log" --dry-run
tin trash old-downloads --yes
```

macos moves files into `~/.trash`. linux uses the xdg trash location at `$xdg_data_home/trash`, falling back to `~/.local/share/trash`, and writes `.trashinfo` metadata for restore support.

errors on unsupported platforms, missing paths, empty glob matches, or failed trash moves.

### `tin zip <path|glob>... [-o archive.zip]`

creates zip archives with go's standard library.

examples:

```sh
tin zip docs
tin zip docs readme.md -o project.zip
tin zip "photos/*.jpg" -o photos.zip
```

default output is `<single-input-name>.zip` for one input and `archive.zip` for multiple inputs.

errors on existing output archives, unsupported file types such as symlinks, duplicate archive paths, missing paths, or empty glob matches.

### `tin unzip <archive.zip> [-o dir]`

extracts zip archives with go's standard library.

examples:

```sh
tin unzip archive.zip
tin unzip archive.zip -o extracted
```

default output directory is the archive name without `.zip`.

errors on non-zip archive extensions, unsafe archive paths like `../file`, existing destination files, unsupported entries such as symlinks, or invalid zip files.

## terminal output

user-facing output is rendered through a shared ui layer. color behavior follows common terminal conventions:

- `no_color` disables color.
- `clicolor=0` disables color unless forced.
- `clicolor_force=1` forces color.
- otherwise, color support is detected from the output writer.

top-level errors are printed as:

```text
error: <message>
```

runtime and validation errors do not print full usage blocks.

