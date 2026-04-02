#!/bin/bash
set -e

PREFIX="${PREFIX:-/usr/local/bin}"
BUILD_DIR="build"

echo "Installing ai-terminal to $PREFIX..."

# Check if build exists
if [ ! -f "$BUILD_DIR/ai-terminal" ]; then
    echo "Build not found. Running 'make build' first..."
    make build
fi

# Copy main binary
cp "$BUILD_DIR/ai-terminal" "$PREFIX/ai-terminal"
chmod +x "$PREFIX/ai-terminal"
echo "  ✓ ai-terminal"

# Create symlinks for wrapper scripts
WRAPPERS=("@" "@!" "@#" "@!#")

for name in "${WRAPPERS[@]}"; do
    script="scripts/wrappers/$name"
    if [ -f "$script" ]; then
        cp "$script" "$PREFIX/$name"
        chmod +x "$PREFIX/$name"
        echo "  ✓ $name"
    fi
done

echo ""
echo "Installation complete. Run 'ai-terminal --help' to get started."
