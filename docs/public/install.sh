#!/usr/bin/env bash
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

${MUTED}Usage:${NC} curl -fsSL https://qdoc.ibrhr.dev/install.sh | sh

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
case "$(uname -s)" in
    Linux)  os="linux"   ;;
    Darwin) os="darwin"  ;;
    *)      echo -e "${RED}Unsupported OS: $(uname -s)${NC}" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) echo -e "${RED}Unsupported arch: $arch${NC}" >&2; exit 1 ;;
esac

combo="${os}_${arch}"
ext=".tar.gz"

# ── version resolution ───────────────────────────────────────────────

if [ -z "$requested_version" ]; then
    tag=$(curl -sf https://api.github.com/repos/$REPO/releases/latest | grep -o '"tag_name": *"v[^"]*"' | grep -o 'v[^"]*')
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
    curl -# -fL -o "$tmp/$APP$ext" "$url"
else
    curl -sfL -o "$tmp/$APP$ext" "$url"
fi

# ── extract & install ────────────────────────────────────────────────

tar xzf "$tmp/$APP$ext" -C "$tmp"
mkdir -p "$INSTALL_DIR"

if [ -f "$tmp/$APP" ]; then
    mv "$tmp/$APP" "$INSTALL_DIR/$APP"
elif [ -f "$tmp/$APP.exe" ]; then
    mv "$tmp/$APP.exe" "$INSTALL_DIR/$APP.exe"
else
    echo -e "${RED}Binary not found in archive${NC}" >&2
    exit 1
fi

chmod +x "$INSTALL_DIR/$APP"

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
            local f="$HOME/.bashrc"
            [ ! -f "$f" ] && f="$HOME/.profile"
            [ ! -f "$f" ] && f="$HOME/.bash_profile"
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
echo -e "  ${GREEN}█▄▄▄▄▄▄▄ █▄▄▄▄▄█ █▄▄▄█${NC}  ${MUTED}installed to${NC} $INSTALL_DIR/$APP"
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
