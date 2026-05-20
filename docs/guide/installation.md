# Installation

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

The script installs to `~/.qdoc/bin` (no sudo) and adds it to your shell config. Options:

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash -s -- --version 0.1.2  # specific version
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash -s -- --no-modify-path  # don't touch shell config
```

## Manual Download

Binaries for all platforms are on the [GitHub Releases](https://github.com/ibrhr/qdoc/releases/latest) page.

| Platform | Archive |
|---|---|
| Linux amd64 | `qdoc_linux_amd64.tar.gz` |
| Linux arm64 | `qdoc_linux_arm64.tar.gz` |
| macOS Intel | `qdoc_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `qdoc_darwin_arm64.tar.gz` |
| Windows amd64 | `qdoc_windows_amd64.zip` |

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz
mkdir -p ~/.local/bin && mv qdoc ~/.local/bin/
```

## From Source

Requires Go 1.26+:

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
mkdir -p ~/.local/bin && mv qdoc ~/.local/bin/
```

## Verify

```bash
qdoc --version
```
