// Package measurement provides immutable exact logistics quantities with
// explicit units, dimensions, conversion policy and arithmetic bounds. Start
// with New or Parse using SymbolProfile, then choose ExactConversion or
// RoundedConversion for compatible-unit operations.
//
// Decimal arithmetic is owned by github.com/faustbrian/go-math/decimal; core
// quantities never convert through binary floating point. The closed unit
// catalog uses exact source-defined ratios; loading metres remain semantically
// distinct from ordinary lengths. Absolute temperatures can be converted and
// compared, but addition/subtraction reject unmodeled temperature intervals.
// Negative and zero amounts are allowed in Quantity; logistics constructors
// separately require positive dimensions and formula inputs.
//
// Quantity, Dimensions, conversion policies and Profiles can be shared for
// concurrent value operations. Pointer decoding methods replace their receiver
// only on success and require exclusive caller ownership during mutation.
// Context-suffixed operations require a non-nil context, check cancellation
// before other arguments and do not retain it. They return no partial result
// after an observed error; cancellation does not forcibly interrupt a decimal
// primitive. Methods without the suffix use context.Background().
//
// No package introduces cgo, external services or background work. The package
// does not infer units/locale, select carrier tariffs, pack cargo or own database
// transactions. The adapters/wire package handles bounded JSON/XML codecs;
// measurementwire retains the deprecated compatibility facade. Use the bounded
// XML parse functions rather than raw encoding/xml unmarshalling.
package measurement
