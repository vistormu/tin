default:
    @just --list

# release dry run: all platforms + the cask, into dist/, nothing published.
# real releases: push a v* tag (see .goreleaser.yaml)
snapshot:
    HOMEBREW_TAP_TOKEN=unused goreleaser release --snapshot --clean

build-local:
    mkdir -p "${DIST:-dist}"
    GOWORK="${GOWORK:-off}" "${GO:-go}" build -buildvcs=false -o "${DIST:-dist}/${BINARY:-tin}" .

test:
    GOWORK="${GOWORK:-off}" "${GO:-go}" test ./...

install:
    mkdir -p "${PREFIX:-$HOME/.local}/bin"
    GOWORK="${GOWORK:-off}" "${GO:-go}" build -buildvcs=false -o "${PREFIX:-$HOME/.local}/bin/${BINARY:-tin}" .

clean:
    rm -rf "${DIST:-dist}"
