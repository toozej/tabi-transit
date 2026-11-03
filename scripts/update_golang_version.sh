#!/usr/bin/env bash
set -Eeuo pipefail

usage() {
  echo "usage: $0 <go-version>" >&2
  echo "example: $0 1.27.0" >&2
  exit 2
}

if [[ "$#" -ne 1 ]]; then
  usage
fi

requested_version="$1"
if [[ ! "$requested_version" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
  echo "invalid Go version: $requested_version (expected major.minor[.patch])" >&2
  exit 2
fi

# This repository pins a patch release. Accepting a major/minor input keeps the
# common `scripts/update_golang_version.sh 1.27` workflow convenient while
# recording the corresponding initial patch release consistently everywhere.
if [[ "$requested_version" =~ ^[0-9]+\.[0-9]+$ ]]; then
  new_version="${requested_version}.0"
else
  new_version="$requested_version"
fi

repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || {
  echo "must be run from inside a Git repository" >&2
  exit 1
}
cd "$repo_root"

old_version="$(awk '$1 == "go" { print $2; exit }' go.mod)"
if [[ ! "$old_version" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
  echo "could not determine the Go version from $repo_root/go.mod" >&2
  exit 1
fi

old_minor="${old_version%.*}"
new_minor="${new_version%.*}"
if [[ "$old_version" =~ ^[0-9]+\.[0-9]+$ ]]; then
  old_minor="$old_version"
fi

if [[ "$old_version" == "$new_version" ]]; then
  echo "No update needed; repository already pins Go $new_version"
  exit 0
fi

sed_in_place() {
  local expression="$1"
  shift

  if sed --version >/dev/null 2>&1; then
    sed -i -e "$expression" "$@"
  else
    sed -i '' -e "$expression" "$@"
  fi
}

replace_literal() {
  local file="$1"
  local from="$2"
  local to="$3"
  local escaped_from="${from//./\\.}"

  if grep -Fq -- "$from" "$file"; then
    sed_in_place "s/${escaped_from}/${to}/g" "$file"
    printf 'updated %s\n' "$file"
  fi
}

# Keep every Go module and the workspace directive on the same patch release.
while IFS= read -r -d '' file; do
  replace_literal "$file" "$old_version" "$new_version"
done < <(git ls-files -z -- 'go.work' ':(glob)**/go.mod')

# These files intentionally accept any patch release of the selected Go minor.
replace_literal scripts/doctor.sh "go${old_minor}." "go${new_minor}."
replace_literal scripts/doctor.sh "Go ${old_minor}.x" "Go ${new_minor}.x"
replace_literal tools/toolchain/version.go "go${old_minor}." "go${new_minor}."
replace_literal tools/toolchain/version_test.go "go${old_minor}." "go${new_minor}."
replace_literal tools/toolchain/version_test.go "Go ${old_minor} patch release" "Go ${new_minor} patch release"

# The local-development runbook documents the current required release. Do not
# search all documentation: benchmark records may correctly mention old Go
# versions as historical context.
replace_literal docs/runbooks/local-development.md "$old_version" "$new_version"

echo "Updated repository Go version from $old_version to $new_version"
