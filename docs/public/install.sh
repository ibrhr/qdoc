#!/bin/sh
set -eu

REPO="ibrhr/qdoc"
BASE="https://github.com/${REPO}/releases/latest/download"

case "$(uname -s)" in
  Linux)  os="linux"   ;;
  Darwin) os="darwin"  ;;
  *)      echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

url="${BASE}/qdoc_${os}_${arch}.tar.gz"
dest="/usr/local/bin/qdoc"

echo "Downloading qdoc ${os}/${arch}..."
curl -fsSL "$url" | tar xz

if [ ! -f qdoc ]; then
  echo "Error: binary not found in archive" >&2
  exit 1
fi

chmod +x qdoc
if [ -w /usr/local/bin ]; then
  mv qdoc "$dest"
else
  echo "Need sudo to install to /usr/local/bin"
  sudo mv qdoc "$dest"
fi

echo "qdoc installed to $dest"
qdoc --version
