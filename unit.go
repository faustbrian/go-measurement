package measurement

import (
	"errors"
	"fmt"
	"slices"

	"github.com/faustbrian/go-math/decimal"
)

// Package errors classify validation, unit, dimension, context, and parser
// boundary failures.
var (
	// ErrInvalidQuantity reports malformed, nonpositive, or unbounded input.
	ErrInvalidQuantity = errors.New("measurement: invalid quantity")
	// ErrMalformedXML distinguishes XML tokenization failures from validation.
	ErrMalformedXML = errors.New("measurement: malformed XML")
	// ErrUnknownUnit reports a symbol outside the closed unit catalog or alias policy.
	ErrUnknownUnit = errors.New("measurement: unknown unit")
	// ErrDimensionMismatch reports incompatible units or a wrong formula-input dimension.
	ErrDimensionMismatch = errors.New("measurement: dimension mismatch")
	// ErrUnsupportedDimension reports a derived result outside the closed model.
	ErrUnsupportedDimension = errors.New("measurement: unsupported dimension")
	// ErrAffineArithmetic reports addition/subtraction of absolute temperatures.
	ErrAffineArithmetic = errors.New("measurement: affine temperature arithmetic")
	// ErrInvalidContext reports absent or invalid conversion/rounding policy.
	ErrInvalidContext = errors.New("measurement: invalid conversion context")
	// ErrUnboundedXML reports use of encoding/xml's callback API, which cannot
	// impose a byte limit before the decoder allocates the current token.
	ErrUnboundedXML = errors.New("measurement: unbounded XML decoder")
)

// Unit is a stable catalog symbol. A cast from string does not validate it;
// Dimension and quantity constructors reject unknown symbols with ErrUnknownUnit.
// Symbols are case-sensitive ASCII wire identifiers, not localized typography.
type Unit string

// String returns the stored symbol without validating or normalizing it.
func (u Unit) String() string { return string(u) }

// Supported units carry stable symbols and dimension identity.
const (
	// One is the canonical dimensionless unit.
	One Unit = "1"

	// Millimetre is a length of one thousandth of a metre.
	Millimetre Unit = "mm"
	// Centimetre is a length of one hundredth of a metre.
	Centimetre Unit = "cm"
	// Metre is the canonical length unit.
	Metre Unit = "m"
	// Kilometre is a length of one thousand metres.
	Kilometre Unit = "km"
	// Inch is the international inch, exactly 0.0254 metres.
	Inch Unit = "in"
	// Foot is the international foot, exactly 0.3048 metres.
	Foot Unit = "ft"
	// Yard is the international yard, exactly 0.9144 metres.
	Yard Unit = "yd"

	// SquareMillimetre is area measured in squared millimetres.
	SquareMillimetre Unit = "mm2"
	// SquareCentimetre is area measured in squared centimetres.
	SquareCentimetre Unit = "cm2"
	// SquareMetre is the canonical area unit.
	SquareMetre Unit = "m2"
	// SquareInch is area measured in squared international inches.
	SquareInch Unit = "in2"
	// SquareFoot is area measured in squared international feet.
	SquareFoot Unit = "ft2"

	// CubicMillimetre is volume measured in cubed millimetres.
	CubicMillimetre Unit = "mm3"
	// CubicCentimetre is volume measured in cubed centimetres.
	CubicCentimetre Unit = "cm3"
	// CubicMetre is the canonical volume unit.
	CubicMetre Unit = "m3"
	// Millilitre is exactly one cubic centimetre.
	Millilitre Unit = "mL"
	// Litre is exactly one thousandth of a cubic metre.
	Litre Unit = "L"
	// CubicInch is volume measured in cubed international inches.
	CubicInch Unit = "in3"
	// CubicFoot is volume measured in cubed international feet.
	CubicFoot Unit = "ft3"

	// Milligram is mass measured in millionths of a kilogram.
	Milligram Unit = "mg"
	// Gram is mass measured in thousandths of a kilogram.
	Gram Unit = "g"
	// Kilogram is the canonical mass unit.
	Kilogram Unit = "kg"
	// Tonne is exactly one thousand kilograms.
	Tonne Unit = "t"
	// Ounce is the avoirdupois ounce, exactly 0.028349523125 kilograms.
	Ounce Unit = "oz"
	// Pound is the avoirdupois pound, exactly 0.45359237 kilograms.
	Pound Unit = "lb"

	// Kelvin is the canonical absolute-temperature unit.
	Kelvin Unit = "K"
	// Celsius is an absolute scale with kelvin = value + 273.15.
	Celsius Unit = "degC"
	// Fahrenheit is an absolute scale with kelvin = (value + 459.67) * 5/9.
	Fahrenheit Unit = "degF"

	// KilogramPerCubicMetre is the canonical density unit.
	KilogramPerCubicMetre Unit = "kg/m3"
	// GramPerCubicCentimetre is density equal to 1000 kilograms per cubic metre.
	GramPerCubicCentimetre Unit = "g/cm3"

	// LoadingMetre is the semantic loading-floor unit, not ordinary Metre.
	LoadingMetre Unit = "ldm"
)

type unitDefinition struct {
	dimension   Dimension
	numerator   decimal.Decimal
	denominator decimal.Decimal
	preOffset   decimal.Decimal
}

// Exact source factors follow BIPM SI Brochure 9th edition v4.01 (sections
// 2.3-2.5 and 4-5) and NIST SP 811 (2008), Appendix B. Fahrenheit is represented
// as (F + 459.67) * 5/9, not a rounded decimal approximation of 5/9. The
// finite catalog is package policy; unit_definitions_test.go protects every
// factor under MEASUREMENT-DEC-001/002 in docs/specification-decisions.md.
var unitDefinitions = map[Unit]unitDefinition{
	One: {Dimensionless, decimal.New(1), decimal.New(1), decimal.New(0)},

	Millimetre: {LengthDimension, decimal.New(1), decimal.New(1000), decimal.New(0)},
	Centimetre: {LengthDimension, decimal.New(1), decimal.New(100), decimal.New(0)},
	Metre:      {LengthDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	Kilometre:  {LengthDimension, decimal.New(1000), decimal.New(1), decimal.New(0)},
	Inch:       {LengthDimension, decimal.MustParse("0.0254"), decimal.New(1), decimal.New(0)},
	Foot:       {LengthDimension, decimal.MustParse("0.3048"), decimal.New(1), decimal.New(0)},
	Yard:       {LengthDimension, decimal.MustParse("0.9144"), decimal.New(1), decimal.New(0)},

	SquareMillimetre: {AreaDimension, decimal.New(1), decimal.New(1_000_000), decimal.New(0)},
	SquareCentimetre: {AreaDimension, decimal.New(1), decimal.New(10_000), decimal.New(0)},
	SquareMetre:      {AreaDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	SquareInch:       {AreaDimension, decimal.MustParse("0.00064516"), decimal.New(1), decimal.New(0)},
	SquareFoot:       {AreaDimension, decimal.MustParse("0.09290304"), decimal.New(1), decimal.New(0)},

	CubicMillimetre: {VolumeDimension, decimal.New(1), decimal.New(1_000_000_000), decimal.New(0)},
	CubicCentimetre: {VolumeDimension, decimal.New(1), decimal.New(1_000_000), decimal.New(0)},
	CubicMetre:      {VolumeDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	Millilitre:      {VolumeDimension, decimal.New(1), decimal.New(1_000_000), decimal.New(0)},
	Litre:           {VolumeDimension, decimal.New(1), decimal.New(1000), decimal.New(0)},
	CubicInch:       {VolumeDimension, decimal.MustParse("0.000016387064"), decimal.New(1), decimal.New(0)},
	CubicFoot:       {VolumeDimension, decimal.MustParse("0.028316846592"), decimal.New(1), decimal.New(0)},

	Milligram: {MassDimension, decimal.New(1), decimal.New(1_000_000), decimal.New(0)},
	Gram:      {MassDimension, decimal.New(1), decimal.New(1000), decimal.New(0)},
	Kilogram:  {MassDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	Tonne:     {MassDimension, decimal.New(1000), decimal.New(1), decimal.New(0)},
	Ounce:     {MassDimension, decimal.MustParse("0.028349523125"), decimal.New(1), decimal.New(0)},
	Pound:     {MassDimension, decimal.MustParse("0.45359237"), decimal.New(1), decimal.New(0)},

	Kelvin:     {TemperatureDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	Celsius:    {TemperatureDimension, decimal.New(1), decimal.New(1), decimal.MustParse("273.15")},
	Fahrenheit: {TemperatureDimension, decimal.New(5), decimal.New(9), decimal.MustParse("459.67")},

	KilogramPerCubicMetre:  {DensityDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
	GramPerCubicCentimetre: {DensityDimension, decimal.New(1000), decimal.New(1), decimal.New(0)},

	LoadingMetre: {LoadingMetreDimension, decimal.New(1), decimal.New(1), decimal.New(0)},
}

var canonicalUnits = map[Dimension]Unit{
	Dimensionless:         One,
	LengthDimension:       Metre,
	AreaDimension:         SquareMetre,
	VolumeDimension:       CubicMetre,
	MassDimension:         Kilogram,
	TemperatureDimension:  Kelvin,
	DensityDimension:      KilogramPerCubicMetre,
	LoadingMetreDimension: LoadingMetre,
}

// Dimension reports the unit's physical dimension.
func (u Unit) Dimension() (Dimension, error) {
	definition, ok := unitDefinitions[u]
	if !ok {
		return Dimensionless, fmt.Errorf("%w: unsupported unit", ErrUnknownUnit)
	}

	return definition.dimension, nil
}

func definitionFor(unit Unit) (unitDefinition, error) {
	definition, ok := unitDefinitions[unit]
	if !ok {
		return unitDefinition{}, fmt.Errorf("%w: unsupported unit", ErrUnknownUnit)
	}

	return definition, nil
}

// Units returns a caller-owned slice of supported symbols sorted lexically for
// dimension, or nil for an unsupported dimension. Mutating it cannot alter
// catalog policy.
func Units(dimension Dimension) []Unit {
	if _, ok := canonicalUnits[dimension]; !ok {
		return nil
	}
	units := make([]Unit, 0)
	for unit, definition := range unitDefinitions {
		if definition.dimension == dimension {
			units = append(units, unit)
		}
	}
	slices.Sort(units)

	return units
}
