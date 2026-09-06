package measurement

import (
	"context"
	"fmt"
	"strconv"

	"github.com/faustbrian/go-math/decimal"
)

// MaxPackageQuantity bounds shipment formula work and integer conversion.
const MaxPackageQuantity uint64 = 1_000_000

// Dimensions is an immutable rectangular package triple and package count.
type Dimensions struct {
	length   Quantity
	width    Quantity
	height   Quantity
	quantity uint64
}

// NewDimensions validates a positive rectangular triple and bounded count.
func NewDimensions(length, width, height Quantity, quantity uint64) (Dimensions, error) {
	for name, value := range map[string]Quantity{
		"length": length,
		"width":  width,
		"height": height,
	} {
		dimension, err := value.Dimension()
		if err != nil {
			return Dimensions{}, err
		}
		if dimension != LengthDimension {
			return Dimensions{}, fmt.Errorf("%w: %s must be length", ErrDimensionMismatch, name)
		}
		if value.amount.Sign() <= 0 {
			return Dimensions{}, fmt.Errorf("%w: %s must be positive", ErrInvalidQuantity, name)
		}
	}
	if quantity == 0 || quantity > MaxPackageQuantity {
		return Dimensions{}, fmt.Errorf("%w: package quantity must be in [1,%d]", ErrInvalidQuantity, MaxPackageQuantity)
	}

	return Dimensions{length: length, width: width, height: height, quantity: quantity}, nil
}

// Length returns the original immutable length quantity.
func (d Dimensions) Length() Quantity { return d.length }

// Width returns the original immutable width quantity.
func (d Dimensions) Width() Quantity { return d.width }

// Height returns the original immutable height quantity.
func (d Dimensions) Height() Quantity { return d.height }

// Quantity returns the package count.
func (d Dimensions) Quantity() uint64 { return d.quantity }

// FloorArea returns length multiplied by width in target without caller
// cancellation.
func (d Dimensions) FloorArea(target Unit, conversion ConversionContext) (Quantity, error) {
	return d.FloorAreaContext(context.Background(), target, conversion)
}

// FloorAreaContext returns length multiplied by width in target under caller
// cancellation.
func (d Dimensions) FloorAreaContext(
	ctx context.Context,
	target Unit,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	area, err := d.length.MultiplyContext(ctx, d.width, conversion)
	if err != nil {
		return Quantity{}, err
	}

	return area.ConvertContext(ctx, target, conversion)
}

// CubicVolume returns length multiplied by width and height in target without
// caller cancellation.
func (d Dimensions) CubicVolume(target Unit, conversion ConversionContext) (Quantity, error) {
	return d.CubicVolumeContext(context.Background(), target, conversion)
}

// CubicVolumeContext returns length multiplied by width and height in target
// under caller cancellation.
func (d Dimensions) CubicVolumeContext(
	ctx context.Context,
	target Unit,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	area, err := d.length.MultiplyContext(ctx, d.width, conversion)
	if err != nil {
		return Quantity{}, err
	}
	volume, err := area.MultiplyContext(ctx, d.height, conversion)
	if err != nil {
		return Quantity{}, err
	}

	return volume.ConvertContext(ctx, target, conversion)
}

// TotalVolume multiplies one package's cubic volume by the package count
// without caller cancellation.
func (d Dimensions) TotalVolume(target Unit, conversion ConversionContext) (Quantity, error) {
	return d.TotalVolumeContext(context.Background(), target, conversion)
}

// TotalVolumeContext multiplies one package's cubic volume by the package
// count under caller cancellation.
func (d Dimensions) TotalVolumeContext(
	ctx context.Context,
	target Unit,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	volume, err := d.CubicVolumeContext(ctx, target, conversion)
	if err != nil {
		return Quantity{}, err
	}
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}

	amount, err := conversion.multiplyContext(ctx, volume.amount, decimal.MustParse(strconv.FormatUint(d.quantity, 10)))
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: target}, nil
}

// TruckWidth is a validated positive vehicle interior width.
type TruckWidth struct{ width Quantity }

// NewTruckWidth validates a positive length quantity.
func NewTruckWidth(width Quantity) (TruckWidth, error) {
	dimension, err := width.Dimension()
	if err != nil {
		return TruckWidth{}, err
	}
	if dimension != LengthDimension {
		return TruckWidth{}, fmt.Errorf("%w: truck width must be length", ErrDimensionMismatch)
	}
	if width.amount.Sign() <= 0 {
		return TruckWidth{}, fmt.Errorf("%w: truck width must be positive", ErrInvalidQuantity)
	}

	return TruckWidth{width: width}, nil
}

// Quantity returns the configured width quantity.
func (t TruckWidth) Quantity() Quantity { return t.width }

// StackingFactor is the positive number of equivalent floor positions shared
// by stacked packages. One means not stackable.
type StackingFactor struct{ factor decimal.Decimal }

// NewStackingFactor validates a positive stacking divisor.
func NewStackingFactor(factor decimal.Decimal) (StackingFactor, error) {
	if err := validateAmount(factor); err != nil {
		return StackingFactor{}, err
	}
	if factor.Sign() <= 0 {
		return StackingFactor{}, fmt.Errorf("%w: stacking factor must be positive", ErrInvalidQuantity)
	}

	return StackingFactor{factor: factor}, nil
}

// Decimal returns the exact stacking factor.
func (s StackingFactor) Decimal() decimal.Decimal { return s.factor }

// LoadingMetres calculates floor area / truck width / stacking factor and
// multiplies by package quantity. The result has semantic loading-metre
// identity and therefore cannot be added to ordinary lengths accidentally.
func (d Dimensions) LoadingMetres(
	truckWidth TruckWidth,
	stacking StackingFactor,
	conversion ConversionContext,
) (Quantity, error) {
	return d.LoadingMetresContext(context.Background(), truckWidth, stacking, conversion)
}

// LoadingMetresContext calculates loading metres under caller cancellation.
func (d Dimensions) LoadingMetresContext(
	ctx context.Context,
	truckWidth TruckWidth,
	stacking StackingFactor,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	area, err := d.FloorAreaContext(ctx, SquareMetre, conversion)
	if err != nil {
		return Quantity{}, err
	}
	width, err := truckWidth.width.ConvertContext(ctx, Metre, conversion)
	if err != nil {
		return Quantity{}, err
	}
	length, err := area.DivideContext(ctx, width, conversion)
	if err != nil {
		return Quantity{}, err
	}
	perPackage, err := conversion.divideContext(ctx, length.amount, stacking.factor)
	if err != nil {
		return Quantity{}, err
	}
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}

	amount, err := conversion.multiplyContext(ctx, perPackage, decimal.MustParse(strconv.FormatUint(d.quantity, 10)))
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: amount, unit: LoadingMetre}, nil
}

// VolumetricDivisor is a positive volume-per-kilogram carrier input with an
// explicit volume unit, for example 5000 cm3/kg.
type VolumetricDivisor struct {
	volumePerKilogram decimal.Decimal
	volumeUnit        Unit
}

// NewVolumetricDivisor validates a positive volume-per-kilogram divisor.
func NewVolumetricDivisor(volumePerKilogram decimal.Decimal, volumeUnit Unit) (VolumetricDivisor, error) {
	dimension, err := volumeUnit.Dimension()
	if err != nil {
		return VolumetricDivisor{}, err
	}
	if dimension != VolumeDimension {
		return VolumetricDivisor{}, fmt.Errorf("%w: divisor unit must be volume", ErrDimensionMismatch)
	}
	if err := validateAmount(volumePerKilogram); err != nil {
		return VolumetricDivisor{}, err
	}
	if volumePerKilogram.Sign() <= 0 {
		return VolumetricDivisor{}, fmt.Errorf("%w: divisor must be positive", ErrInvalidQuantity)
	}

	return VolumetricDivisor{volumePerKilogram: volumePerKilogram, volumeUnit: volumeUnit}, nil
}

// VolumePerKilogram returns the exact divisor amount.
func (d VolumetricDivisor) VolumePerKilogram() decimal.Decimal { return d.volumePerKilogram }

// VolumeUnit returns the divisor's explicit volume unit.
func (d VolumetricDivisor) VolumeUnit() Unit { return d.volumeUnit }

// Weight calculates dimensional weight in kilograms without caller
// cancellation.
func (d VolumetricDivisor) Weight(volume Quantity, conversion ConversionContext) (Quantity, error) {
	return d.WeightContext(context.Background(), volume, conversion)
}

// WeightContext calculates dimensional weight in kilograms under caller
// cancellation.
func (d VolumetricDivisor) WeightContext(
	ctx context.Context,
	volume Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	dimension, err := volume.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	if dimension != VolumeDimension {
		return Quantity{}, fmt.Errorf("%w: volumetric input must be volume", ErrDimensionMismatch)
	}
	converted, err := volume.ConvertContext(ctx, d.volumeUnit, conversion)
	if err != nil {
		return Quantity{}, err
	}
	weight, err := conversion.divideContext(ctx, converted.amount, d.volumePerKilogram)
	if err != nil {
		return Quantity{}, err
	}

	return Quantity{amount: weight, unit: Kilogram}, nil
}

// VolumetricIndex is a positive mass-per-volume factor, usually expressed as
// kilograms per cubic metre.
type VolumetricIndex struct{ density Quantity }

// NewVolumetricIndex validates a positive density quantity.
func NewVolumetricIndex(density Quantity) (VolumetricIndex, error) {
	dimension, err := density.Dimension()
	if err != nil {
		return VolumetricIndex{}, err
	}
	if dimension != DensityDimension {
		return VolumetricIndex{}, fmt.Errorf("%w: volumetric index must be density", ErrDimensionMismatch)
	}
	if density.amount.Sign() <= 0 {
		return VolumetricIndex{}, fmt.Errorf("%w: volumetric index must be positive", ErrInvalidQuantity)
	}

	return VolumetricIndex{density: density}, nil
}

// Density returns the configured mass-per-volume quantity.
func (i VolumetricIndex) Density() Quantity { return i.density }

// Weight multiplies the index by a volume and returns canonical kilograms
// without caller cancellation.
func (i VolumetricIndex) Weight(volume Quantity, conversion ConversionContext) (Quantity, error) {
	return i.WeightContext(context.Background(), volume, conversion)
}

// WeightContext multiplies the index by a volume and returns canonical
// kilograms under caller cancellation.
func (i VolumetricIndex) WeightContext(
	ctx context.Context,
	volume Quantity,
	conversion ConversionContext,
) (Quantity, error) {
	if err := validateOperationContext(ctx); err != nil {
		return Quantity{}, err
	}
	dimension, err := volume.Dimension()
	if err != nil {
		return Quantity{}, err
	}
	if dimension != VolumeDimension {
		return Quantity{}, fmt.Errorf("%w: volumetric input must be volume", ErrDimensionMismatch)
	}
	weight, err := i.density.MultiplyContext(ctx, volume, conversion)
	if err != nil {
		return Quantity{}, err
	}

	return weight.ConvertContext(ctx, Kilogram, conversion)
}
