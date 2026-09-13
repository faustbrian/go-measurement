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

## Replace raw XML unmarshalling

This migration targets the unpublished
`github.com/faustbrian/go-measurement/v2` module. Existing consumers must stay
on released v1 until v2 is published; do not add a replace directive or a
pseudo-version to bypass that release boundary. After publication, update the
module import path before adopting the new XML entrypoints.

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

The owned direct consumers `go-knapsack` (including its objective and
reference modules) and `go-rule-engine/adapters/measurement` remain pinned to
v1. Their migration and consumer verification are release-blocked until an
actual v2 version exists.
