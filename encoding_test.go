package measurement_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
)

func TestQuantityJSONPreservesDecimalAndUnitMetadata(t *testing.T) {
	t.Parallel()

	original := measurement.MustNew(decimal.MustParse("9007199254740993.125"), measurement.Kilogram)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got, want := string(data), `{"value":"9007199254740993.125","unit":"kg"}`; got != want {
		t.Fatalf("Marshal() = %s, want %s", got, want)
	}

	var decoded measurement.Quantity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := decoded.String(); got != original.String() {
		t.Fatalf("round trip = %q, want %q", got, original)
	}
	if err := json.Unmarshal([]byte(`{"value":1.25,"unit":"kg"}`), &decoded); !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("numeric JSON value error = %v", err)
	}
}

func TestQuantityCodecsRejectAmbiguousFields(t *testing.T) {
	t.Parallel()

	jsonPayloads := []string{
		`{"value":"1","value":"2","unit":"m"}`,
		`{"value":"1","unit":"m","unit":"kg"}`,
	}
	for _, payload := range jsonPayloads {
		var quantity measurement.Quantity
		if err := json.Unmarshal([]byte(payload), &quantity); !errors.Is(err, measurement.ErrInvalidQuantity) {
			t.Fatalf("json.Unmarshal(%s) error = %v, want ErrInvalidQuantity", payload, err)
		}
	}

	xmlPayloads := []string{
		`<quantity><value>1</value><value>2</value><unit>m</unit></quantity>`,
		`<quantity><value>1</value><unit>m</unit><unit>kg</unit></quantity>`,
		`<quantity><value>1</value><unit>m</unit><unknown>x</unknown></quantity>`,
	}
	for _, payload := range xmlPayloads {
		if _, err := measurement.ParseQuantityXML([]byte(payload)); !errors.Is(err, measurement.ErrInvalidQuantity) {
			t.Fatalf("ParseQuantityXML(%s) error = %v, want ErrInvalidQuantity", payload, err)
		}
	}
}

func TestQuantityXMLPreservesDecimalAndUnitMetadata(t *testing.T) {
	t.Parallel()

	original := measurement.MustNew(decimal.MustParse("12.50"), measurement.Centimetre)
	data, err := xml.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got, want := string(data), `<quantity><value>12.50</value><unit>cm</unit></quantity>`; got != want {
		t.Fatalf("Marshal() = %s, want %s", got, want)
	}

	decoded, err := measurement.ParseQuantityXML(data)
	if err != nil {
		t.Fatalf("ParseQuantityXML() error = %v", err)
	}
	if got := decoded.String(); got != original.String() {
		t.Fatalf("round trip = %q, want %q", got, original)
	}
}

func TestQuantitySQLValueAndScannerRoundTrip(t *testing.T) {
	t.Parallel()

	original := measurement.MustNew(decimal.MustParse("3.25"), measurement.Litre)
	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	var decoded measurement.Quantity
	if err := decoded.Scan(value); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if got := decoded.String(); got != original.String() {
		t.Fatalf("round trip = %q, want %q", got, original)
	}
	if err := decoded.Scan(123); !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("Scan(int) error = %v", err)
	}
}

func TestQuantitySQLScannerRejectsOversizeBeforeCopy(t *testing.T) {
	oversize := bytes.Repeat([]byte(" "), 4<<20)
	assertRejected := func(t *testing.T, source any) {
		t.Helper()

		var decoded measurement.Quantity
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		for range 8 {
			if err := decoded.Scan(source); !errors.Is(err, measurement.ErrInvalidQuantity) {
				t.Fatalf("Scan() error = %v, want ErrInvalidQuantity", err)
			}
		}
		runtime.ReadMemStats(&after)
		if got := after.TotalAlloc - before.TotalAlloc; got > 1<<20 {
			t.Fatalf("Scan() allocated %d bytes across eight oversized inputs, want at most 1 MiB", got)
		}
	}

	t.Run("bytes", func(t *testing.T) {
		assertRejected(t, oversize)
	})
	t.Run("string", func(t *testing.T) {
		assertRejected(t, string(oversize))
	})
}

func TestQuantitySQLScannerAcceptsExactSerializedByteLimit(t *testing.T) {
	t.Parallel()

	const canonical = `{"value":"1","unit":"m"}`
	payload := append(bytes.Repeat([]byte(" "), measurement.MaxSerializedBytes-len(canonical)), canonical...)
	if len(payload) != measurement.MaxSerializedBytes {
		t.Fatalf("fixture length = %d, want %d", len(payload), measurement.MaxSerializedBytes)
	}

	for _, test := range []struct {
		name   string
		source any
	}{
		{name: "bytes", source: payload},
		{name: "string", source: string(payload)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var decoded measurement.Quantity
			if err := decoded.Scan(test.source); err != nil {
				t.Fatalf("Scan(exact byte limit) error = %v", err)
			}
			if got, want := decoded.String(), "1 m"; got != want {
				t.Fatalf("Scan(exact byte limit) = %q, want %q", got, want)
			}
		})
	}
}

func TestDirectDecodersRejectOversizeDocumentsBeforeParsing(t *testing.T) {
	t.Parallel()

	oversize := bytes.Repeat([]byte(" "), measurement.MaxSerializedBytes+1)
	tests := []struct {
		name string
		call func([]byte) error
		want string
	}{
		{"quantity JSON", func(data []byte) error { return new(measurement.Quantity).UnmarshalJSON(data) }, "JSON exceeds"},
		{"dimensions JSON", func(data []byte) error { return new(measurement.Dimensions).UnmarshalJSON(data) }, "JSON exceeds"},
		{"quantity XML", func(data []byte) error { _, err := measurement.ParseQuantityXML(data); return err }, "document exceeds byte limit"},
		{"dimensions XML", func(data []byte) error { _, err := measurement.ParseDimensionsXML(data); return err }, "document exceeds byte limit"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.call(oversize)
			if !errors.Is(err, measurement.ErrInvalidQuantity) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("decode(oversize malformed document) error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestFailedJSONAndSQLDecodesPreserveQuantityForRetry(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		apply func(*measurement.Quantity, []byte) error
	}{
		{"JSON", func(q *measurement.Quantity, data []byte) error { return json.Unmarshal(data, q) }},
		{"SQL bytes", func(q *measurement.Quantity, data []byte) error { return q.Scan(data) }},
		{"SQL string", func(q *measurement.Quantity, data []byte) error { return q.Scan(string(data)) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			quantity := measurement.MustNew(decimal.New(9), measurement.Kilogram)
			if err := test.apply(&quantity, []byte(`{"value":"2","unit":"unknown"}`)); !errors.Is(err, measurement.ErrUnknownUnit) {
				t.Fatalf("invalid decode error = %v, want ErrUnknownUnit", err)
			}
			if got, want := quantity.String(), "9 kg"; got != want {
				t.Fatalf("quantity after failed decode = %q, want %q", got, want)
			}
			if err := test.apply(&quantity, []byte(`{"value":"2","unit":"m"}`)); err != nil {
				t.Fatalf("retry decode error = %v", err)
			}
			if got, want := quantity.String(), "2 m"; got != want {
				t.Fatalf("quantity after retry = %q, want %q", got, want)
			}
		})
	}
}

func TestFailedDimensionsJSONDecodePreservesReceiverForRetry(t *testing.T) {
	t.Parallel()

	one := measurement.MustNew(decimal.New(1), measurement.Metre)
	dimensions, err := measurement.NewDimensions(one, one, one, 2)
	if err != nil {
		t.Fatalf("NewDimensions() error = %v", err)
	}
	before := dimensions
	if err := json.Unmarshal([]byte(`{"length":{"value":"2","unit":"m"}}`), &dimensions); err == nil {
		t.Fatal("incomplete dimensions decoded successfully")
	}
	if !reflect.DeepEqual(dimensions, before) {
		t.Fatal("failed dimensions decode mutated the existing receiver")
	}

	two := measurement.MustNew(decimal.New(2), measurement.Metre)
	want, err := measurement.NewDimensions(two, two, two, 3)
	if err != nil {
		t.Fatalf("NewDimensions(retry) error = %v", err)
	}
	payload, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal(retry) error = %v", err)
	}
	if err := json.Unmarshal(payload, &dimensions); err != nil {
		t.Fatalf("retry dimensions decode error = %v", err)
	}
	if got := dimensions.Length().String(); got != "2 m" || dimensions.Quantity() != 3 {
		t.Fatalf("dimensions after retry = %q x%d, want 2 m x3", got, dimensions.Quantity())
	}
}

func TestFormatRequiresExplicitTargetAndRounding(t *testing.T) {
	t.Parallel()

	quantity := measurement.MustNew(decimal.New(1), measurement.Metre)
	formatted, err := quantity.Format(measurement.FormatOptions{
		Unit:       measurement.Foot,
		Conversion: measurement.RoundedConversion(3, decimal.HalfEven),
		Scale:      2,
		Rounding:   decimal.HalfEven,
		Separator:  " ",
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if formatted != "3.28 ft" {
		t.Fatalf("Format() = %q, want %q", formatted, "3.28 ft")
	}
}

func TestDimensionsJSONAndXMLRoundTripAllUnits(t *testing.T) {
	t.Parallel()

	original, err := measurement.NewDimensions(
		measurement.MustNew(decimal.MustParse("1.2"), measurement.Metre),
		measurement.MustNew(decimal.New(80), measurement.Centimetre),
		measurement.MustNew(decimal.New(600), measurement.Millimetre),
		2,
	)
	if err != nil {
		t.Fatalf("NewDimensions() error = %v", err)
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("JSON Marshal() error = %v", err)
	}
	var fromJSON measurement.Dimensions
	if err := json.Unmarshal(jsonData, &fromJSON); err != nil {
		t.Fatalf("JSON Unmarshal() error = %v", err)
	}
	if fromJSON.Length().String() != "1.2 m" || fromJSON.Width().String() != "80 cm" ||
		fromJSON.Height().String() != "600 mm" || fromJSON.Quantity() != 2 {
		t.Fatalf("JSON round trip lost metadata: %s %s %s x%d", fromJSON.Length(), fromJSON.Width(), fromJSON.Height(), fromJSON.Quantity())
	}

	xmlData, err := xml.Marshal(original)
	if err != nil {
		t.Fatalf("XML Marshal() error = %v", err)
	}
	fromXML, err := measurement.ParseDimensionsXML(xmlData)
	if err != nil {
		t.Fatalf("ParseDimensionsXML() error = %v", err)
	}
	if fromXML.Length().String() != "1.2 m" || fromXML.Width().String() != "80 cm" ||
		fromXML.Height().String() != "600 mm" || fromXML.Quantity() != 2 {
		t.Fatalf("XML round trip lost metadata: %s %s %s x%d", fromXML.Length(), fromXML.Width(), fromXML.Height(), fromXML.Quantity())
	}

	if err := json.Unmarshal([]byte(`{"length":{"value":"1","unit":"m"},"width":{"value":"1","unit":"m"},"height":{"value":"1","unit":"m"},"quantity":0}`), &fromJSON); !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("invalid dimensions error = %v", err)
	}
}

func TestDimensionsCodecsRejectAmbiguousFields(t *testing.T) {
	t.Parallel()

	quantity := `{"value":"1","unit":"m"}`
	jsonPayloads := []string{
		`{"length":` + quantity + `,"length":` + quantity + `,"width":` + quantity + `,"height":` + quantity + `,"quantity":1}`,
		`{"length":` + quantity + `,"width":` + quantity + `,"height":` + quantity + `,"quantity":1,"quantity":2}`,
	}
	for _, payload := range jsonPayloads {
		var dimensions measurement.Dimensions
		if err := json.Unmarshal([]byte(payload), &dimensions); !errors.Is(err, measurement.ErrInvalidQuantity) {
			t.Fatalf("json.Unmarshal(%s) error = %v, want ErrInvalidQuantity", payload, err)
		}
	}

	xmlPayloads := []string{
		`<dimensions><length><value>1</value><unit>m</unit></length><length><value>1</value><unit>m</unit></length><width><value>1</value><unit>m</unit></width><height><value>1</value><unit>m</unit></height><quantity>1</quantity></dimensions>`,
		`<dimensions><length><value>1</value><unit>m</unit></length><width><value>1</value><unit>m</unit></width><height><value>1</value><unit>m</unit></height><quantity>1</quantity><unknown>x</unknown></dimensions>`,
	}
	for _, payload := range xmlPayloads {
		if _, err := measurement.ParseDimensionsXML([]byte(payload)); !errors.Is(err, measurement.ErrInvalidQuantity) {
			t.Fatalf("ParseDimensionsXML(%s) error = %v, want ErrInvalidQuantity", payload, err)
		}
	}
}
