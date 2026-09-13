package measurement_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
	measurementwire "github.com/faustbrian/go-measurement/v2/adapters/wire"
	"github.com/faustbrian/go-wire"
)

func TestInputDiagnosticsDoNotEchoControlledValues(t *testing.T) {
	t.Parallel()

	const controlled = "private-marker"
	unknownField := []byte(`{"private-marker":0}`)
	unknownUnit := []byte(`{"value":"1","unit":"private-marker"}`)
	tests := []struct {
		name string
		call func() error
		kind error
	}{
		{
			name: "quantity JSON field",
			call: func() error { return json.Unmarshal(unknownField, new(measurement.Quantity)) },
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name: "dimensions JSON field",
			call: func() error { return json.Unmarshal(unknownField, new(measurement.Dimensions)) },
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name: "quantity JSON unit",
			call: func() error { return json.Unmarshal(unknownUnit, new(measurement.Quantity)) },
			kind: measurement.ErrUnknownUnit,
		},
		{
			name: "SQL string",
			call: func() error { return new(measurement.Quantity).Scan(string(unknownField)) },
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name: "SQL bytes",
			call: func() error { return new(measurement.Quantity).Scan(unknownUnit) },
			kind: measurement.ErrUnknownUnit,
		},
		{
			name: "wire adapter",
			call: func() error {
				_, err := measurementwire.Decode(unknownField, wire.FormatJSON, measurementwire.Options{})
				return err
			},
			kind: wire.ErrParse,
		},
		{
			name: "profile alias",
			call: func() error {
				_, err := measurement.SymbolProfile().Resolve(controlled)
				return err
			},
			kind: measurement.ErrUnknownUnit,
		},
		{
			name: "unit dimension",
			call: func() error {
				_, err := measurement.Unit(controlled).Dimension()
				return err
			},
			kind: measurement.ErrUnknownUnit,
		},
		{
			name: "quantity constructor",
			call: func() error {
				_, err := measurement.New(decimal.New(1), measurement.Unit(controlled))
				return err
			},
			kind: measurement.ErrUnknownUnit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.call()
			if !errors.Is(err, test.kind) {
				t.Fatalf("error = %v, want %v", err, test.kind)
			}
			if strings.Contains(err.Error(), controlled) {
				t.Fatalf("error echoed controlled input: %v", err)
			}
		})
	}
}

func TestJSONDiagnosticsDoNotEchoNumericPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		controlled string
		call       func() error
		kind       error
	}{
		{
			name:       "dimensions integer overflow",
			controlled: "123456789012345678901234567890",
			call: func() error {
				return json.Unmarshal(
					[]byte(`{"quantity":123456789012345678901234567890}`),
					new(measurement.Dimensions),
				)
			},
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name:       "quantity root overflow",
			controlled: "1e123456789",
			call: func() error {
				return json.Unmarshal([]byte(`1e123456789`), new(measurement.Quantity))
			},
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name:       "quantity trailing overflow",
			controlled: "1e123456789",
			call: func() error {
				return new(measurement.Quantity).UnmarshalJSON(
					[]byte(`{"value":"1","unit":"m"} 1e123456789`),
				)
			},
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name:       "SQL numeric overflow",
			controlled: "1e123456789",
			call: func() error {
				return new(measurement.Quantity).Scan([]byte(`1e123456789`))
			},
			kind: measurement.ErrInvalidQuantity,
		},
		{
			name:       "wire numeric overflow",
			controlled: "1e123456789",
			call: func() error {
				_, err := measurementwire.Decode(
					[]byte(`1e123456789`),
					wire.FormatJSON,
					measurementwire.Options{},
				)
				return err
			},
			kind: wire.ErrValidation,
		},
		{
			name:       "wire trailing overflow",
			controlled: "1e123456789",
			call: func() error {
				_, err := measurementwire.Decode(
					[]byte(`{"value":"1","unit":"m"} 1e123456789`),
					wire.FormatJSON,
					measurementwire.Options{},
				)
				return err
			},
			kind: wire.ErrParse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.call()
			if !errors.Is(err, test.kind) {
				t.Fatalf("error = %v, want %v", err, test.kind)
			}
			if strings.Contains(err.Error(), test.controlled) {
				t.Fatalf("error echoed controlled input: %v", err)
			}
		})
	}
}
