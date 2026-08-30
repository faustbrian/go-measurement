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
