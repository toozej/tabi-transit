#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
nvm_dir="${root}/.tools/nvm"
node_version="$(tr -d '[:space:]' < "${root}/.nvmrc")"

if [[ "$#" -eq 0 ]]; then
  echo "usage: $0 <command> [argument ...]" >&2
  exit 2
fi

if [[ ! "${node_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "invalid Node.js version in ${root}/.nvmrc: ${node_version}" >&2
  exit 1
fi

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew nvm is unavailable; run make nvm-install" >&2
  exit 1
fi

nvm_script="$(brew --prefix nvm 2>/dev/null)/nvm.sh"
if [[ ! -s "${nvm_script}" ]]; then
  echo "Homebrew nvm is unavailable; run make nvm-install" >&2
  exit 1
fi

mkdir -p "${nvm_dir}"
export NVM_DIR="${nvm_dir}"
nvm_arguments=("$@")
set -- --no-use
# shellcheck source=/dev/null
. "${nvm_script}"
set -- "${nvm_arguments[@]}"

installed_version="$(nvm version "${node_version}" || true)"
if [[ "${installed_version}" == "N/A" ]]; then
  nvm install "${node_version}"
fi

nvm exec --silent "${node_version}" "$@"
