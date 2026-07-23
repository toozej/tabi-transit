#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

exec 9>/run/lock/tabi-restic-check.lock
flock -n 9 || exit 0

test -r /etc/tabi/secrets/restic_environment
test -r /etc/tabi/secrets/restic_password
# shellcheck disable=SC1091
source /etc/tabi/secrets/restic_environment
export RESTIC_PASSWORD_FILE=/etc/tabi/secrets/restic_password
restic check
