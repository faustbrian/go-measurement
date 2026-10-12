package measurement

import (
	"context"
	"fmt"

	gomath "github.com/faustbrian/go-math"
	"github.com/faustbrian/go-math/decimal"
)

// ConversionContext makes exact versus rounded conversion an explicit caller
// choice. It is an immutable arithmetic policy, not an operation context;
// values can be reused concurrently. Its zero value is invalid. Rounding applies
// at each conversion/division boundary, not once across a composed formula.
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

// ExactConversion returns a policy using go-math defaults that rejects
// non-terminating base-ten quotients with go-math.ErrConversion. Terminating
// results can still exceed resource limits; exactness is not unbounded precision.
func ExactConversion() ConversionContext {
	return ConversionContext{mode: conversionExact, limits: gomath.DefaultLimits()}
}

// RoundedConversion returns a context that rounds the combined final ratio at
// scale fractional places, before subtracting a target affine offset. Negative
// scales round to powers of ten. Validation occurs when used: invalid scales,
// rounding modes or limits return ErrInvalidContext. Same-unit conversion is
// a no-op and does not quantize; use Quantity.Round for final display scale.
func RoundedConversion(scale int32, mode decimal.RoundingMode) ConversionContext {
	return ConversionContext{
		mode:     conversionRounded,
		scale:    scale,
		rounding: mode,
		limits:   gomath.DefaultLimits(),
	}
}

// WithLimits returns a copy using explicit arithmetic resource limits. Zero
// limits select go-math defaults; nonzero limits are validated when used. New
// and explicit Quantity.Round still use their own default admission limits.
func (c ConversionContext) WithLimits(limits gomath.Limits) ConversionContext {
	c.limits = limits

	return c
}

// Quantity is a decimal amount with explicit unit identity. Value operations
// return new values and can run concurrently. Pointer decoder methods mutate
// the receiver and require caller synchronization. The zero value has no unit;
// construct values with New or Parse. Numeric zero/negative amounts are allowed;
// physical validity and business positivity are not inferred from a unit.
type Quantity struct {
	amount decimal.Decimal
	unit   Unit
}

// New validates unit and default go-math amount limits, preserving represented
// scale without rounding or normalization. Unknown units match ErrUnknownUnit;
// amount-limit failures match ErrInvalidQuantity and the go-math cause. It does
// not reject signed/zero amounts or temperatures below absolute zero.
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

// MustNew constructs a trusted quantity and panics with any error returned by
// New, including unknown units and amount resource limits. Use New for input
// requiring returned validation errors.
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

// String returns represented decimal text, one space and the stored symbol.
// It does not validate the unit; the absent zero value therefore renders "0 ".
func (q Quantity) String() string { return q.amount.String() + " " + string(q.unit) }

// Convert converts q into target without caller cancellation.
//
// Use ConvertContext when the operation belongs to a caller-owned context.
func (q Quantity) Convert(target Unit, conversion ConversionContext) (Quantity, error) {
	return q.ConvertContext(context.Background(), target, conversion)
}

// ConvertContext converts q into target under caller cancellation and an
// explicit conversion policy. Context admission precedes policy/unit validation:
// nil ctx wraps go-math.ErrInvalidArgument, while observed cancellation/deadline
// errors are returned unchanged. Unsupported units and dimension mismatches fail
// without a partial result. A same-unit target returns q after policy validation,
// without quantizing or rechecking amount against caller limits. Other conversions
// round only the combined quotient; target affine-offset subtraction follows
// that boundary and can leave a different represented scale.
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

	// MEASUREMENT-DEC-003: keep source/target factors in one quotient. Rounding
	// after the source ratio but before applying the target ratio loses precision;
	// the target affine offset is subtracted only after quotient evaluation.
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
// caller cancellation. The conversion may round other first; addition itself
// is exact and retains q's unit. Absolute temperatures return ErrAffineArithmetic.
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
// difference under caller cancellation. Other is converted (and may be rounded)
// first; subtraction is exact and retains q's unit. Absolute temperatures return
// ErrAffineArithmetic.
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
// under caller cancellation. It returns -1, 0 or +1; the comparison reflects
// any rounding used to convert other and is meaningful only when err is nil.
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
// conversion under caller cancellation. Rounded conversion can equate distinct
// source amounts; use exact conversion when equality must not discard precision.
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
// cancellation. Operands are converted separately to canonical units using
// conversion, then multiplied exactly. Rounded conversions therefore need not
// preserve exact algebraic identities. Unmodeled results return ErrUnsupportedDimension.
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
// cancellation. Operands are separately converted to canonical units before the
// final quotient uses conversion. Division by zero matches go-math.ErrDivisionByZero;
// unmodeled dimensions return ErrUnsupportedDimension.
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
// cancellation. Negative scales round to powers of ten; scale/mode and resource
// errors come from go-math. This only quantizes the amount and retains the stored
// unit without validating it, including an absent zero-value unit.
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
// caller cancellation. Count must be in [1, MaxPackageQuantity] or return
// ErrInvalidQuantity. It retains the stored unit without validating it and uses
// default arithmetic limits, independent of a ConversionContext.
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
// cancellation. Both bounds are converted to q's unit first, potentially with
// rounding; a reversed converted interval returns decimal.ErrInvalid.
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
