#!/bin/bash

# Push staged changes with ai-generated commit message

set -euo pipefail

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo -e "Not in a git repository."
  exit 1
fi

git add -A

if [[ -z "$(git diff --cached --name-only)" ]]; then
  echo -e "Nothing staged to commit."
  exit 0
fi

diff="$(git diff --cached)"
max_chars=50000
if (( ${#diff} > max_chars )); then
  diff="${diff:0:max_chars}
... <diff truncated at ${max_chars} characters>"
fi

prompt="
You are an expert software engineer. 
Draft a concise, helpful Git commit message for the staged changes below.

Rules:
- Prefer Conventional Commits (e.g., feat:, fix:, chore:, docs:, refactor:, test:, build:)
- 1 short subject line (<= 72 chars), then optional body with wrapped bullets
- No code fences, no quotes around the subject, no emojis
- Be specific (mention files/areas and intent), but don't exceed a few lines
- Return ONLY the commit message, nothing else

STAGED DIFF:
${diff}"

if ! command -v claude >/dev/null 2>&1; then
  echo -e "Claude CLI not found."
  exit 1
fi

claude_msg="$(claude --model claude-sonnet-4-0 -p "$prompt" 2>&1)"
claude_exit=$?

if [[ $claude_exit -ne 0 ]]; then
  echo -e "Claude Error: ${claude_msg:0:100}"
  exit 1
fi

git commit -m "$claude_msg"
git push -u origin "$(git rev-parse --abbrev-ref HEAD)"
