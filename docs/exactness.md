# Exactness And Rounding

All amounts and constants are `math/decimal.Decimal`. No package operation
converts through `float32` or `float64`. SI prefixes and international inch,
foot, yard, pound, and ounce definitions are finite decimal ratios. Fahrenheit
uses the exact affine ratio `(F + 459.67) * 5 / 9` before rounding.

`ExactConversion()` rejects a non-terminating base-10 quotient with the shared
`math` conversion error. `RoundedConversion(scale, mode)` combines the full
source-to-target ratio and rounds exactly once at the requested fractional
scale. The zero `ConversionContext` is invalid, preventing ambient defaults.

`Quantity.RoundContext` quantizes an existing amount. `FormatContext` first
converts with its conversion context and then quantizes with its independent
display scale. This separation prevents carrier payload precision from
becoming display policy.

`context.Context` controls cancellation and deadlines. `ConversionContext`
controls exact-versus-rounded arithmetic and limits; it does not retain the
operation context. Context-suffixed methods propagate the same caller context
through every composed decimal operation and return no partial quantity after
observing cancellation. The v1 methods without a `Context` suffix delegate with
`context.Background()` and therefore cannot be canceled by a caller.

Arithmetic uses `math` limits. `WithLimits` can apply tighter bounds. Invalid
limits, oversized exponents, coefficients, inputs, counts, and outputs fail
before unbounded work.
