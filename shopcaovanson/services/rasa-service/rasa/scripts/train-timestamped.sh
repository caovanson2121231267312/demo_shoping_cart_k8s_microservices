#!/bin/bash
set -euo pipefail
cd /app

ts="$(date +%Y%m%d-%H%M%S)"
name="chatbot-${ts}"

echo "[rasa] training model: ${name}.tar.gz"
rasa train \
  --fixed-model-name "$name" \
  --data data \
  --config config.yml \
  --domain domain.yml \
  --out models

echo "${name}.tar.gz" > models/latest.txt
echo "[rasa] saved models/${name}.tar.gz (latest)"
