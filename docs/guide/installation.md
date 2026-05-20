# Installation

## Binary Download

Download prebuilt binaries from [GitHub Releases](https://github.com/ibrhr/qdoc/releases/latest).

### Linux

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz
sudo mv qdoc /usr/local/bin/
```

For ARM64 (Raspberry Pi, AWS Graviton):

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_arm64.tar.gz
tar xzf qdoc_linux_arm64.tar.gz
sudo mv qdoc /usr/local/bin/
```

### macOS

Intel:

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_darwin_amd64.tar.gz
tar xzf qdoc_darwin_amd64.tar.gz
sudo mv qdoc /usr/local/bin/
```

Apple Silicon:

```bash
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_darwin_arm64.tar.gz
tar xzf qdoc_darwin_arm64.tar.gz
sudo mv qdoc /usr/local/bin/
```

### Windows

```powershell
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_windows_amd64.zip
tar xzf qdoc_windows_amd64.zip
move qdoc.exe C:\Windows\System32\
```

## From Source

If you have Go 1.26+:

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
