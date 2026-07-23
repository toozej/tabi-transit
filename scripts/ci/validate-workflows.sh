#!/usr/bin/env bash
# Validate the repository's GitHub Actions files without needing GitHub access.
# This intentionally checks the small workflow schema we rely on and then uses
# actionlint when it is installed by a developer or CI image.
set -Eeuo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
workflow_dir="$repo_root/.github/workflows"

if [[ ! -d "$workflow_dir" ]]; then
  printf 'error: workflow directory is missing: %s\n' "$workflow_dir" >&2
  exit 1
fi

mapfile -t workflows < <(find "$workflow_dir" -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) -print | sort)
if (( ${#workflows[@]} == 0 )); then
  printf 'error: no GitHub Actions workflow files found\n' >&2
  exit 1
fi

for workflow in "${workflows[@]}"; do
  ruby -e '
    require "yaml"
    path = ARGV.fetch(0)
    document = YAML.safe_load(File.read(path), aliases: true)
    abort("error: #{path} must contain a mapping") unless document.is_a?(Hash)
    abort("error: #{path} is missing name") unless document["name"].is_a?(String) && !document["name"].empty?
    trigger = document["on"] || document[true]
    abort("error: #{path} is missing on") if trigger.nil?
    jobs = document["jobs"]
    abort("error: #{path} is missing jobs") unless jobs.is_a?(Hash) && !jobs.empty?
    jobs.each do |name, job|
      abort("error: #{path} job #{name} must contain a mapping") unless job.is_a?(Hash)
      if !job.key?("uses") && !job.key?("runs-on")
        abort("error: #{path} job #{name} must define runs-on or uses")
      end
    end
  ' "$workflow"
done

if command -v actionlint >/dev/null 2>&1; then
  actionlint "${workflows[@]}"
else
  printf 'GATE: actionlint is not installed; YAML and required job fields were checked. Install actionlint for expression-level validation.\n' >&2
fi

printf 'workflow validation passed for %d file(s)\n' "${#workflows[@]}"
