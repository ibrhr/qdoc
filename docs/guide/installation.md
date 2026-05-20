# Installation

## Shell script (Linux, macOS)

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

Installs `qdoc` to `~/.qdoc/bin` (no root required). Adds the directory to your shell config automatically (`.bash_profile`, `.zshrc`, or `config.fish`).

Options:

```bash
# Pin a specific version
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash -s -- --version 0.1.2

# Don't modify shell config (add to PATH yourself)
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash -s -- --no-modify-path
```

The script detects your OS and architecture and downloads the correct binary from [GitHub Releases](https://github.com/ibrhr/qdoc/releases/latest).

## npm (all platforms)

```bash
npm install -g qdoc-agent
```

The `postinstall` script downloads the correct platform binary to `node_modules/qdoc-agent/qdoc_bin`. A shim script at `bin/qdoc.js` proxies CLI invocations to the binary. Requires Node.js 18+.

## Manual Download

Pre-built binaries for all platforms:

| Platform | Archive |
|----------|---------|
| Linux x86_64 | `qdoc_linux_amd64.tar.gz` |
| Linux arm64 | `qdoc_linux_arm64.tar.gz` |
| macOS Intel | `qdoc_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `qdoc_darwin_arm64.tar.gz` |
| Windows x86_64 | `qdoc_windows_amd64.zip` |

Download from [GitHub Releases](https://github.com/ibrhr/qdoc/releases/latest):

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz
mkdir -p ~/.local/bin && mv qdoc ~/.local/bin/
```

## From Source

Requires Go 1.26+.

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
mkdir -p ~/.local/bin && mv qdoc ~/.local/bin/
```

## Verify

```bash
qdoc --version
# qdoc 0.1.2 (abc1234)
```

If `qdoc` is not found, you may need to restart your shell or source your config:

```bash
source ~/.bashrc   # or ~/.zshrc, ~/.config/fish/config.fish
```
