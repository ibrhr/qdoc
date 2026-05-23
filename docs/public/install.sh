#!/usr/bin/env bash
# Re-exec with bash if invoked via sh (e.g. curl | sh on systems where /bin/sh is dash)
if [ -z "${BASH_VERSION:-}" ]; then
    if command -v bash >/dev/null 2>&1; then
        exec bash -c "$(curl -fsSL https://qdoc.ibrhr.dev/install.sh)" bash "$@"
    else
        echo "Error: bash is required for this installer." >&2
        echo "Install bash or download the binary manually:" >&2
        echo "  https://github.com/ibrhr/qdoc/releases" >&2
        exit 1
    fi
fi
set -euo pipefail

APP=qdoc
REPO=ibrhr/qdoc
INSTALL_DIR="$HOME/.$APP/bin"

MUTED='\033[0;2m'
ORANGE='\033[38;5;214m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

usage() {
    cat <<EOF
${BOLD}qdoc installer${NC}

${MUTED}Usage:${NC} curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash

${MUTED}Options:${NC}
    -h, --help               Show this message
    -v, --version <version>  Install a specific version (e.g. 0.1.2)
    --no-modify-path         Don't touch shell config files
EOF
}

requested_version=""
no_modify_path=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        -v|--version)
            requested_version="${2#v}"
            shift 2
            ;;
        --no-modify-path)
            no_modify_path=true
            shift
            ;;
        *)  shift ;;
    esac
done

# ── platform detection ──────────────────────────────────────────────

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
    linux)
        is_windows=false
        ext=".tar.gz"
        ;;
    darwin)
        is_windows=false
        ext=".tar.gz"
        ;;
    mingw*|msys*|cygwin*)
        is_windows=true
        ext=".zip"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $(uname -s)${NC}" >&2
        echo -e "${MUTED}See https://github.com/ibrhr/qdoc/releases for manual download${NC}" >&2
        exit 1
        ;;
esac

arch=$(uname -m)
case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) echo -e "${RED}Unsupported arch: $arch${NC}" >&2; exit 1 ;;
esac

combo="${os}_${arch}"
# Normalize OS for the download URL (mingw64_nt → windows, msys_nt → windows)
download_os="$os"
if $is_windows; then
    download_os="windows"
    combo="windows_${arch}"
fi

# ── version resolution ───────────────────────────────────────────────

if [ -z "$requested_version" ]; then
    tag=$(curl -s -o /dev/null -w '%{url_effective}' -L "https://github.com/$REPO/releases/latest" | grep -oE 'v[^/]*$' || echo "")
    if [ -z "$tag" ]; then
        echo -e "${RED}Failed to resolve latest version.${NC}" >&2
        echo -e "${RED}Specify a version manually: curl ... | bash -s -- --version X.Y.Z${NC}" >&2
        exit 1
    fi
    specific_version="${tag#v}"
else
    tag="v$requested_version"
    http_code=$(curl -sI -o /dev/null -w "%{http_code}" "https://github.com/$REPO/releases/tag/$tag")
    if [ "$http_code" = "404" ]; then
        echo -e "${RED}Release $tag not found${NC}" >&2
        exit 1
    fi
    specific_version="$requested_version"
fi

# ── version check ────────────────────────────────────────────────────

if $is_windows; then
    app_bin="$APP.exe"
else
    app_bin="$APP"
fi

if command -v "$APP" >/dev/null 2>&1; then
    installed=$("$APP" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "")
    if [ "$installed" = "$specific_version" ]; then
        echo -e "${GREEN}$APP $specific_version already installed${NC}"
        exit 0
    fi
    echo -e "${MUTED}Upgrading $APP $installed → $specific_version${NC}"
fi

# ── download ─────────────────────────────────────────────────────────

url="https://github.com/$REPO/releases/download/$tag/qdoc_${combo}${ext}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo -e "${MUTED}Downloading ${NC}$APP ${BOLD}$specific_version${NC} ${MUTED}($combo)${NC}"

if [ -t 2 ]; then
    curl -# -fL -o "$tmp/$APP$ext" "$url" || {
        echo -e "${RED}Download failed${NC}" >&2
        exit 1
    }
else
    curl -sfL -o "$tmp/$APP$ext" "$url" || {
        echo -e "${RED}Download failed${NC}" >&2
        exit 1
    }
fi

# ── checksum verification ────────────────────────────────────────────

checksum_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"
curl -sfL -o "$tmp/checksums.txt" "$checksum_url"
archive_name=$(basename "$url")

if $is_windows; then
    # sha256sum is not available on Git Bash by default; fall back to openssl
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$tmp/$APP$ext" | awk '{print $1}')
    elif command -v openssl >/dev/null 2>&1; then
        actual=$(openssl dgst -sha256 "$tmp/$APP$ext" | awk '{print $NF}')
    elif command -v certutil >/dev/null 2>&1; then
        actual=$(certutil -hashfile "$tmp/$APP$ext" SHA256 | tail -n +2 | head -n 1 | tr -d '[:space:]')
    else
        echo -e "${RED}No sha256 tool found${NC}" >&2
        exit 1
    fi
else
    actual=$(sha256sum "$tmp/$APP$ext" | awk '{print $1}')
fi

expected=$(grep -F "$archive_name" "$tmp/checksums.txt" | awk '{print $1}')
if [ "$expected" != "$actual" ] || [ -z "$expected" ]; then
    echo -e "${RED}Checksum verification failed for $archive_name${NC}" >&2
    rm -rf "$tmp"
    exit 1
fi
echo -e "${MUTED}Checksum verified${NC}"

# ── extract & install ────────────────────────────────────────────────

if $is_windows; then
    if command -v unzip >/dev/null 2>&1; then
        unzip -qo "$tmp/$APP$ext" -d "$tmp"
    elif command -v tar >/dev/null 2>&1; then
        # Windows 10 build 1803+ has built-in tar
        tar -xf "$tmp/$APP$ext" -C "$tmp"
    else
        echo -e "${RED}No extraction tool found (need unzip or tar)${NC}" >&2
        exit 1
    fi
else
    tar xzf "$tmp/$APP$ext" -C "$tmp"
fi

mkdir -p "$INSTALL_DIR"

if [ -f "$tmp/$app_bin" ]; then
    mv "$tmp/$app_bin" "$INSTALL_DIR/$app_bin"
else
    echo -e "${RED}Binary not found in archive${NC}" >&2
    exit 1
fi

if ! $is_windows; then
    chmod +x "$INSTALL_DIR/$app_bin"
fi

installed_ver=$("$INSTALL_DIR/$app_bin" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "")
if [ "$installed_ver" != "$specific_version" ]; then
    echo -e "${RED}Version mismatch: expected $specific_version, got ${installed_ver:-none}${NC}" >&2
    rm -f "$INSTALL_DIR/$app_bin"
    exit 1
fi

# ── PATH setup ───────────────────────────────────────────────────────

xdg_config="${XDG_CONFIG_HOME:-$HOME/.config}"
current_shell=$(basename "${SHELL:-/bin/sh}")

add_to_path() {
    local file="$1" cmd="$2"
    if grep -Fxq "$cmd" "$file" 2>/dev/null; then
        return
    fi
    echo "" >> "$file"
    echo "# $APP" >> "$file"
    echo "$cmd" >> "$file"
    echo -e "  ${MUTED}Added to${NC} $file"
}

maybe_add_path() {
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        return
    fi

    echo -e "${MUTED}Adding ${NC}$INSTALL_DIR${MUTED} to PATH${NC}"

    case "$current_shell" in
        fish)
            local f="$xdg_config/fish/config.fish"
            mkdir -p "$(dirname "$f")"
            add_to_path "$f" "fish_add_path $INSTALL_DIR"
            ;;
        zsh)
            local f="${ZDOTDIR:-$HOME}/.zshrc"
            add_to_path "$f" "export PATH=\"$INSTALL_DIR:\$PATH\""
            ;;
        bash)
            if $is_windows; then
                local f="$HOME/.bashrc"
            else
                local f="$HOME/.bash_profile"
                [ ! -f "$f" ] && f="$HOME/.bashrc"
                [ ! -f "$f" ] && f="$HOME/.profile"
            fi
            add_to_path "$f" "export PATH=\"$INSTALL_DIR:\$PATH\""
            ;;
        *)
            echo -e "  ${ORANGE}Unknown shell ($current_shell). Add this to your shell config:${NC}"
            echo -e "  ${BOLD}export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
            ;;
    esac
}

if [ "$no_modify_path" != "true" ]; then
    maybe_add_path
fi

# ── GitHub Actions ───────────────────────────────────────────────────

if [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "$INSTALL_DIR" >> "$GITHUB_PATH"
fi

# ── done ─────────────────────────────────────────────────────────────

echo ""
echo -e "  ${GREEN}▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄${NC}"
echo -e "  ${GREEN}█  ▄  █ █  ▄▄ █ █  ▄▄ █${NC}"
echo -e "  ${GREEN}█ █▄█ █ █ █▄█ █ █ █▄▄ ${NC}  ${BOLD}qdoc $specific_version${NC}"
echo -e "  ${GREEN}█▄▄▄▄▄▄▄ █▄▄▄▄▄█ █▄▄▄█${NC}  ${MUTED}installed to${NC} $INSTALL_DIR/$app_bin"
echo ""

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${ORANGE}Restart your shell or run:${NC}"
    echo -e "  ${BOLD}export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
    echo ""
fi

echo -e "${MUTED}Get started:${NC}"
echo -e "  qdoc set key openai sk-your-key"
echo -e "  qdoc go \"how do generics work?\""
echo ""
