#!/usr/bin/env bash
set -euo pipefail

for topic in transactions balances; do
  kafka-topics --bootstrap-server kafka:29092 --create --if-not-exists \
    --topic "$topic" --partitions 1 --replication-factor 1
done
