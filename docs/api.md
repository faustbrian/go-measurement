# API And Supported Units

`Quantity` owns a `decimal.Decimal` amount and `Unit`. `New` validates unit
identity; accessors return immutable values. `Convert`, `Add`, `Subtract`,
`Compare`, `Equal`, `Multiply`, `Divide`, `Round`, `Clamp`, `Times`, and
`Format` return new values. Their additive context-suffixed counterparts accept
`context.Context` first and propagate cancellation and deadlines through every
composed decimal operation:

- `ConvertContext`, `AddContext`, `SubtractContext`, `CompareContext`, and
  `EqualContext`;
- `MultiplyContext`, `DivideContext`, `RoundContext`, `TimesContext`,
  `ClampContext`, and `FormatContext`;
- `FloorAreaContext`, `CubicVolumeContext`, `TotalVolumeContext`, and
  `LoadingMetresContext`; and
- both `VolumetricDivisor.WeightContext` and
  `VolumetricIndex.WeightContext`.

Nil context is invalid. Pre-cancellation and expired deadlines take precedence
over operation arguments. The v1 methods without a `Context` suffix remain
available and delegate with `context.Background()`.

Quantities, dimensions, and conversion policies are immutable and safe for
concurrent reuse. `FormatOptions` is caller-owned and copied by value; it is
safe to reuse concurrently only when callers do not mutate it concurrently.
Operations own no goroutines, resources, or shutdown lifecycle, and no
operation context is retained after a call returns.

| Dimension | Units |
| --- | --- |
| Dimensionless | `1` |
| Length | `mm`, `cm`, `m`, `km`, `in`, `ft`, `yd` |
| Area | `mm2`, `cm2`, `m2`, `in2`, `ft2` |
| Volume | `mm3`, `cm3`, `m3`, `mL`, `L`, `in3`, `ft3` |
| Mass | `mg`, `g`, `kg`, `t`, `oz`, `lb` |
| Absolute temperature | `K`, `degC`, `degF` |
| Density | `kg/m3`, `g/cm3` |
| Loading metre | `ldm` |

`Units(dimension)` returns a stable sorted copy of the catalog. V1 dimensions
are closed: unsupported derived exponents return `ErrUnsupportedDimension`.
Loading metres deliberately cannot combine with ordinary lengths.

`Dimensions` validates positive length, width, and height plus a bounded count.
It calculates floor area, package volume, total volume, and loading metres.
`VolumetricDivisor` and `VolumetricIndex` calculate dimensional mass.

## Wire adapter

Import `github.com/faustbrian/go-measurement/adapters/wire` when an untrusted
JSON or XML quantity document needs a caller-selected byte limit. The adapter
exports `Options{MaxBytes int64}`, `Encode(Quantity, wire.Format, Options)`, and
`Decode([]byte, wire.Format, Options)`. A zero `MaxBytes` retains the selected
`go-wire` codec's default limit; a positive value supplies an explicit
limit.

JSON decoding is strict and rejects unknown fields. XML decoding requires the
`quantity` root. Both formats preserve decimal text and unit identity, and the
decoded value does not retain the caller's payload. Oversized documents remain
classifiable with `wire.ErrSizeLimit`; unsupported or unknown formats remain
classifiable with `wire.ErrUnsupportedFormat`.

`github.com/faustbrian/go-measurement/measurementwire` retains the same
`Options`, `Encode`, and `Decode` signatures as a deprecated delegation-only
compatibility facade. Its named `Options` type keeps the legacy package's
reflection identity. New callers should use `adapters/wire`; the earliest
permitted removal of the legacy path is v2.0.0 and never before 2027-03-08.
