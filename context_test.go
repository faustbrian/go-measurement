package measurement_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	gomath "github.com/faustbrian/go-math"
	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
)

func TestContextFirstSurfaceCompiles(t *testing.T) {
	t.Parallel()

	requireSignature[func(measurement.Quantity, context.Context, measurement.Unit, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.ConvertContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.AddContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.SubtractContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (int, error)](measurement.Quantity.CompareContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (bool, error)](measurement.Quantity.EqualContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.MultiplyContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.DivideContext)
	requireSignature[func(measurement.Quantity, context.Context, int32, decimal.RoundingMode) (measurement.Quantity, error)](measurement.Quantity.RoundContext)
	requireSignature[func(measurement.Quantity, context.Context, uint64) (measurement.Quantity, error)](measurement.Quantity.TimesContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.Quantity, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Quantity.ClampContext)
	requireSignature[func(measurement.Quantity, context.Context, measurement.FormatOptions) (string, error)](measurement.Quantity.FormatContext)
	requireSignature[func(measurement.Dimensions, context.Context, measurement.Unit, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Dimensions.FloorAreaContext)
	requireSignature[func(measurement.Dimensions, context.Context, measurement.Unit, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Dimensions.CubicVolumeContext)
	requireSignature[func(measurement.Dimensions, context.Context, measurement.Unit, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Dimensions.TotalVolumeContext)
	requireSignature[func(measurement.Dimensions, context.Context, measurement.TruckWidth, measurement.StackingFactor, measurement.ConversionContext) (measurement.Quantity, error)](measurement.Dimensions.LoadingMetresContext)
	requireSignature[func(measurement.VolumetricDivisor, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.VolumetricDivisor.WeightContext)
	requireSignature[func(measurement.VolumetricIndex, context.Context, measurement.Quantity, measurement.ConversionContext) (measurement.Quantity, error)](measurement.VolumetricIndex.WeightContext)
}

func TestContextFirstMethodsCheckContextBeforeInvalidArguments(t *testing.T) {
	t.Parallel()

	quantity := measurement.MustNew(decimal.New(1), measurement.Metre)
	invalidQuantity := measurement.Quantity{}
	invalidConversion := measurement.ConversionContext{}
	invalidDimensions := measurement.Dimensions{}
	invalidTruckWidth := measurement.TruckWidth{}
	invalidStacking := measurement.StackingFactor{}
	invalidDivisor := measurement.VolumetricDivisor{}
	invalidIndex := measurement.VolumetricIndex{}
	twoMetres := measurement.MustNew(decimal.New(2), measurement.Metre)
	conversion := measurement.ExactConversion()

	operations := []struct {
		name string
		call func(context.Context) error
	}{
		{"invalid conversion", func(ctx context.Context) error {
			_, err := quantity.ConvertContext(ctx, measurement.Metre, invalidConversion)
			return err
		}},
		{"invalid unit", func(ctx context.Context) error {
			_, err := quantity.ConvertContext(ctx, measurement.Unit("invalid"), conversion)
			return err
		}},
		{"invalid quantity", func(ctx context.Context) error {
			_, err := invalidQuantity.AddContext(ctx, quantity, conversion)
			return err
		}},
		{"invalid count", func(ctx context.Context) error { _, err := quantity.TimesContext(ctx, 0); return err }},
		{"invalid clamp interval", func(ctx context.Context) error {
			_, err := quantity.ClampContext(ctx, twoMetres, quantity, conversion)
			return err
		}},
		{"invalid format options", func(ctx context.Context) error {
			_, err := quantity.FormatContext(ctx, measurement.FormatOptions{})
			return err
		}},
		{"ConvertContext", func(ctx context.Context) error {
			_, err := invalidQuantity.ConvertContext(ctx, measurement.Unit("invalid"), invalidConversion)
			return err
		}},
		{"AddContext", func(ctx context.Context) error {
			_, err := invalidQuantity.AddContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"SubtractContext", func(ctx context.Context) error {
			_, err := invalidQuantity.SubtractContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"CompareContext", func(ctx context.Context) error {
			_, err := invalidQuantity.CompareContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"EqualContext", func(ctx context.Context) error {
			_, err := invalidQuantity.EqualContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"MultiplyContext", func(ctx context.Context) error {
			_, err := invalidQuantity.MultiplyContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"DivideContext", func(ctx context.Context) error {
			_, err := invalidQuantity.DivideContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"RoundContext", func(ctx context.Context) error {
			_, err := invalidQuantity.RoundContext(ctx, 0, decimal.RoundingMode(255))
			return err
		}},
		{"TimesContext", func(ctx context.Context) error { _, err := invalidQuantity.TimesContext(ctx, 0); return err }},
		{"ClampContext", func(ctx context.Context) error {
			_, err := invalidQuantity.ClampContext(ctx, quantity, invalidQuantity, invalidConversion)
			return err
		}},
		{"FormatContext", func(ctx context.Context) error {
			_, err := invalidQuantity.FormatContext(ctx, measurement.FormatOptions{})
			return err
		}},
		{"FloorAreaContext", func(ctx context.Context) error {
			_, err := invalidDimensions.FloorAreaContext(ctx, measurement.Unit("invalid"), invalidConversion)
			return err
		}},
		{"CubicVolumeContext", func(ctx context.Context) error {
			_, err := invalidDimensions.CubicVolumeContext(ctx, measurement.Unit("invalid"), invalidConversion)
			return err
		}},
		{"TotalVolumeContext", func(ctx context.Context) error {
			_, err := invalidDimensions.TotalVolumeContext(ctx, measurement.Unit("invalid"), invalidConversion)
			return err
		}},
		{"LoadingMetresContext", func(ctx context.Context) error {
			_, err := invalidDimensions.LoadingMetresContext(ctx, invalidTruckWidth, invalidStacking, invalidConversion)
			return err
		}},
		{"VolumetricDivisor.WeightContext", func(ctx context.Context) error {
			_, err := invalidDivisor.WeightContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
		{"VolumetricIndex.WeightContext", func(ctx context.Context) error {
			_, err := invalidIndex.WeightContext(ctx, invalidQuantity, invalidConversion)
			return err
		}},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()

			if err := operation.call(nil); err == nil || err.Error() != "math: invalid argument: nil context" || !errors.Is(err, gomath.ErrInvalidArgument) {
				t.Fatalf("nil context error = %v", err)
			}

			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			// Exact error identity is part of the public context contract.
			if err := operation.call(canceled); err != context.Canceled { //nolint:errorlint
				t.Fatalf("canceled context error = %v, want context.Canceled", err)
			}

			expired, cancelDeadline := context.WithDeadline(context.Background(), time.Unix(0, 0))
			defer cancelDeadline()
			// Exact error identity is part of the public context contract.
			if err := operation.call(expired); err != context.DeadlineExceeded { //nolint:errorlint
				t.Fatalf("expired context error = %v, want context.DeadlineExceeded", err)
			}
		})
	}
}

func TestContextFirstMethodsMatchLegacyResults(t *testing.T) {
	t.Parallel()

	conversion := measurement.ExactConversion()
	oneMetre := measurement.MustNew(decimal.New(1), measurement.Metre)
	oneHundredCentimetres := measurement.MustNew(decimal.New(100), measurement.Centimetre)
	twoMetres := measurement.MustNew(decimal.New(2), measurement.Metre)
	threeSquareMetres := measurement.MustNew(decimal.New(3), measurement.SquareMetre)
	volume := measurement.MustNew(decimal.New(10_000), measurement.CubicCentimetre)
	dimensions, err := measurement.NewDimensions(oneMetre, twoMetres, measurement.MustNew(decimal.New(3), measurement.Metre), 2)
	if err != nil {
		t.Fatal(err)
	}
	truckWidth, err := measurement.NewTruckWidth(twoMetres)
	if err != nil {
		t.Fatal(err)
	}
	stacking, err := measurement.NewStackingFactor(decimal.New(2))
	if err != nil {
		t.Fatal(err)
	}
	divisor, err := measurement.NewVolumetricDivisor(decimal.New(5_000), measurement.CubicCentimetre)
	if err != nil {
		t.Fatal(err)
	}
	index, err := measurement.NewVolumetricIndex(measurement.MustNew(decimal.New(2), measurement.KilogramPerCubicMetre))
	if err != nil {
		t.Fatal(err)
	}
	format := measurement.FormatOptions{Unit: measurement.Metre, Conversion: conversion, Scale: 2, Rounding: decimal.HalfEven, Separator: " "}

	tests := []struct {
		name    string
		want    string
		legacy  func() (string, error)
		withCtx func() (string, error)
	}{
		{"Convert", "100 cm", quantityString(func() (measurement.Quantity, error) { return oneMetre.Convert(measurement.Centimetre, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return oneMetre.ConvertContext(context.Background(), measurement.Centimetre, conversion)
		})},
		{"Add", "2 m", quantityString(func() (measurement.Quantity, error) { return oneMetre.Add(oneHundredCentimetres, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return oneMetre.AddContext(context.Background(), oneHundredCentimetres, conversion)
		})},
		{"Subtract", "1 m", quantityString(func() (measurement.Quantity, error) { return twoMetres.Subtract(oneHundredCentimetres, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return twoMetres.SubtractContext(context.Background(), oneHundredCentimetres, conversion)
		})},
		{"Compare", "0", scalarString(func() (int, error) { return oneMetre.Compare(oneHundredCentimetres, conversion) }), scalarString(func() (int, error) {
			return oneMetre.CompareContext(context.Background(), oneHundredCentimetres, conversion)
		})},
		{"Equal", "true", scalarString(func() (bool, error) { return oneMetre.Equal(oneHundredCentimetres, conversion) }), scalarString(func() (bool, error) {
			return oneMetre.EqualContext(context.Background(), oneHundredCentimetres, conversion)
		})},
		{"Multiply", "2 m2", quantityString(func() (measurement.Quantity, error) { return oneMetre.Multiply(twoMetres, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return oneMetre.MultiplyContext(context.Background(), twoMetres, conversion)
		})},
		{"Divide", "3 m", quantityString(func() (measurement.Quantity, error) { return threeSquareMetres.Divide(oneMetre, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return threeSquareMetres.DivideContext(context.Background(), oneMetre, conversion)
		})},
		{"Round", "1.00 m", quantityString(func() (measurement.Quantity, error) { return oneMetre.Round(2, decimal.HalfEven) }), quantityString(func() (measurement.Quantity, error) {
			return oneMetre.RoundContext(context.Background(), 2, decimal.HalfEven)
		})},
		{"Times", "2 m", quantityString(func() (measurement.Quantity, error) { return oneMetre.Times(2) }), quantityString(func() (measurement.Quantity, error) { return oneMetre.TimesContext(context.Background(), 2) })},
		{"Clamp", "1 m", quantityString(func() (measurement.Quantity, error) {
			return oneMetre.Clamp(oneHundredCentimetres, twoMetres, conversion)
		}), quantityString(func() (measurement.Quantity, error) {
			return oneMetre.ClampContext(context.Background(), oneHundredCentimetres, twoMetres, conversion)
		})},
		{"Format", "1.00 m", func() (string, error) { return oneMetre.Format(format) }, func() (string, error) { return oneMetre.FormatContext(context.Background(), format) }},
		{"FloorArea", "2 m2", quantityString(func() (measurement.Quantity, error) { return dimensions.FloorArea(measurement.SquareMetre, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return dimensions.FloorAreaContext(context.Background(), measurement.SquareMetre, conversion)
		})},
		{"CubicVolume", "6 m3", quantityString(func() (measurement.Quantity, error) {
			return dimensions.CubicVolume(measurement.CubicMetre, conversion)
		}), quantityString(func() (measurement.Quantity, error) {
			return dimensions.CubicVolumeContext(context.Background(), measurement.CubicMetre, conversion)
		})},
		{"TotalVolume", "12 m3", quantityString(func() (measurement.Quantity, error) {
			return dimensions.TotalVolume(measurement.CubicMetre, conversion)
		}), quantityString(func() (measurement.Quantity, error) {
			return dimensions.TotalVolumeContext(context.Background(), measurement.CubicMetre, conversion)
		})},
		{"LoadingMetres", "1.0 ldm", quantityString(func() (measurement.Quantity, error) {
			return dimensions.LoadingMetres(truckWidth, stacking, conversion)
		}), quantityString(func() (measurement.Quantity, error) {
			return dimensions.LoadingMetresContext(context.Background(), truckWidth, stacking, conversion)
		})},
		{"VolumetricDivisor.Weight", "2 kg", quantityString(func() (measurement.Quantity, error) { return divisor.Weight(volume, conversion) }), quantityString(func() (measurement.Quantity, error) {
			return divisor.WeightContext(context.Background(), volume, conversion)
		})},
		{"VolumetricIndex.Weight", "6 kg", quantityString(func() (measurement.Quantity, error) {
			return index.Weight(measurement.MustNew(decimal.New(3), measurement.CubicMetre), conversion)
		}), quantityString(func() (measurement.Quantity, error) {
			return index.WeightContext(context.Background(), measurement.MustNew(decimal.New(3), measurement.CubicMetre), conversion)
		})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			legacy, legacyErr := test.legacy()
			got, gotErr := test.withCtx()
			if legacyErr != nil || gotErr != nil || got != test.want || legacy != test.want {
				t.Fatalf("context = %q, %v; legacy = %q, %v; want %q", got, gotErr, legacy, legacyErr, test.want)
			}
		})
	}
}

func TestConvertContextPreservesCancellationIdentityAtQuotient(t *testing.T) {
	t.Parallel()

	quantity := measurement.MustNew(decimal.New(1), measurement.Metre)
	for _, sentinel := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(sentinel.Error(), func(t *testing.T) {
			t.Parallel()

			ctx := &cancelAfterErrChecks{remaining: 5, err: sentinel}
			result, err := quantity.ConvertContext(ctx, measurement.Centimetre, measurement.ExactConversion())
			// Exact error identity is the regression contract at this boundary.
			if err != sentinel || result.Unit() != "" { //nolint:errorlint
				t.Fatalf("ConvertContext() = %v, %v, want zero and exact %v", result, err, sentinel)
			}
		})
	}
}

func TestContextCancellationStopsComposedAndPureDecimalWork(t *testing.T) {
	t.Parallel()

	conversion := measurement.ExactConversion()
	oneMetre := measurement.MustNew(decimal.New(1), measurement.Metre)
	oneHundredCentimetres := measurement.MustNew(decimal.New(100), measurement.Centimetre)
	twoMetres := measurement.MustNew(decimal.New(2), measurement.Metre)
	threeSquareMetres := measurement.MustNew(decimal.New(3), measurement.SquareMetre)
	dimensions, err := measurement.NewDimensions(oneMetre, twoMetres, measurement.MustNew(decimal.New(3), measurement.Metre), 2)
	if err != nil {
		t.Fatal(err)
	}
	truckWidth, err := measurement.NewTruckWidth(twoMetres)
	if err != nil {
		t.Fatal(err)
	}
	stacking, err := measurement.NewStackingFactor(decimal.New(2))
	if err != nil {
		t.Fatal(err)
	}
	divisor, err := measurement.NewVolumetricDivisor(decimal.New(5_000), measurement.CubicCentimetre)
	if err != nil {
		t.Fatal(err)
	}
	index, err := measurement.NewVolumetricIndex(measurement.MustNew(decimal.New(2), measurement.KilogramPerCubicMetre))
	if err != nil {
		t.Fatal(err)
	}
	volume := measurement.MustNew(decimal.New(10_000), measurement.CubicCentimetre)
	format := measurement.FormatOptions{Unit: measurement.Metre, Conversion: conversion, Scale: 2, Rounding: decimal.HalfEven, Separator: " "}

	operations := []struct {
		name      string
		remaining int
		call      func(context.Context) (bool, error)
	}{
		{"Convert", 2, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return oneMetre.ConvertContext(ctx, measurement.Centimetre, conversion)
		})},
		{"Add", 3, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return oneMetre.AddContext(ctx, oneHundredCentimetres, conversion)
		})},
		{"Subtract", 3, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return twoMetres.SubtractContext(ctx, oneHundredCentimetres, conversion)
		})},
		{"Equal", 2, func(ctx context.Context) (bool, error) {
			result, callErr := oneMetre.EqualContext(ctx, oneMetre, conversion)
			return !result, callErr
		}},
		{"Multiply", 2, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return oneMetre.MultiplyContext(ctx, twoMetres, conversion)
		})},
		{"Divide", 2, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return threeSquareMetres.DivideContext(ctx, oneMetre, conversion)
		})},
		{"Format", 2, func(ctx context.Context) (bool, error) {
			result, callErr := oneMetre.FormatContext(ctx, format)
			return result == "", callErr
		}},
		{"FloorArea", 3, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return dimensions.FloorAreaContext(ctx, measurement.SquareMetre, conversion)
		})},
		{"CubicVolume", 5, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return dimensions.CubicVolumeContext(ctx, measurement.CubicMetre, conversion)
		})},
		{"TotalVolume", 11, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return dimensions.TotalVolumeContext(ctx, measurement.CubicMetre, conversion)
		})},
		{"LoadingMetres", 13, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return dimensions.LoadingMetresContext(ctx, truckWidth, stacking, conversion)
		})},
		{"VolumetricDivisor.Weight", 2, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return divisor.WeightContext(ctx, volume, conversion)
		})},
		{"VolumetricIndex.Weight", 4, zeroQuantity(func(ctx context.Context) (measurement.Quantity, error) {
			return index.WeightContext(ctx, measurement.MustNew(decimal.New(3), measurement.CubicMetre), conversion)
		})},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()

			zero, callErr := operation.call(&cancelAfterErrChecks{remaining: operation.remaining})
			// Exact error identity is part of the public context contract.
			if callErr != context.Canceled || !zero { //nolint:errorlint
				t.Fatalf("operation = zero %t, %v; want zero and context.Canceled", zero, callErr)
			}
		})
	}

	beforeCompare := &cancelAfterErrChecks{remaining: 2}
	// Exact error identity is part of the public context contract.
	if result, err := oneMetre.CompareContext(beforeCompare, oneMetre, conversion); err != context.Canceled || result != 0 { //nolint:errorlint
		t.Fatalf("CompareContext() = %d, %v, want zero and context.Canceled", result, err)
	}

	beforeClamp := &cancelAfterErrChecks{remaining: 3}
	// Exact error identity is part of the public context contract.
	if result, err := oneMetre.ClampContext(beforeClamp, oneMetre, twoMetres, conversion); err != context.Canceled || result.Unit() != "" { //nolint:errorlint
		t.Fatalf("ClampContext() = %v, %v, want zero and context.Canceled", result, err)
	}
}

func TestContextIsNotRetainedAcrossConcurrentCalls(t *testing.T) {
	t.Parallel()

	quantity := measurement.MustNew(decimal.New(1), measurement.Metre)
	conversion := measurement.ExactConversion()
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	var wait sync.WaitGroup
	for index := range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			ctx := context.Background()
			if index%2 == 0 {
				ctx = canceled
			}
			result, err := quantity.ConvertContext(ctx, measurement.Centimetre, conversion)
			if ctx == canceled {
				// Exact error identity is part of the public context contract.
				if err != context.Canceled || result.Unit() != "" { //nolint:errorlint
					t.Errorf("canceled ConvertContext() = %v, %v", result, err)
				}
				return
			}
			if err != nil || result.String() != "100 cm" {
				t.Errorf("live ConvertContext() = %v, %v", result, err)
			}
		}()
	}
	wait.Wait()

	result, err := quantity.ConvertContext(context.Background(), measurement.Centimetre, conversion)
	if err != nil || result.String() != "100 cm" {
		t.Fatalf("post-cancellation ConvertContext() = %v, %v", result, err)
	}
}

func quantityString(call func() (measurement.Quantity, error)) func() (string, error) {
	return func() (string, error) {
		quantity, err := call()
		return quantity.String(), err
	}
}

func scalarString[T any](call func() (T, error)) func() (string, error) {
	return func() (string, error) {
		value, err := call()
		return fmt.Sprint(value), err
	}
}

func zeroQuantity(call func(context.Context) (measurement.Quantity, error)) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		result, err := call(ctx)
		return result.Unit() == "", err
	}
}

type cancelAfterErrChecks struct {
	mu        sync.Mutex
	remaining int
	err       error
}

func requireSignature[T any](value T) { _ = value }

func (c *cancelAfterErrChecks) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelAfterErrChecks) Done() <-chan struct{}       { return nil }
func (c *cancelAfterErrChecks) Value(any) any               { return nil }

func (c *cancelAfterErrChecks) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.remaining == 0 {
		if c.err != nil {
			return c.err
		}
		return context.Canceled
	}
	c.remaining--

	return nil
}
