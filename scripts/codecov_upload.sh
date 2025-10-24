#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

COVERDIR="${COVERDIR:-coverage}"
COVERPROFILE="${COVER_PROFILE:-${COVERDIR}/coverage.out}"
CODECOV_FLAGS="${CODECOV_FLAGS:-all}"
UPLOADER_PATH="${CODECOV_UPLOADER:-${REPO_ROOT}/bin/codecov.sh}"

if [ ! -s "${COVERPROFILE}" ]; then
	echo "coverage profile ${COVERPROFILE} not found or empty" >&2
	exit 1
fi

mkdir -p "$(dirname "${UPLOADER_PATH}")"
if [ ! -f "${UPLOADER_PATH}" ]; then
	echo ">> downloading codecov uploader script"
	curl -sSfL https://codecov.io/bash -o "${UPLOADER_PATH}"
	chmod +x "${UPLOADER_PATH}"
fi

echo ">> uploading ${COVERPROFILE} to Codecov"
bash "${UPLOADER_PATH}" \
	-f "${COVERPROFILE}" \
	-F "${CODECOV_FLAGS}" \
	-Z
