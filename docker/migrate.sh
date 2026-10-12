#!/usr/bin/env bash
set -euo pipefail

db() {
  mysql --protocol=tcp --host="$DB_HOST" --user="$DB_USER" \
    --connect-timeout=5 --batch --skip-column-names "$DB_NAME" "$@"
}

db -e 'CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(100) PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;'

for migration in /migrations/*.sql; do
  version="$(basename "$migration")"
  # Filenames are also SQL values; accept only the versioned filename convention.
  [[ "$version" =~ ^[0-9]+_[a-z_]+\.sql$ ]] || { echo "Invalid migration filename" >&2; exit 1; }
  applied="$(db -e "SELECT COUNT(*) FROM schema_migrations WHERE version = '$version';")"
  if [ "$applied" = "0" ]; then
    echo "Applying $DB_NAME/$version"
    db < "$migration"
    db -e "INSERT INTO schema_migrations (version) VALUES ('$version');"
  else
    echo "Already applied $DB_NAME/$version"
  fi
done
