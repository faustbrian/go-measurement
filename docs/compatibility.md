# Compatibility

The minimum supported Go version is 1.26.6. Public API compatibility follows
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

`math` owns decimal representation, rounding modes, errors, conditions, and
resource limits. `wire` is used only by the optional `measurementwire`
subpackage. `money` owns monetary values and `geo` owns earth-coordinate
distance semantics.
