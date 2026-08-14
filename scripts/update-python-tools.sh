#!/usr/bin/env bash

set -euo pipefail

root_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
requirements_file="$root_dir/tools/requirements.txt"
temporary_dir=$(mktemp -d)
temporary_requirements=$(mktemp "$root_dir/tools/requirements.txt.XXXXXX")
trap 'rm -rf "$temporary_dir"; rm -f "$temporary_requirements"' EXIT

mapfile -t packages < <(
  awk -F '==' '/^[[:alnum:]_.-]+==[[:alnum:]_.+!-]+$/ { print $1 }' "$requirements_file"
)
if ((${#packages[@]} == 0)); then
  printf '%s\n' "No direct Python requirements found in $requirements_file" >&2
  exit 1
fi

python3 -m venv "$temporary_dir/venv"
"$temporary_dir/venv/bin/python" -m pip install \
  --disable-pip-version-check \
  --upgrade \
  "${packages[@]}" >/dev/null

while IFS= read -r line || [[ -n "$line" ]]; do
  if [[ "$line" =~ ^([[:alnum:]_.-]+)== ]]; then
    package=${BASH_REMATCH[1]}
    version=$("$temporary_dir/venv/bin/python" -c \
      'from importlib.metadata import version; import sys; print(version(sys.argv[1]))' \
      "$package")
    printf '%s==%s\n' "$package" "$version" >>"$temporary_requirements"
  else
    printf '%s\n' "$line" >>"$temporary_requirements"
  fi
done <"$requirements_file"

mv "$temporary_requirements" "$requirements_file"
trap - EXIT
rm -rf "$temporary_dir"
