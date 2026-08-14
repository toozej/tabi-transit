#!/usr/bin/env bash
set -u

failures=0
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
node_version_file="${root}/.nvmrc"

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

check_command 'Go' go
check_command 'Docker' docker
check_command 'Git' git

if command -v brew >/dev/null 2>&1 && [ -f "${node_version_file}" ]; then
  nvm_script="$(brew --prefix nvm 2>/dev/null)/nvm.sh"
  if [ ! -s "${nvm_script}" ]; then
    printf 'missing: Homebrew nvm; run make nvm-install\n' >&2
    failures=$((failures + 1))
  else
  export NVM_DIR="${root}/.tools/nvm"
  set -- --no-use
  # shellcheck source=/dev/null
  . "${nvm_script}"
  set --
  requested_node_version="$(tr -d '[:space:]' < "${node_version_file}")"
  printf 'ok: nvm %s\n' "$(nvm --version)"

  if [ "$(nvm version "${requested_node_version}")" != 'N/A' ]; then
    node_version="$(nvm exec --silent "${requested_node_version}" node --version)"
    printf 'ok: Node.js version %s via repository nvm\n' "${node_version}"
    if nvm exec --silent "${requested_node_version}" corepack pnpm --version >/dev/null 2>&1; then
      printf 'ok: pnpm %s via repository nvm\n' "$(nvm exec --silent "${requested_node_version}" corepack pnpm --version)"
    else
      printf 'missing: pnpm via Corepack for Node.js %s; run make prereqs with network access\n' "${requested_node_version}" >&2
      failures=$((failures + 1))
    fi
  else
    printf 'missing: Node.js %s in repository nvm; run make prereqs\n' "${requested_node_version}" >&2
    failures=$((failures + 1))
  fi
  fi
else
  printf 'missing: Homebrew nvm; run make nvm-install\n' >&2
  failures=$((failures + 1))
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

if [ -f .env ]; then
  printf 'warning: .env exists locally and is intentionally ignored by Git\n' >&2
fi

if [ "$failures" -gt 0 ]; then
  printf 'doctor: %s required prerequisite(s) failed\n' "$failures" >&2
  exit 1
fi

printf 'doctor: required local prerequisites are available\n'
