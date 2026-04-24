#!/usr/bin/env bash
# Uploads an OpenAPI spec file to openapiviewer.com and optionally creates a
# GitHub commit status with the resulting viewer URL.
#
# Usage: upload_openapi.sh <file> [context]
#   file    - path to the OpenAPI YAML file
#   context - GitHub status context label (optional; skipped if GH_TOKEN or SHA are unset)

set -euo pipefail

FILE="${1:?usage: upload_openapi.sh <file> [context]}"
CONTEXT="${2:-}"
UPLOAD_URL="${OPENAPI_UPLOAD_URL:-https://www.openapiviewer.com/api/upload}"

echo ">> uploading ${FILE}" >&2
url=$(curl -fsSL -X POST \
    -F "file=@${FILE};type=application/yaml;charset=utf-8" \
    "${UPLOAD_URL}" | jq -r .fullPath)
echo "${url}"

if [[ -n "${CONTEXT}" && -n "${GH_TOKEN:-}" && -n "${SHA:-}" ]]; then
    gh api "repos/${GITHUB_REPOSITORY}/statuses/${SHA}" \
        --method POST \
        --field state=success \
        --field context="${CONTEXT}" \
        --field description="View ${CONTEXT} spec" \
        --field target_url="${url}" \
        || echo "warning: could not create GitHub status (fork PR or insufficient permissions)." >&2
fi
