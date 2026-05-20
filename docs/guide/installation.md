# Installation

## Go Install

Requires [Go](https://go.dev/dl/) 1.26+:

```bash
go install github.com/ibrhr/qdoc@latest
```

## Binary Download

Download prebuilt binaries from the [GitHub Releases](https://github.com/ibrhr/qdoc/releases) page:

| Platform | Architecture |
|---|---|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64 |

```bash
# Example: Linux amd64
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz
sudo mv qdoc /usr/local/bin/
```

## From Source

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build -ldflags "-X main.version=$(git describe --tags)" -o qdoc .
sudo mv qdoc /usr/local/bin/
```

## Verify

```bash
qdoc --version
# qdoc 0.1.0 (abc1234)
```
