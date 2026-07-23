#!/usr/bin/env bash
# Validate a candidate immutable backend image reference before a future deploy
# workflow is permitted to use it. This script never authenticates, pushes, or
# deploys an image.
set -Eeuo pipefail

if (( $# != 1 )); then
  printf 'usage: %s REGISTRY/IMAGE@sha256:64-hex-digest\n' "$0" >&2
  exit 2
fi

image_ref="$1"
if [[ ! "$image_ref" =~ ^[a-z0-9][a-z0-9._/-]*@sha256:[a-f0-9]{64}$ ]]; then
  printf 'error: release image must be a lowercase immutable sha256 digest, not a tag\n' >&2
  exit 1
fi

if [[ ! -f Dockerfile || ! -f docker/healthcheck/main.go || ! -f docker/migrate/main.go ]]; then
  printf 'error: backend image build inputs are incomplete; container release remains disabled\n' >&2
  exit 1
fi

printf 'candidate immutable image reference is valid; Dockerfile is present for a separate build/scan workflow\n'
