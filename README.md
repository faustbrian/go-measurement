# measurement

[![CI](https://github.com/faustbrian/go-measurement/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-measurement/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-measurement/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-measurement/v3.svg)](https://pkg.go.dev/github.com/faustbrian/go-measurement/v3)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-measurement?sort=semver)](https://github.com/faustbrian/go-measurement/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`measurement` is an immutable, exact, unit-safe measurement package for
Track, Postal, Location, and logistics services. It uses
[`math`](https://github.com/faustbrian/go-math) decimals exclusively and never requires binary
floating-point conversion.

Install the v3 module after its public release is available:

```sh
go get github.com/faustbrian/go-measurement/v3@v3.0.0
```

V3 adopts `github.com/faustbrian/go-wire/v3@v3.0.0`. Update measurement
imports and the Wire `Format`, error type, and sentinel imports together;
see the [migration guide](docs/migration.md). Existing v1 and v2 modules
remain separate major-version dependencies.

```go
length := measurement.MustNew(decimal.MustParse("1.25"), measurement.Metre)
centimetres, err := length.Convert(measurement.Centimetre, measurement.ExactConversion())
// centimetres.String() == "125.00 cm"
```

The package covers length, area, volume, mass, absolute temperature, density,
loading metre, dimensional weight, and rectangular package triples. Compatible
quantities can be converted, compared, added, subtracted, multiplied, divided,
rounded, clamped, counted, formatted, and serialized. Absolute temperatures
can be converted and compared but not added or subtracted because temperature
intervals are not part of the measurement model.

Every conversion selects either `ExactConversion()` or
`RoundedConversion(scale, mode)`. Unit aliases are accepted only through a
caller-selected `Profile`; no locale or preferred unit is inferred.
`context.Context` controls caller cancellation and deadlines, while
`ConversionContext` controls arithmetic limits and rounding. Use the
context-suffixed arithmetic, conversion, formatting, and logistics methods for
request- or job-scoped work. The v1 methods without that suffix remain
compatible and run with `context.Background()`.

## Documentation

Use the [documentation index](docs/README.md) for adoption, numeric contracts,
supported units, logistics formulas, serialization, and operations guidance.
Specification-backed behavior and its executable evidence are recorded in the
[specification decision register](docs/specification-decisions.md) and
[conformance matrix](specification/README.md).

The v3 module retains v2's `ParseQuantityXML`, `ParseDimensionsXML`, and
`adapters/wire` entrypoints bound XML decoding. Raw `encoding/xml`
unmarshalling into measurement types fails closed because its callback
cannot bound the already-parsed start token. V1 consumers must update
imports to `github.com/faustbrian/go-measurement/v3` and adopt these bounded
XML entrypoints. V2 consumers retain the same XML entrypoints in v3.

For ecosystem-wide selection and ownership guidance, see the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Domain utilities family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

Run `make check` for all blocking local gates. See [CONTRIBUTING.md](CONTRIBUTING.md)
and [CHANGELOG.md](CHANGELOG.md).
