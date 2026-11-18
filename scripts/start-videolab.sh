#!/usr/bin/env bash
set -e
cd "$(dirname "$(readlink -f "$0")")/.."
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
branch=$(git branch --show-current)
dirty=$(git status --porcelain)
if [ "$branch" = "main" ] && [ -z "$dirty" ]; then
  echo "Checking GitHub for updates..."
  if git fetch --quiet origin main; then
    git merge --ff-only origin/main || echo "Fast-forward unavailable; local checkout preserved."
  else
    echo "Offline: using current checkout."
  fi
else
  echo "Skipping update (branch=$branch, local changes present=$([ -n "$dirty" ] && echo yes || echo no))."
fi
exec wails dev -tags webkit2_41
