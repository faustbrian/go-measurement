package measurementwire_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
	measurementwire "github.com/faustbrian/go-measurement/v2/adapters/wire"
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
	for _, format := range []wire.Format{
		wire.FormatSOAP,
		wire.FormatYAML,
		wire.FormatTOML,
		wire.FormatMessagePack,
		wire.FormatCBOR,
		wire.FormatBSON,
		wire.Format("unknown"),
	} {
		format := format
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()

			if _, err := measurementwire.Encode(quantity, format, measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
				t.Fatalf("Encode(%q) error = %v", format, err)
			}
			if _, err := measurementwire.Decode(nil, format, measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
				t.Fatalf("Decode(%q) error = %v", format, err)
			}
		})
	}
	if _, err := measurementwire.Decode([]byte(`{"value":"1","unit":"m"}`), wire.FormatJSON, measurementwire.Options{MaxBytes: 4}); !errors.Is(err, wire.ErrSizeLimit) {
		t.Fatalf("Decode(oversize) error = %v", err)
	}
	if _, err := measurementwire.Decode([]byte(`<quantity/>`), wire.FormatXML, measurementwire.Options{MaxBytes: 4}); !errors.Is(err, wire.ErrSizeLimit) {
		t.Fatalf("Decode(XML oversize) error = %v", err)
	}
	if _, err := measurementwire.Decode([]byte(`<quantity/>`), wire.FormatXML, measurementwire.Options{MaxBytes: -1}); !errors.Is(err, wire.ErrValidation) {
		t.Fatalf("Decode(XML negative limit) error = %v", err)
	}
	coreOversize := []byte(
		`<quantity><value>` +
			strings.Repeat("9", measurement.MaxSerializedBytes) +
			`</value><unit>m</unit></quantity>`,
	)
	if _, err := measurementwire.Decode(coreOversize, wire.FormatXML, measurementwire.Options{MaxBytes: int64(len(coreOversize))}); !errors.Is(err, wire.ErrValidation) || !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("Decode(XML core limit) error = %v", err)
	}
}

func TestXMLDecodeAcceptsExactConfiguredLimit(t *testing.T) {
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

func TestXMLDecodeClassifiesSyntaxAndValueErrors(t *testing.T) {
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

func TestDecodeDoesNotRetainCallerPayload(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"value":"4.50","unit":"kg"}`)
	decoded, err := measurementwire.Decode(payload, wire.FormatJSON, measurementwire.Options{})
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	clear(payload)
	if decoded.String() != "4.50 kg" {
		t.Fatalf("decoded quantity after caller mutation = %q", decoded)
	}
}

func TestXMLDecodeUsesBoundedDefault(t *testing.T) {
	t.Parallel()

	decoded, err := measurementwire.Decode(
		[]byte(`<quantity><value>4.50</value><unit>kg</unit></quantity>`),
		wire.FormatXML,
		measurementwire.Options{},
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got := decoded.String(); got != "4.50 kg" {
		t.Fatalf("Decode() = %q, want %q", got, "4.50 kg")
	}
}
