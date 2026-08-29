#!/usr/bin/env bash
set -euo pipefail

if find . -type f -name '*.go' ! -path './.git/*' ! -path './.golib-tooling/*' \
	! -path './.verification/*' ! -name '*_test.go' -exec grep -En \
	'go:linkname|import "C"|(^|[^[:alnum:]_])unsafe([^[:alnum:]_]|$)' {} +; then
	printf 'forbidden unsafe or cgo mechanism found\n' >&2
	exit 1
fi
grep -Eq 'MaxTextBytes' parse.go
grep -Eq 'MaxSerializedBytes' encoding.go
grep -Eq 'MaxPackageQuantity' dimensions.go
printf 'measurement trust-boundary guards verified\n'
