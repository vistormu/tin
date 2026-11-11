<p align="center">
  <img src="docs/tin.webp" width="200" alt="tin logo" />
</p>

<h1 align="center">tin</h1>

<p align="center">
  <strong>a bundle of (t)erm(in)al tools</strong>
</p>

---

`tin` is a small go cli that bundles the file, media, and inspection commands i reach for most often into one binary:

```sh
tin <command> [args] [flags]
```

the goal is to keep common terminal tasks fast and predictable, with sensible defaults and friendly output.

## install

with homebrew (macos and linux):

```sh
brew install --cask vistormu/tap/tin
```

from source: `just install` builds it into `~/.local/bin/tin`.

## release

push a version tag and the release workflow does the rest (github release +
homebrew cask, via goreleaser):

```sh
git tag v0.2.0 && git push origin v0.2.0
```

`just snapshot` is the local dry run.

## commands

### media

| command        | purpose                                   |
| -------------- | ----------------------------------------- |
| `tin compress` | compress pdfs, images, and videos         |
| `tin convert`  | convert between image, pdf, svg, video, and audio formats |
| `tin resize`   | resize images                             |
| `tin merge`    | merge multiple pdfs into one              |
| `tin split`    | split a pdf into one file per page        |

### archives

| command     | purpose                |
| ----------- | ---------------------- |
| `tin zip`   | create a zip archive   |
| `tin unzip` | extract zip archives   |

### inspect

| command     | purpose                                      |
| ----------- | -------------------------------------------- |
| `tin find`  | find files by name, extension, and type      |
| `tin count` | count files, lines, and bytes                |
| `tin size`  | print human-readable file or directory sizes |

### organize

| command       | purpose                                                       |
| ------------- | ------------------------------------------------------------- |
| `tin new`     | create files or directories with parent dirs auto-created     |
| `tin rename`  | normalize filenames to a chosen case style                    |
| `tin replace` | replace a string in files recursively                        |
| `tin flatten` | move nested files into a directory root                       |
| `tin dedup`   | find duplicate files and trash extras                         |
| `tin trash`   | move files or folders to the system trash                     |

### git

| command              | purpose                                                |
| -------------------- | ------------------------------------------------------ |
| `tin git undo`       | undo the latest commit, preserving changes by default |
| `tin git obliterate` | remove a path from the entire git history              |

### system

| command    | purpose                                                  |
| ---------- | -------------------------------------------------------- |
| `tin open` | open a file, folder, or url with the default handler     |
| `tin ip`   | show local interface and public ip addresses             |

run `tin <command> --help` for full flags and examples.

## a taste

```sh
tin compress report.pdf -i high
tin convert photo.heic -t jpg
tin convert diagram.pdf -t svg
tin convert scan.tiff -t png
tin resize "*.jpg" -w 1600
tin merge chapter-*.pdf -o book.pdf
tin rename "*.jpg" --style kebab --dry-run
tin replace example.com/old example.com/new . -e go
tin dedup ./downloads
tin git undo
tin git obliterate .env --dry-run
tin ip
```

commands that move or remove files show the plan first, confirm by default, and support `--dry-run` and `--yes`.

`tin git obliterate` requires
[`git-filter-repo`](https://github.com/newren/git-filter-repo). history rewrites
change commit hashes and may remove the `origin` remote; coordinate before
force-pushing rewritten history.
