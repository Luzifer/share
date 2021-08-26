#!/bin/bash
set -euo pipefail

outfile=${1:-}
[[ -n $outfile ]] || {
	echo "Missing outfile parameter" >&2
	exit 1
}
shift

[ $# -gt 0 ] || {
	echo "Missing combine files" >&2
	exit 1
}

IFS=$','
exec curl -sSfLo "${outfile}" "https://cdn.jsdelivr.net/combine/$*"
