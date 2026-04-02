#!/bin/bash
set -e

REPO="shezw/ai-terminal"
PREFIX="${PREFIX:-/usr/local/bin}"
TMPDIR="${TMPDIR:-/tmp}"

echo "=== ai-terminal quick installer ==="
echo ""

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin|linux) ;;
    *)           echo "Unsupported OS: $OS"; exit 1 ;;
esac

BINARY="ai-terminal-${OS}-${ARCH}"
echo "Platform: ${OS}/${ARCH}"

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -sI "https://github.com/${REPO}/releases/latest" | grep -i '^location:' | sed 's|.*/||' | tr -d '\r')
if [ -z "$LATEST" ]; then
    echo "Failed to determine latest release."
    exit 1
fi
echo "Latest release: $LATEST"

# Download binary
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}"
DEST="${TMPDIR}/ai-terminal"

echo "Downloading ${URL}..."
curl -fSL "$URL" -o "$DEST"
chmod +x "$DEST"

# Install
echo "Installing to ${PREFIX}..."
if [ -w "$PREFIX" ]; then
    mv "$DEST" "${PREFIX}/ai-terminal"
else
    sudo mv "$DEST" "${PREFIX}/ai-terminal"
fi
echo "  ✓ ai-terminal"

# Install wrapper scripts
WRAPPERS=("@" "@!" "@#" "@!#")
WRAPPER_CMDS=(
    '#!/bin/bash\nexec ai-terminal "$@"'
    '#!/bin/bash\nexec ai-terminal exec "$@"'
    '#!/bin/bash\nexec ai-terminal show --think "$@"'
    '#!/bin/bash\nexec ai-terminal exec --think "$@"'
)

for i in "${!WRAPPERS[@]}"; do
    name="${WRAPPERS[$i]}"
    cmd="${WRAPPER_CMDS[$i]}"
    wpath="${PREFIX}/${name}"
    if [ -w "$PREFIX" ]; then
        printf '%b\n' "$cmd" > "$wpath"
        chmod +x "$wpath"
    else
        printf '%b\n' "$cmd" | sudo tee "$wpath" > /dev/null
        sudo chmod +x "$wpath"
    fi
    echo "  ✓ $name"
done

echo ""
echo "Installation complete!"
echo "Run 'ai-terminal --help' to get started."
echo ""
echo "To use a local model:"
echo "  ai-terminal model install"
