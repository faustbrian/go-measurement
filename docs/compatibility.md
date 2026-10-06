# Compatibility

The minimum supported Go version is 1.27.0. Public API compatibility follows
semantic versioning. Incompatible changes are recorded in
the changelog and API baseline.

Unit identities, symbols, dimension assignments, conversion ratios,
serialization field names, error sentinels, and formula meanings are public
contracts. Aliases outside `SymbolProfile` are caller policy and are not global
compatibility commitments. Adding a unit requires collision review because a
previously unknown symbol may become accepted.

The context-suffixed operation methods added in v1.1.0 are additive. Every
corresponding v1 method without that suffix retains its signature, result,
validation, and error behavior and delegates with `context.Background()`.
Migrate only callers that need operation cancellation or deadline propagation.

The published v2 stream uses `github.com/faustbrian/go-measurement/v2`.
The v2.0.2 maintenance patch retains its Go 1.27 minimum, APIs, and bounded
JSON/XML contracts. The v2 major makes raw `encoding/xml` unmarshalling
into `Quantity` and `Dimensions` fail closed: those callbacks cannot
constrain allocation of the start token parsed by a caller-owned decoder.
V1 consumers must update imports and migrate to `ParseQuantityXML`,
`ParseDimensionsXML`, or `adapters/wire` when adopting the published v2
module. Existing v2 consumers require no additional migration for this patch.

The v3 stream uses `github.com/faustbrian/go-measurement/v3` and adopts
`github.com/faustbrian/go-wire/v3@v3.0.0`. The adapter's public `wire.Format`
parameter and `*wire.Error` classification move together to Wire v3; update
Wire imports and sentinels alongside measurement imports. This nominal type
change requires the new Measurement major even though supported formats,
roundtrips, byte limits, unit semantics, and bounded XML behavior are retained.
Wire v3 renders fixed error classes rather than detailed diagnostic text.
Existing major versions remain separate dependencies; migration is explicit.
See [the migration guide](migration.md#adopt-measurement-v3-and-wire-v3-together).

`math` owns decimal representation, rounding modes, errors, conditions, and
resource limits. `wire` is used only by the optional `adapters/wire` package.
The former `measurementwire` path is a deprecated delegation-only facade that
retains its named `Options` type and function signatures. It is first eligible
for removal in a future major release and never before 2027-03-08. `money` owns monetary values
and `geo` owns earth-coordinate distance semantics.
