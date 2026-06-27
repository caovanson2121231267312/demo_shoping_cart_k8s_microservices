#!/bin/sh
set -e

cd /app
chmod +x scripts/*.sh 2>/dev/null || true

echo "[rasa] waiting for action server..."
i=0
while [ "$i" -lt 45 ]; do
  if wget -q -O /dev/null http://rasa-actions:5055/health 2>/dev/null; then
    echo "[rasa] action server ready"
    break
  fi
  i=$((i + 1))
  sleep 2
done

if [ "${FORCE_RASA_TRAIN:-false}" = "true" ]; then
  bash scripts/train-timestamped.sh
elif ! MODEL=$(bash scripts/resolve-model.sh 2>/dev/null); then
  bash scripts/train-timestamped.sh
fi
MODEL=$(bash scripts/resolve-model.sh)

echo "[rasa] starting server with model $MODEL"
exec rasa run --enable-api --cors "*" --model "$MODEL"
