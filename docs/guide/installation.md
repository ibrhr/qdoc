# Installation

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | sh
```

The script detects your OS and architecture, downloads the right binary, and installs to `/usr/local/bin`.

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
# Linux / macOS
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz
sudo mv qdoc /usr/local/bin/
```

## From Source

Requires Go 1.26+:

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
sudo mv qdoc /usr/local/bin/
```

## Verify

```bash
qdoc --version
```
