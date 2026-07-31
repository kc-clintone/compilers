#!/usr/bin/env bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <stage_number (0-4)>"
    echo "  0: 0-base        (Nuru workshop starter — all QUEST tasks ready)"
    echo "  1: 1-lexer       (Nuru lexer QUEST complete — digits, identifiers, and var/func tokens)"
    echo "  2: 2-parser      (Nuru parser QUEST complete — branches, variables, functions, calls, and returns)"
    echo "  3: 3-interpreter (Nuru interpreter QUEST complete — environments, branches, and function call frames)"
    echo "  4: 4-compiler    (Nuru transpiler QUEST complete — Go code generation)"
    exit 1
fi

case "$1" in
    0) TAG="0-base" ;;
    1) TAG="1-lexer" ;;
    2) TAG="2-parser" ;;
    3) TAG="3-interpreter" ;;
    4) TAG="4-compiler" ;;
    *)
        echo "Error: Invalid stage number '$1'. Must be 0, 1, 2, 3, or 4."
        exit 1
        ;;
esac

echo "Resetting repository to checkpoint tag: $TAG..."
git reset --hard "$TAG"
git clean -fd
echo "Successfully restored repository checkpoint $TAG!"
