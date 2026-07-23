#!/usr/bin/env sh
set -u

failures=0

check_command() {
  name="$1"
  command="$2"
  if command -v "$command" >/dev/null 2>&1; then
    printf 'ok: %s (%s)\n' "$name" "$(command -v "$command")"
  else
    printf 'missing: %s (%s)\n' "$name" "$command" >&2
    failures=$((failures + 1))
  fi
}

check_command 'Node.js' node
check_command 'Corepack' corepack
check_command 'Go' go
check_command 'Docker' docker
check_command 'Git' git

if command -v node >/dev/null 2>&1; then
  node_version="$(node --version)"
  case "$node_version" in
    v20.*|v21.*|v22.*) printf 'ok: Node.js version %s\n' "$node_version" ;;
    *) printf 'unsupported: Node.js version %s; require >=20.19.0 and <23\n' "$node_version" >&2; failures=$((failures + 1)) ;;
  esac
fi

if command -v go >/dev/null 2>&1; then
  go_version="$(go env GOVERSION)"
  case "$go_version" in
    go1.26.*) printf 'ok: Go version %s\n' "$go_version" ;;
    *) printf 'unsupported: Go version %s; require Go 1.26.x\n' "$go_version" >&2; failures=$((failures + 1)) ;;
  esac
fi

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1; then
    printf 'ok: Docker Compose plugin\n'
  else
    printf 'missing: Docker Compose plugin\n' >&2
    failures=$((failures + 1))
  fi
fi

if corepack pnpm --version >/dev/null 2>&1; then
  printf 'ok: pnpm %s\n' "$(corepack pnpm --version)"
else
  printf 'missing: pnpm via Corepack; run make bootstrap with network access\n' >&2
  failures=$((failures + 1))
fi

if [ -f .env ]; then
  printf 'warning: .env exists locally and is intentionally ignored by Git\n' >&2
fi

if [ "$failures" -gt 0 ]; then
  printf 'doctor: %s required prerequisite(s) failed\n' "$failures" >&2
  exit 1
fi

printf 'doctor: required local prerequisites are available\n'
