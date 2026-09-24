package measurementwire_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
	"github.com/faustbrian/go-measurement/v2/measurementwire"
	"github.com/faustbrian/go-wire"
)

func TestJSONAndXMLRoundTripsPreserveUnitMetadata(t *testing.T) {
	t.Parallel()

	original := measurement.MustNew(decimal.MustParse("12.50"), measurement.Kilogram)
	for _, format := range []wire.Format{wire.FormatJSON, wire.FormatXML} {
		format := format
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()
			payload, err := measurementwire.Encode(original, format, measurementwire.Options{MaxBytes: 1024})
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			decoded, err := measurementwire.Decode(payload, format, measurementwire.Options{MaxBytes: 1024})
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if decoded.String() != original.String() {
				t.Fatalf("round trip = %q, want %q", decoded, original)
			}
		})
	}
}

func TestAdapterRejectsUnsupportedFormatsAndOversizePayloads(t *testing.T) {
	t.Parallel()

	quantity := measurement.MustNew(decimal.New(1), measurement.Metre)
	if _, err := measurementwire.Encode(quantity, wire.FormatYAML, measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
		t.Fatalf("Encode(YAML) error = %v", err)
	}
	if _, err := measurementwire.Decode(nil, wire.FormatYAML, measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
		t.Fatalf("Decode(YAML) error = %v", err)
	}
	if _, err := measurementwire.Encode(quantity, wire.Format("unknown"), measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
		t.Fatalf("Encode(unknown) error = %v", err)
	}
	if _, err := measurementwire.Decode(nil, wire.Format("unknown"), measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
		t.Fatalf("Decode(unknown) error = %v", err)
	}
	if _, err := measurementwire.Decode([]byte(`{"value":"1","unit":"m"}`), wire.FormatJSON, measurementwire.Options{MaxBytes: 4}); !errors.Is(err, wire.ErrSizeLimit) {
		t.Fatalf("Decode(oversize) error = %v", err)
	}
}

func TestLegacyXMLDecodeMatchesExactConfiguredLimit(t *testing.T) {
	t.Parallel()

	payload := []byte(`<quantity><value>1</value><unit>m</unit></quantity>`)
	quantity, err := measurementwire.Decode(payload, wire.FormatXML, measurementwire.Options{MaxBytes: int64(len(payload))})
	if err != nil {
		t.Fatalf("Decode(exact configured limit) error = %v", err)
	}
	if got, want := quantity.String(), "1 m"; got != want {
		t.Fatalf("Decode(exact configured limit) = %q, want %q", got, want)
	}
	if _, err := measurementwire.Decode(payload, wire.FormatXML, measurementwire.Options{MaxBytes: int64(len(payload) - 1)}); !errors.Is(err, wire.ErrSizeLimit) {
		t.Fatalf("Decode(one below payload length) error = %v, want ErrSizeLimit", err)
	}
}

func TestLegacyXMLDecodeClassifiesSyntaxAndValueErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		payload string
		want    error
	}{
		{name: "truncated XML", payload: `<quantity>`, want: wire.ErrParse},
		{name: "invalid value", payload: `<quantity><value>1</value><unit>unknown</unit></quantity>`, want: wire.ErrValidation},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := measurementwire.Decode([]byte(test.payload), wire.FormatXML, measurementwire.Options{})
			if !errors.Is(err, test.want) {
				t.Fatalf("Decode() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestDecodeErrorsDoNotEchoPayload(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		controlled string
		payload    string
		kind       error
	}{
		{
			name:       "unknown unit",
			controlled: "private-marker",
			payload:    `{"value":"1","unit":"private-marker"}`,
			kind:       wire.ErrParse,
		},
		{
			name:       "numeric overflow",
			controlled: "1e123456789",
			payload:    `1e123456789`,
			kind:       wire.ErrValidation,
		},
		{
			name:       "trailing numeric overflow",
			controlled: "1e123456789",
			payload:    `{"value":"1","unit":"m"} 1e123456789`,
			kind:       wire.ErrParse,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := measurementwire.Decode(
				[]byte(test.payload),
				wire.FormatJSON,
				measurementwire.Options{},
			)
			if !errors.Is(err, test.kind) {
				t.Fatalf("Decode() error = %v, want %v", err, test.kind)
			}
			if strings.Contains(err.Error(), test.controlled) {
				t.Fatalf("Decode() error echoed controlled input: %v", err)
			}
		})
	}
}
