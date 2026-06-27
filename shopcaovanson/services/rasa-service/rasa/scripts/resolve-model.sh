#!/bin/bash
set -euo pipefail
cd /app

if [ -f models/latest.txt ]; then
  name="$(tr -d '\r\n' < models/latest.txt)"
  if [ -n "$name" ] && [ -f "models/$name" ]; then
    echo "models/$name"
    exit 0
  fi
fi

latest="$(ls -t models/chatbot-*.tar.gz 2>/dev/null | head -1 || true)"
if [ -n "$latest" ]; then
  echo "$latest"
  exit 0
fi

if [ -f models/chatbot.tar.gz ]; then
  echo models/chatbot.tar.gz
  exit 0
fi

exit 1
