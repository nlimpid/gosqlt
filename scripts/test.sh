#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

COVERDIR="${COVERDIR:-coverage}"
COVERMODE="${COVERMODE:-set}"
GO_COVERPKG="${GO_COVERPKG:-github.com/nlimpid/gosqlt/scanner}"
ROOT_PROFILE="${COVERDIR}/unit.coverprofile"
INTEGRATION_PROFILE="${COVERDIR}/integration.coverprofile"
MERGED_PROFILE="${COVERDIR}/coverage.out"

ARGS=("$@")

rm -rf "${COVERDIR}"
mkdir -p "${COVERDIR}"

echo ">> running unit tests with coverage"
if ! ROOT_PACKAGES_OUTPUT="$(go list ./...)"; then
	echo "failed to list root packages" >&2
	exit 1
fi
ROOT_PACKAGES="$(echo "${ROOT_PACKAGES_OUTPUT}" | grep -v "/tests" || true)"
if [ -n "${ROOT_PACKAGES}" ]; then
	# shellcheck disable=SC2086
	go test \
		-covermode="${COVERMODE}" \
		-coverprofile="${ROOT_PROFILE}" \
		"${ARGS[@]}" \
		${ROOT_PACKAGES}
else
	echo "mode: ${COVERMODE}" > "${ROOT_PROFILE}"
fi

echo ">> running integration tests with coverage"
if [ -d "${REPO_ROOT}/tests" ]; then
	pushd "${REPO_ROOT}/tests" >/dev/null
	if ! INTEGRATION_OUTPUT="$(go list ./...)"; then
		echo "failed to list integration packages" >&2
		exit 1
	fi
	INTEGRATION_PACKAGES="${INTEGRATION_OUTPUT}"
	if [ -n "${INTEGRATION_PACKAGES}" ]; then
		# shellcheck disable=SC2086
		go test \
			-covermode="${COVERMODE}" \
			-coverpkg="${GO_COVERPKG}" \
			-coverprofile="../${INTEGRATION_PROFILE}" \
			"${ARGS[@]}" \
			${INTEGRATION_PACKAGES}
	else
		echo "mode: ${COVERMODE}" > "../${INTEGRATION_PROFILE}"
	fi
	popd >/dev/null
fi

echo ">> merging coverage profiles into ${MERGED_PROFILE}"
echo "mode: ${COVERMODE}" > "${MERGED_PROFILE}"
for profile in "${ROOT_PROFILE}" "${INTEGRATION_PROFILE}"; do
	if [ -s "${profile}" ]; then
		tail -n +2 "${profile}" >> "${MERGED_PROFILE}"
	fi
done

echo ">> coverage ready at ${MERGED_PROFILE}"
