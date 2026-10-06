//lint:file-ignore SA1019 Compatibility coverage intentionally exercises the deprecated facade.
//nolint:staticcheck // Compatibility coverage intentionally exercises the deprecated facade.
package measurement_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v3"
	measurementwire "github.com/faustbrian/go-measurement/v3/adapters/wire"
	legacywire "github.com/faustbrian/go-measurement/v3/measurementwire"
	"github.com/faustbrian/go-wire/v3"
)

func TestWireAdaptersExposeCurrentWireErrorIdentity(t *testing.T) {
	t.Parallel()

	for name, decode := range map[string]func([]byte, wire.Format) error{
		"adapter": func(payload []byte, format wire.Format) error {
			_, err := measurementwire.Decode(payload, format, measurementwire.Options{MaxBytes: 128})
			return err
		},
		"facade": func(payload []byte, format wire.Format) error {
			_, err := legacywire.Decode(payload, format, legacywire.Options{MaxBytes: 128})
			return err
		},
	} {
		for _, test := range []struct {
			name    string
			payload []byte
			format  wire.Format
			kind    error
		}{
			{"JSON parse", []byte(`{"value":`), wire.FormatJSON, wire.ErrParse},
			{"XML parse", []byte(`<quantity>`), wire.FormatXML, wire.ErrParse},
			{"JSON unknown unit", []byte(`{"value":"1","unit":"unknown"}`), wire.FormatJSON, wire.ErrParse},
			{"XML validation", []byte(`<quantity><value>1</value><unit>unknown</unit></quantity>`), wire.FormatXML, wire.ErrValidation},
			{"unsupported", nil, wire.FormatYAML, wire.ErrUnsupportedFormat},
			{"size", make([]byte, 129), wire.FormatXML, wire.ErrSizeLimit},
		} {
			t.Run(name+"/"+test.name, func(t *testing.T) {
				err := decode(test.payload, test.format)
				var classified *wire.Error
				if !errors.Is(err, test.kind) || !errors.As(err, &classified) {
					t.Fatalf("error = %v, want Wire v3 classification %v and *wire.Error", err, test.kind)
				}
				if classified.Format != test.format || classified.Op != "decode" {
					t.Fatalf("classification = %+v, want selected format and decode operation", classified)
				}
			})
		}
	}
}

func TestWireSuccessorPreservesBoundedSerializationContract(t *testing.T) {
	t.Parallel()

	original := measurement.MustNew(decimal.MustParse("12.50"), measurement.Kilogram)
	for _, format := range []wire.Format{wire.FormatJSON, wire.FormatXML} {
		format := format
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()

			options := measurementwire.Options{MaxBytes: 1024}
			payload, err := measurementwire.Encode(original, format, options)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			decoded, err := measurementwire.Decode(payload, format, options)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if decoded.String() != original.String() {
				t.Fatalf("round trip = %q, want %q", decoded, original)
			}
		})
	}

	if _, err := measurementwire.Decode([]byte(`{"value":"1","unit":"m"}`), wire.FormatJSON, measurementwire.Options{MaxBytes: 4}); !errors.Is(err, wire.ErrSizeLimit) {
		t.Fatalf("Decode(oversize) error = %v", err)
	}
	if _, err := measurementwire.Encode(original, wire.FormatYAML, measurementwire.Options{}); !errors.Is(err, wire.ErrUnsupportedFormat) {
		t.Fatalf("Encode(YAML) error = %v", err)
	}
}

func TestLegacyWireRemainsDistinctDelegatingCompatibilityFacade(t *testing.T) {
	t.Parallel()

	if got := reflect.TypeOf(legacywire.Options{}).PkgPath(); got != "github.com/faustbrian/go-measurement/v3/measurementwire" {
		t.Fatalf("legacy Options package = %q", got)
	}
	if got := reflect.TypeOf(measurementwire.Options{}).PkgPath(); got != "github.com/faustbrian/go-measurement/v3/adapters/wire" {
		t.Fatalf("successor Options package = %q", got)
	}

	original := measurement.MustNew(decimal.MustParse("7.25"), measurement.Metre)
	legacyPayload, legacyErr := legacywire.Encode(original, wire.FormatJSON, legacywire.Options{})
	successorPayload, successorErr := measurementwire.Encode(original, wire.FormatJSON, measurementwire.Options{})
	if legacyErr != nil || successorErr != nil || string(legacyPayload) != string(successorPayload) {
		t.Fatalf("default encode mismatch: legacy=(%q, %v), successor=(%q, %v)", legacyPayload, legacyErr, successorPayload, successorErr)
	}
	legacyDecoded, legacyErr := legacywire.Decode(legacyPayload, wire.FormatJSON, legacywire.Options{})
	successorDecoded, successorErr := measurementwire.Decode(successorPayload, wire.FormatJSON, measurementwire.Options{})
	if legacyErr != nil || successorErr != nil || legacyDecoded.String() != successorDecoded.String() {
		t.Fatalf("default decode mismatch: legacy=(%v, %v), successor=(%v, %v)", legacyDecoded, legacyErr, successorDecoded, successorErr)
	}

	legacyUnsupported, legacyErr := legacywire.Encode(original, wire.FormatYAML, legacywire.Options{})
	successorUnsupported, successorErr := measurementwire.Encode(original, wire.FormatYAML, measurementwire.Options{})
	if legacyUnsupported != nil || successorUnsupported != nil ||
		!errors.Is(legacyErr, wire.ErrUnsupportedFormat) || !errors.Is(successorErr, wire.ErrUnsupportedFormat) {
		t.Fatalf("unsupported format mismatch: legacy=(%v, %v), successor=(%v, %v)", legacyUnsupported, legacyErr, successorUnsupported, successorErr)
	}
}
