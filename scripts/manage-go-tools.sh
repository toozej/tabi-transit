#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MANIFEST="${GO_TOOL_MANIFEST:-${ROOT}/tools/go-tools.tsv}"
TOOLS_BIN="${TOOLS_BIN:-${ROOT}/.tools/bin}"

usage() {
	echo "usage: $0 <install [tool ...]|update>" >&2
	exit 2
}

require_tool() {
	local requested="$1"
	if ! awk -F '\t' -v name="${requested}" '$1 == name { found = 1 } END { exit !found }' "${MANIFEST}"; then
		echo "unknown pinned Go tool: ${requested}" >&2
		exit 2
	fi
}

install_with_dependency() (
	local name="$1" package="$2" version="$3" dependency="$4" temporary
	temporary="$(mktemp -d)"
	trap 'rm -rf "${temporary}"' EXIT
	cd "${temporary}"
	export GOWORK=off
	go mod init "tabi.tools/${name}"
	go get "${package}@${version}" "${dependency}"
	go build -trimpath -o "${TOOLS_BIN}/${name}" "${package}"
)

install_tools() {
	local requested name package _module version dependency go_version
	go_version="$(go env GOVERSION)"
	for requested in "$@"; do
		require_tool "${requested}"
	done

	mkdir -p "${TOOLS_BIN}"
	while IFS=$'\t' read -r name package _module version dependency || [[ -n "${name}" ]]; do
		if [[ -z "${name}" || "${name}" == \#* ]]; then
			continue
		fi
		if (( $# > 0 )); then
			local selected=false
			for requested in "$@"; do
				if [[ "${requested}" == "${name}" ]]; then
					selected=true
					break
				fi
			done
			if [[ "${selected}" != true ]]; then
				continue
			fi
		fi

		local stamp="${TOOLS_BIN}/.stamps/${name}"
		local stamp_value="${package}@${version} ${go_version}${dependency:+ ${dependency}}"
		if [[ -x "${TOOLS_BIN}/${name}" && -f "${stamp}" \
			&& "$(cat "${stamp}")" == "${stamp_value}" ]]; then
			echo "Skipping ${name} (${package}@${version} built with ${go_version})"
			continue
		fi

		echo "Installing ${name} (${package}@${version} with ${go_version})"
		if [[ -n "${dependency}" ]]; then
			install_with_dependency "${name}" "${package}" "${version}" "${dependency}"
		else
			GOBIN="${TOOLS_BIN}" go install "${package}@${version}"
		fi
		mkdir -p "${TOOLS_BIN}/.stamps"
		printf '%s\n' "${stamp_value}" >"${stamp}"
	done <"${MANIFEST}"
}

write_manifest_entry() {
	printf '%s\t%s\t%s\t%s' "$1" "$2" "$3" "$4"
	if [[ -n "$5" ]]; then
		printf '\t%s' "$5"
	fi
	printf '\n'
}

update_tools() {
	local name package module version dependency latest updated=false
	local temp
	temp="$(mktemp)"
	trap 'rm -f "${temp:-}"' EXIT

	while IFS=$'\t' read -r name package module version dependency || [[ -n "${name}" ]]; do
		if [[ -z "${name}" || "${name}" == \#* ]]; then
			write_manifest_entry "${name}" "${package}" "${module}" "${version}" "${dependency}" >>"${temp}"
			continue
		fi

		latest="$(GOWORK=off go list -mod=mod -m -f '{{.Version}}' "${module}@latest")"
		if [[ -z "${latest}" ]]; then
			echo "could not resolve latest version for ${module}" >&2
			exit 1
		fi
		if [[ "${latest}" != "${version}" ]]; then
			echo "Updating ${name}: ${version} -> ${latest}"
			version="${latest}"
			updated=true
		fi
		if [[ -n "${dependency}" ]]; then
			latest="$(GOWORK=off go list -mod=mod -m -f '{{.Version}}' "${dependency%@*}@latest")"
			if [[ -z "${latest}" ]]; then
				echo "could not resolve latest version for ${dependency%@*}" >&2
				exit 1
			fi
			if [[ "${dependency}" != "${dependency%@*}@${latest}" ]]; then
				echo "Updating ${name} build dependency: ${dependency} -> ${latest}"
				dependency="${dependency%@*}@${latest}"
				updated=true
			fi
		fi
		write_manifest_entry "${name}" "${package}" "${module}" "${version}" "${dependency}" >>"${temp}"
	done <"${MANIFEST}"

	mv "${temp}" "${MANIFEST}"
	trap - EXIT
	if [[ "${updated}" != true ]]; then
		echo "Pinned Go tools are already current"
	fi
}

case "${1:-}" in
install)
	shift
	install_tools "$@"
	;;
update)
	if (( $# != 1 )); then
		usage
	fi
	update_tools
	;;
*)
	usage
	;;
esac
