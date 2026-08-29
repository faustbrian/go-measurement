#!/usr/bin/env bash
set -euo pipefail

if grep -REn --include='*.go' --exclude='*_test.go' \
	'go:linkname|import "C"|(^|[^[:alnum:]_])unsafe([^[:alnum:]_]|$)' .; then
	printf 'forbidden unsafe or cgo mechanism found\n' >&2
	exit 1
fi
grep -Eq 'MaxTextBytes' parse.go
grep -Eq 'MaxSerializedBytes' encoding.go
grep -Eq 'MaxPackageQuantity' dimensions.go
printf 'measurement trust-boundary guards verified\n'
