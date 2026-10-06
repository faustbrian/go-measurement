# Migration From shipit/measurements

Inventory every legacy numeric field together with its implicit unit before
changing types. Replace floats or bare decimals with `Quantity` at service
boundaries, using the payload's declared unit. Replace implicit conversions
with `ExactConversion` or a documented `RoundedConversion` context.

Map legacy width, length, and height structures to `NewDimensions`; reject
missing, zero, negative, or mixed non-length fields rather than preserving an
invalid zero value. Move loading-metre defaults, stacking decisions, and
volumetric divisors into caller configuration because they are carrier policy.

For persisted values, migrate to `{value,unit}` JSON or two explicit columns.
Do not assume historical rows use the current preferred unit. Dual-read old and
new representations during rollout, compare exact canonical values, then stop
writing the old shape. Preserve golden fixtures from Track, Postal, and
Location during migration.

## Adopt caller cancellation

Existing v1 operations without a `Context` suffix remain source- and
behavior-compatible and run with `context.Background()`. For request-, job-,
or deadline-bounded work, replace such a call with its additive `...Context`
counterpart and pass the operation context first:

```go
converted, err := quantity.ConvertContext(ctx, target, conversion)
```

Keep passing `ConversionContext` separately; it owns arithmetic precision,
rounding, and limits, not cancellation. There is no required migration for
callers that intentionally retain the detached v1 behavior.

## Adopt Measurement v3 and Wire v3 together

V3 changes the root import to `github.com/faustbrian/go-measurement/v3` and
uses the public `github.com/faustbrian/go-wire/v3@v3.0.0` dependency. Update
measurement imports, both adapter paths, and Wire imports together:

```go
import (
    measurement "github.com/faustbrian/go-measurement/v3"
    measurementwire "github.com/faustbrian/go-measurement/v3/adapters/wire"
    "github.com/faustbrian/go-wire/v3"
)
```

The adapter's `Encode` and `Decode` signatures now take Wire v3's nominal
`wire.Format` type. Classify returned errors using Wire v3 sentinels and
`errors.As` with `*wire.Error` from the same import. Wire v1/v2 types and
sentinels are different identities and cannot substitute for v3's contracts.
The selected formats remain JSON and XML; quantity fields, decimal scale,
unit metadata, caller byte limits, bounded XML entrypoints, and redacted
diagnostics retain their existing contracts. Wire v3's top-level error text
is a fixed class rather than detailed diagnostic text; use sentinel matching
for classification, not error strings.

`github.com/faustbrian/go-measurement/v3/measurementwire` remains a distinct
delegating facade with its own named `Options` type; it is not a type alias
for `adapters/wire.Options`. This major migration does not remove the facade.
The earliest permitted removal remains a future major release and never
before 2027-03-08.

Existing v1/v2 consumers can coexist with v3 and do not change automatically.
Owned v1 consumers `go-knapsack` and `go-rule-engine/adapters/measurement`
must retain the bounded-XML migration and consumer verification described
below when they explicitly adopt the new major.

## Replace raw XML unmarshalling

This migration applies to the published
`github.com/faustbrian/go-measurement/v2` module, introduced in v2.0.0.
V1 consumers must update the module import path and adopt the bounded XML
entrypoints. The v2.0.2 maintenance patch retains these requirements and
does not require an additional migration from v2.0.1.

Raw `encoding/xml.Unmarshal` and `Decoder.Decode` calls targeting `Quantity`
or `Dimensions` now fail with `ErrUnboundedXML`. Replace them with the bounded
entrypoint matching the document root:

```go
quantity, err := measurement.ParseQuantityXML(payload)
dimensions, err := measurement.ParseDimensionsXML(payload)
```

Use `adapters/wire.Decode` when the caller also needs wire-format selection and
a smaller caller-defined byte limit. Apply a transport body limit before
reading the payload into memory. This migration is required because an
`encoding/xml.Unmarshaler` callback cannot bound allocation of the start token
that the external decoder parses before invoking it.
The adapter classifies malformed XML tokenization as `wire.ErrParse`, distinct
from schema and value failures classified as `wire.ErrValidation`.

The owned direct consumers `go-knapsack` (including its objective and
reference modules) and `go-rule-engine/adapters/measurement` require this
migration and consumer verification when adopting v2. Their v1 dependencies
must be updated only with the bounded-XML migration and consumer verification.

## Handle redacted JSON diagnostics

The v2 module retains `errors.Is` classification while replacing
attacker-controlled JSON diagnostic text. `errors.As` no longer exposes an
underlying `json.SyntaxError` from package-owned decoding; a reachable
`json.UnmarshalTypeError` is a sanitized copy whose `Value` is `invalid value`.
Other typed fields may describe the decode location, but are not stable keys
for logging, classification, or retries. Classify with the measurement or wire
sentinels and log only fixed error classes.

The standard library can reject malformed JSON before calling a type's
`UnmarshalJSON` method. When accepting untrusted bytes, call the bounded
`Quantity.UnmarshalJSON` or `adapters/wire.Decode` entrypoint directly and
apply a transport limit before materializing the payload.
