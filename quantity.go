package measurement

import (
	"context"
	"fmt"

	gomath "github.com/faustbrian/go-math"
	"github.com/faustbrian/go-math/decimal"
)

// ConversionContext makes exact versus rounded conversion an explicit caller
// choice.
type ConversionContext struct {
	mode     conversionMode
	scale    int32
	rounding decimal.RoundingMode
	limits   gomath.Limits
}

type conversionMode uint8

const (
	conversionUnset conversionMode = iota
	conversionExact
	conversionRounded
)

// ExactConversion returns a context that rejects non-terminating quotients.
func ExactConversion() ConversionContext {
	return ConversionContext{mode: conversionExact, limits: gomath.DefaultLimits()}
}

// RoundedConversion returns a context that rounds the combined final ratio at
// scale fractional places.
func RoundedConversion(scale int32, mode decimal.RoundingMode) ConversionContext {
	return ConversionContext{
		mode:     conversionRounded,
		scale:    scale,
		rounding: mode,
		limits:   gomath.DefaultLimits(),
	}
}

// WithLimits returns a copy using explicit arithmetic resource limits.
func (c ConversionContext) WithLimits(limits gomath.Limits) ConversionContext {
	c.limits = limits

	return c
}

// Quantity is an immutable decimal amount with explicit unit identity.
type Quantity struct {
	amount decimal.Decimal
	unit   Unit
}

// New validates unit and constructs an immutable quantity.
func New(amount decimal.Decimal, unit Unit) (Quantity, error) {
	if _, err := definitionFor(unit); err != nil {
		return Quantity{}, err
	}
	if err := validateAmount(amount); err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: unit}, nil
}

func validateAmount(amount decimal.Decimal) error {
	if _, err := decimal.FromBig(amount.Coefficient(), amount.Exponent(), gomath.DefaultLimits()); err != nil {
		return fmt.Errorf("%w: amount: %w", ErrInvalidQuantity, err)
	}

	return nil
}

// MustNew constructs a trusted quantity and panics for an unknown unit.
func MustNew(amount decimal.Decimal, unit Unit) Quantity {
	quantity, err := New(amount, unit)
	if err != nil {
		panic(err)
	}

	return quantity
}

// Amount returns the immutable decimal amount.
func (q Quantity) Amount() decimal.Decimal { return q.amount }

// Unit returns the quantity's unit identity.
func (q Quantity) Unit() Unit { return q.unit }

// Dimension returns the unit's physical dimension.
func (q Quantity) Dimension() (Dimension, error) { return q.unit.Dimension() }

func (q Quantity) String() string { return q.amount.String() + " " + string(q.unit) }

// Convert converts q into target without caller cancellation.
//
// Use ConvertContext when the operation belongs to a caller-owned context.
func (q Quantity) Convert(target Unit, conversion ConversionContext) (Quantity, error) {
	return q.ConvertContext(context.Background(), target, conversion)
}

// ConvertContext converts q into target under caller cancellation and an
// explicit conversion policy.
func (q Quantity) ConvertContext(
	ctx context.Context,
	target Unit,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	if err := conversion.validate(); err != nil {
		return Quantity{}, err
	}
	from, err := definitionFor(q.unit)
	if err != nil {
		return Quantity{}, err
	}
	to, err := definitionFor(target)
	if err != nil {
		return Quantity{}, err
	}
	if from.dimension != to.dimension {
		return Quantity{}, fmt.Errorf("%w: %s and %s", ErrDimensionMismatch, from.dimension, to.dimension)
	}
	if q.unit == target {
		return q, nil
	}

	offsetAmount, err := conversion.addContext(ctx, q.amount, from.preOffset)
	if err != nil {
		return Quantity{}, err
	}
	numerator, err := conversion.multiplyContext(ctx, offsetAmount, from.numerator)
	if err != nil {
		return Quantity{}, err
	}
	numerator, err = conversion.multiplyContext(ctx, numerator, to.denominator)
	if err != nil {
		return Quantity{}, err
	}
	denominator, err := conversion.multiplyContext(ctx, from.denominator, to.numerator)
	if err != nil {
		return Quantity{}, err
	}
	targetAmount, err := conversion.divideContext(ctx, numerator, denominator)
	if err != nil {
		// Context sentinels have an exact-identity contract at this boundary.
		if err == context.Canceled || err == context.DeadlineExceeded { //nolint:errorlint
			return Quantity{}, err
		}
		return Quantity{}, fmt.Errorf("convert %s to %s: %w", q.unit, target, err)
	}

	amount, err := conversion.subtractContext(ctx, targetAmount, to.preOffset)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: target}, nil
}

// Add converts other into q's unit and returns their exact sum without caller
// cancellation.
func (q Quantity) Add(other Quantity, conversion ConversionContext) (Quantity, error) {
	return q.AddContext(context.Background(), other, conversion)
}

// AddContext converts other into q's unit and returns their exact sum under
// caller cancellation.
func (q Quantity) AddContext(
	ctx context.Context,
	other Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	if dimension, err := q.Dimension(); err != nil {
		return Quantity{}, err
	} else if dimension == TemperatureDimension {
		return Quantity{}, ErrAffineArithmetic
	}
	converted, err := q.compatibleContext(ctx, other, conversion)
	if err != nil {
		return Quantity{}, err
	}

	amount, err := conversion.addContext(ctx, q.amount, converted.amount)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: q.unit}, nil
}

// Subtract converts other into q's unit and returns their exact difference
// without caller cancellation.
func (q Quantity) Subtract(other Quantity, conversion ConversionContext) (Quantity, error) {
	return q.SubtractContext(context.Background(), other, conversion)
}

// SubtractContext converts other into q's unit and returns their exact
// difference under caller cancellation.
func (q Quantity) SubtractContext(
	ctx context.Context,
	other Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	if dimension, err := q.Dimension(); err != nil {
		return Quantity{}, err
	} else if dimension == TemperatureDimension {
		return Quantity{}, ErrAffineArithmetic
	}
	converted, err := q.compatibleContext(ctx, other, conversion)
	if err != nil {
		return Quantity{}, err
	}

	amount, err := conversion.subtractContext(ctx, q.amount, converted.amount)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: q.unit}, nil
}

// Compare converts other into q's unit and compares numeric amounts without
// caller cancellation.
func (q Quantity) Compare(other Quantity, conversion ConversionContext) (int, error) {
	return q.CompareContext(context.Background(), other, conversion)
}

// CompareContext converts other into q's unit and compares numeric amounts
// under caller cancellation.
func (q Quantity) CompareContext(ctx context.Context, other Quantity, conversion ConversionContext) (int, error) {
	if err := validateOperationContext(ctx); err != nil {
		return 0, err
	}
	converted, err := q.compatibleContext(ctx, other, conversion)
	if err != nil {
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	return q.amount.Cmp(converted.amount), nil
}

// Equal reports numeric equality without caller cancellation.
func (q Quantity) Equal(other Quantity, conversion ConversionContext) (bool, error) {
	return q.EqualContext(context.Background(), other, conversion)
}

// EqualContext reports numeric equality after explicit compatible-unit
// conversion under caller cancellation.
func (q Quantity) EqualContext(ctx context.Context, other Quantity, conversion ConversionContext) (bool, error) {
	if err := validateOperationContext(ctx); err != nil {
		return false, err
	}
	comparison, err := q.CompareContext(ctx, other, conversion)
	if err != nil {
		return false, err
	}

	return comparison == 0, nil
}

// Multiply returns a supported canonical derived quantity without caller
// cancellation.
func (q Quantity) Multiply(other Quantity, conversion ConversionContext) (Quantity, error) {
	return q.MultiplyContext(context.Background(), other, conversion)
}

// MultiplyContext returns a supported canonical derived quantity under caller
// cancellation.
func (q Quantity) MultiplyContext(
	ctx context.Context,
	other Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	leftDimension, err := q.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	rightDimension, err := other.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	resultDimension, err := leftDimension.multiply(rightDimension)
	if err != nil {
		return Quantity{}, err
	}
	resultUnit := canonicalUnits[resultDimension]

	left, err := q.ConvertContext(ctx, canonicalUnits[leftDimension], conversion)
	if err != nil {
		return Quantity{}, err
	}
	right, err := other.ConvertContext(ctx, canonicalUnits[rightDimension], conversion)
	if err != nil {
		return Quantity{}, err
	}

	amount, err := conversion.multiplyContext(ctx, left.amount, right.amount)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: resultUnit}, nil
}

// Divide returns a supported canonical derived quantity without caller
// cancellation.
func (q Quantity) Divide(other Quantity, conversion ConversionContext) (Quantity, error) {
	return q.DivideContext(context.Background(), other, conversion)
}

// DivideContext returns a supported canonical derived quantity under caller
// cancellation.
func (q Quantity) DivideContext(
	ctx context.Context,
	other Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	leftDimension, err := q.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	rightDimension, err := other.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	resultDimension, err := leftDimension.divide(rightDimension)
	if err != nil {
		return Quantity{}, err
	}
	resultUnit := canonicalUnits[resultDimension]

	left, err := q.ConvertContext(ctx, canonicalUnits[leftDimension], conversion)
	if err != nil {
		return Quantity{}, err
	}
	right, err := other.ConvertContext(ctx, canonicalUnits[rightDimension], conversion)
	if err != nil {
		return Quantity{}, err
	}
	amount, err := conversion.divideContext(ctx, left.amount, right.amount)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: resultUnit}, nil
}

// Round quantizes q at scale fractional places without caller cancellation.
func (q Quantity) Round(scale int32, mode decimal.RoundingMode) (Quantity, error) {
	return q.RoundContext(context.Background(), scale, mode)
}

// RoundContext quantizes q at scale fractional places under caller
// cancellation.
func (q Quantity) RoundContext(
	ctx context.Context,
	scale int32,
	mode decimal.RoundingMode,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	result, err := q.amount.Quantize(ctx, scale, mode, gomath.DefaultLimits())
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: result.Value, unit: q.unit}, nil
}

// Times multiplies an amount by a bounded positive package count without
// caller cancellation.
func (q Quantity) Times(count uint64) (Quantity, error) {
	return q.TimesContext(context.Background(), count)
}

// TimesContext multiplies an amount by a bounded positive package count under
// caller cancellation.
func (q Quantity) TimesContext(ctx context.Context, count uint64) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	if count == 0 || count > MaxPackageQuantity {
		return Quantity{}, fmt.Errorf("%w: count must be in [1,%d]", ErrInvalidQuantity, MaxPackageQuantity)
	}
	amount, err := ExactConversion().multiplyContext(ctx, q.amount, decimal.New(int64(count)))
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: q.unit}, nil
}

// Clamp restricts q to an inclusive compatible interval without caller
// cancellation.
func (q Quantity) Clamp(minimum, maximum Quantity, conversion ConversionContext) (Quantity, error) {
	return q.ClampContext(context.Background(), minimum, maximum, conversion)
}

// ClampContext restricts q to an inclusive compatible interval under caller
// cancellation.
func (q Quantity) ClampContext(
	ctx context.Context,
	minimum, maximum Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	convertedMinimum, err := q.compatibleContext(ctx, minimum, conversion)
	if err != nil {
		return Quantity{}, err
	}
	convertedMaximum, err := q.compatibleContext(ctx, maximum, conversion)
	if err != nil {
		return Quantity{}, err
	}
	if err := ctx.Err(); err != nil {
		return Quantity{}, err
	}
	amount, err := q.amount.Clamp(convertedMinimum.amount, convertedMaximum.amount)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: q.unit}, nil
}

func (q Quantity) compatibleContext(
	ctx context.Context,
	other Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	leftDimension, err := q.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	rightDimension, err := other.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	if leftDimension != rightDimension {
		return Quantity{}, fmt.Errorf("%w: %s and %s", ErrDimensionMismatch, leftDimension, rightDimension)
	}

	return other.ConvertContext(ctx, q.unit, conversion)
}

func (c ConversionContext) divide(numerator, denominator decimal.Decimal) (decimal.Decimal, error) {
	return c.divideContext(context.Background(), numerator, denominator)
}

func (c ConversionContext) divideContext(
	ctx context.Context,
	numerator, denominator decimal.Decimal,
) (decimal.Decimal, error) {
	if err := c.validate(); err != nil {
		return decimal.Decimal{}, err
	}
	limits := c.arithmeticLimits()
	if c.mode == conversionExact {
		return numerator.QuoExact(ctx, denominator, limits)
	}
	result, err := decimal.QuantizedQuo(
		ctx,
		numerator,
		denominator,
		c.scale,
		c.rounding,
		limits,
	)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return result.Value, nil
}

func (c ConversionContext) add(left, right decimal.Decimal) (decimal.Decimal, error) {
	return c.addContext(context.Background(), left, right)
}

func (c ConversionContext) addContext(ctx context.Context, left, right decimal.Decimal) (decimal.Decimal, error) {
	if err := c.validate(); err != nil {
		return decimal.Decimal{}, err
	}
	return left.AddExact(ctx, right, c.arithmeticLimits())
}

func (c ConversionContext) subtract(left, right decimal.Decimal) (decimal.Decimal, error) {
	return c.subtractContext(context.Background(), left, right)
}

func (c ConversionContext) subtractContext(ctx context.Context, left, right decimal.Decimal) (decimal.Decimal, error) {
	if err := c.validate(); err != nil {
		return decimal.Decimal{}, err
	}
	return left.SubExact(ctx, right, c.arithmeticLimits())
}

func (c ConversionContext) multiply(left, right decimal.Decimal) (decimal.Decimal, error) {
	return c.multiplyContext(context.Background(), left, right)
}

func (c ConversionContext) multiplyContext(ctx context.Context, left, right decimal.Decimal) (decimal.Decimal, error) {
	if err := c.validate(); err != nil {
		return decimal.Decimal{}, err
	}
	return left.MulExact(ctx, right, c.arithmeticLimits())
}

func validateOperationContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("%w: nil context", gomath.ErrInvalidArgument)
	}

	return ctx.Err()
}

func (c ConversionContext) validate() error {
	if c.mode == conversionUnset || c.mode > conversionRounded {
		return ErrInvalidContext
	}
	limits := c.arithmeticLimits()
	if err := limits.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContext, err)
	}
	if c.mode == conversionRounded && (!c.rounding.Valid() ||
		int64(c.scale) > int64(limits.MaxExponentMagnitude) ||
		int64(c.scale) < -int64(limits.MaxExponentMagnitude)) {
		return ErrInvalidContext
	}

	return nil
}

func (c ConversionContext) arithmeticLimits() gomath.Limits {
	if c.limits == (gomath.Limits{}) {
		return gomath.DefaultLimits()
	}

	return c.limits
}
