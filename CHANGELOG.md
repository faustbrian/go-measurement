# Changelog

All notable changes follow Keep a Changelog. The project uses semantic
versioning.

## Unreleased

### Changed

- Align isolated dependency checks and architecture guards with standalone
  package module paths.
- Adopt the released `go-library-tools` v1.0.13 contract while retaining the
  measurement-specific architecture, security, and documentation checks.

### Documentation

- Replace repeated README links with the repository-local documentation index.
- Add the [specification decision register](docs/specification-decisions.md),
  conformance matrix, source monitoring, and typed conformance and
  interoperability gates for the existing measurement contract.

### Specification Decisions

- MEASUREMENT-DEC-001 sha256:de9aa2fd27fcd6f8776bbc9ff08aabe008922542fd298dc9b8e442b51cff71e3
- MEASUREMENT-DEC-002 sha256:b404435f23d7772bb49bdf31d5340af3dfbe724d782672ffb4ab5720f0264ecf
- MEASUREMENT-DEC-003 sha256:9e91e6be2094c8f5716604c6aaea6db691e31d44257b5742a872bbd304a69aa9
- MEASUREMENT-DEC-004 sha256:4c5693dec5e261ac8a39893d0e9c836f157b2635ae7faaef63f392bbb7a7fe52
- MEASUREMENT-DEC-005 sha256:9f1530a6409651204e38c81bc9183b40e461e77f83f1753f6b0bc1e346921416
- MEASUREMENT-DEC-006 sha256:0f86101b227f12e9982fdcd30dbc96513ff780d4e2afa1bd35f5dfe5b8d3e1e4

## 1.0.0 - 2026-08-25

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Correct stale package, standalone, and authoritative-source links in public
  documentation.

### Documentation

- Add package discovery documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-measurement` identity while preserving its documented API and behavior.
- Strengthen exact-limit mutation coverage and simplify equivalent boundary
  expressions without changing the accepted measurement domain.
- Delegate local mutation checks to the canonical exact-100 repository runner
  and remove the superseded package-local Gremlins configuration.
- Require owned sibling modules at local `v0.0.0`; clean external consumers
  pin each module to an exact main pseudo-version.

- Refresh owned-module checksums against the final consolidated archives.
- Normalized standalone module metadata against the canonical owned dependency
  graph, including complete checksums for clean consumer resolution.

### Added

- Immutable quantities backed exclusively by `math/decimal`.
- Closed dimensions and explicit exact or rounded conversion contexts.
- SI and logistics units for length, area, volume, mass, temperature, density,
  and loading metre.
- Compatible arithmetic, comparison, rounding, clamping, and package counts.
- Validated dimension triples, volume, floor area, loading metre, volumetric
  divisor, and volumetric index formulas.
- Lossless JSON, XML, SQL, and bounded `wire` adapters.
- Property, fixture, fuzz, race, mutation, coverage, and benchmark gates.

- `NewProfile` now returns an error and rejects oversized or invalid alias
  catalogs.
- JSON and XML decoding rejects duplicate fields; direct constructors enforce
  the default `math` decimal limits.
