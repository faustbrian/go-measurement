# Changelog

All notable changes follow Keep a Changelog. The project uses semantic
versioning.

## 3.0.0 - 2026-10-06

This entry describes the v3.0.0 release source. Adoption requires the public
tag and module artifacts; the date alone does not establish publication.

### Changed

- Adopt `github.com/faustbrian/go-measurement/v3` and the public
  `github.com/faustbrian/go-wire/v3@v3.0.0` dependency. Update measurement and
  Wire imports together because adapter format parameters, error types, and
  sentinels now use Wire v3's nominal identities.
- Retain JSON/XML roundtrips, bounded decoding, caller byte limits, unit and
  decimal metadata, redacted diagnostics, and the Go 1.27 minimum. Wire errors
  render fixed classifications; classify with sentinels instead of strings.
- Keep both `adapters/wire` and the deprecated `measurementwire` facade with
  distinct named `Options` types. The facade is not removed in this major;
  removal remains prohibited before 2027-03-08.

## 2.0.2 - 2026-10-04

This entry describes the v2.0.2 release source. The preceding v2.0.1 was
published on 2026-10-01; its original release record is retained below.

### Changed

- Update the optional Wire codec dependency to v1.0.1 while preserving
  Measurement's v2 unit, conversion, bounded JSON/XML, and cancellation
  contracts, module path, and Go 1.27 minimum.

### Documentation

- Identify the published v2 module in adoption guidance and retain the
  existing bounded-XML migration requirements for v1 consumers.

## 2.0.1 - 2026-10-01

This entry describes the planned v2.0.1 release source. Its date does not
establish publication; adoption requires the public tag and module artifacts.

### Changed

- Select the public `go-math` v1.1.2 security patch for exact decimal arithmetic
  while retaining Measurement's v2 unit, conversion, wire, and cancellation
  contracts.

## 2.0.0 - 2026-09-27

This entry describes the v2.0.0 release source. Its date does not establish
publication; adoption requires the public tag and module artifacts.

### Security

- Bound XML decoding by bytes, schema depth, token and field counts, scalar
  sizes, and disabled reference processing. JSON, XML, SQL, profile, and unit
  diagnostics no longer echo hostile field names, aliases, or values.

### Changed

- Adopt the `github.com/faustbrian/go-measurement/v2` module path. Upgrade
  imports and owned consumers only after its public artifacts are available.
- Make `Quantity.UnmarshalXML` and `Dimensions.UnmarshalXML` fail closed with
  `ErrUnboundedXML`. Replace raw `encoding/xml` decoding with
  `ParseQuantityXML`, `ParseDimensionsXML`, or `adapters/wire`; this is a
  breaking security migration contained in the next major module path.

### Documentation

- Require Go 1.27.0 across the repository's module language, minimum
  compatibility, development, and CI toolchain claims.

## 1.1.0 - 2026-09-09

### Added

- Add `adapters/wire` as the target-oriented entry point for bounded JSON and
  XML quantity encoding and decoding.

- Add context-first conversion, arithmetic, formatting, dimension, and
  volumetric operations that propagate caller cancellation and deadlines
  through composed decimal work.

### Deprecated

- Retain `measurementwire` as a delegation-only compatibility facade with its
  existing `Options`, `Encode`, and `Decode` contracts; new consumers should
  import `adapters/wire`. The earliest permitted removal is v2.0.0 and never
  before 2027-03-08.

### Changed

- Pin `go-math` v1.1.0 while preserving numeric, unit, wire, and non-context
  error behavior.
- Retain every v1 operation without a `Context` suffix as a compatibility
  surface that delegates with `context.Background()`.

- Adopt the `go-library-tools` v1.4.0 schema-v2 cohesion contract and local
  `make cohesion` gate without changing the measurement API or runtime
  behavior.
- Pin reusable CI to the final v1.4.0 W14 workflow, preserve authoritative
  source resolution, and enforce cohesion metadata in the required contract.
- Refresh the reviewed DSV authority fingerprint through its site migration
  after confirming its mapped road-freight formulas and examples remain
  unchanged.
- Adopt the public Go proxy checksums for `go-math` and `go-wire` v1.0.0.

- Align isolated dependency checks and architecture guards with standalone
  package module paths.
- Adopt the released `go-library-tools` v1.2.0 CLI and immutable merged
  workflow `1f9629e5f27418600460b55a50a5b2fc81697fab` while retaining the
  measurement-specific architecture, security, and documentation checks.

### Documentation

- Publish the module's family, capabilities, ownership, lifecycle, supported
  environments, package selection, and delivery status, and link the README to
  the immutable v1.4.0 ecosystem index and family guidance.

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
