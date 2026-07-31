#!/usr/bin/env bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <stage_number (0-4)>"
    echo "  0: 0-base        (Initial scaffold with QUEST tasks)"
    echo "  1: 1-lexer       (Lexer QUEST solved: digits, identifiers, var/func tokens)"
    echo "  2: 2-parser      (Parser QUEST solved: if/else, var, func, call, return)"
    echo "  3: 3-interpreter (Interpreter QUEST solved: branches, env, function frames)"
    echo "  4: 4-compiler    (Compiler QUEST solved: Go transpilation complete)"
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
echo "Successfully restored worktree to $TAG!"
