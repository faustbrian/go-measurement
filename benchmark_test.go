package measurement_test

import (
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v3"
)

// BenchmarkConvertMetresToFeet measures conversion of 123.456 m to ft at scale
// 6 using half-even rounding; setup is excluded by b.Loop. This is a bounded
// single-workload benchmark, not an exact-versus-rounded performance comparison.
func BenchmarkConvertMetresToFeet(b *testing.B) {
	quantity := measurement.MustNew(decimal.MustParse("123.456"), measurement.Metre)
	context := measurement.RoundedConversion(6, decimal.HalfEven)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := quantity.Convert(measurement.Foot, context); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDimensionsLoadingMetres measures the exact formula for ten
// 1.2 x 0.8 x 1.5 m packages, a 2.4 m truck and stacking factor 2. Construction
// is outside b.Loop; every measured iteration includes all formula stages.
func BenchmarkDimensionsLoadingMetres(b *testing.B) {
	dimensions, _ := measurement.NewDimensions(
		measurement.MustNew(decimal.MustParse("1.2"), measurement.Metre),
		measurement.MustNew(decimal.MustParse("0.8"), measurement.Metre),
		measurement.MustNew(decimal.MustParse("1.5"), measurement.Metre),
		10,
	)
	truckWidth, _ := measurement.NewTruckWidth(
		measurement.MustNew(decimal.MustParse("2.4"), measurement.Metre),
	)
	stacking, _ := measurement.NewStackingFactor(decimal.New(2))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := dimensions.LoadingMetres(truckWidth, stacking, measurement.ExactConversion()); err != nil {
			b.Fatal(err)
		}
	}
}
