package measurement_test

import (
	"encoding/xml"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v2"
)

func TestDirectXMLUnmarshalFailsClosed(t *testing.T) {
	t.Parallel()
	quantity := measurement.MustNew(decimal.New(9), measurement.Kilogram)
	length := measurement.MustNew(decimal.New(9), measurement.Metre)
	dimensions, err := measurement.NewDimensions(length, length, length, 2)
	if err != nil {
		t.Fatalf("NewDimensions() error = %v", err)
	}
	beforeDimensions := dimensions

	tests := []struct {
		name      string
		payload   string
		target    any
		unchanged func() bool
	}{
		{
			name:      "quantity",
			payload:   `<quantity><value>12.50</value><unit>cm</unit></quantity>`,
			target:    &quantity,
			unchanged: func() bool { return quantity.String() == "9 kg" },
		},
		{
			name: "dimensions",
			payload: `<dimensions>` +
				`<length><value>1.2</value><unit>m</unit></length>` +
				`<width><value>80</value><unit>cm</unit></width>` +
				`<height><value>600</value><unit>mm</unit></height>` +
				`<quantity>2</quantity>` +
				`</dimensions>`,
			target:    &dimensions,
			unchanged: func() bool { return reflect.DeepEqual(dimensions, beforeDimensions) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := xml.Unmarshal([]byte(test.payload), test.target)
			if !errors.Is(err, measurement.ErrInvalidQuantity) ||
				!errors.Is(err, measurement.ErrUnboundedXML) {
				t.Fatalf("xml.Unmarshal() error = %v, want ErrInvalidQuantity and ErrUnboundedXML", err)
			}
			if !test.unchanged() {
				t.Fatal("raw XML failure mutated the existing receiver")
			}
		})
	}
}

func TestBoundedXMLParsersPreserveMetadata(t *testing.T) {
	t.Parallel()

	quantity, err := measurement.ParseQuantityXML(
		[]byte("<?xml version=\"1.0\"?><quantity><value>12.50</value><unit>cm</unit></quantity>\n"),
	)
	if err != nil {
		t.Fatalf("ParseQuantityXML() error = %v", err)
	}
	if got := quantity.String(); got != "12.50 cm" {
		t.Fatalf("ParseQuantityXML() = %q, want %q", got, "12.50 cm")
	}

	dimensions, err := measurement.ParseDimensionsXML([]byte(
		`<dimensions>` +
			`<length><value>1.2</value><unit>m</unit></length>` +
			`<width><value>80</value><unit>cm</unit></width>` +
			`<height><value>600</value><unit>mm</unit></height>` +
			`<quantity>2</quantity>` +
			`</dimensions>`,
	))
	if err != nil {
		t.Fatalf("ParseDimensionsXML() error = %v", err)
	}
	if dimensions.Length().String() != "1.2 m" ||
		dimensions.Width().String() != "80 cm" ||
		dimensions.Height().String() != "600 mm" ||
		dimensions.Quantity() != 2 {
		t.Fatalf(
			"ParseDimensionsXML() lost metadata: %s %s %s x%d",
			dimensions.Length(),
			dimensions.Width(),
			dimensions.Height(),
			dimensions.Quantity(),
		)
	}
}

func TestBoundedXMLParsersRejectHostileWork(t *testing.T) {
	t.Parallel()

	valid := `<quantity><value>1</value><unit>m</unit></quantity>`
	exactLimit := []byte(strings.Repeat(" ", measurement.MaxSerializedBytes-len(valid)) + valid)
	if _, err := measurement.ParseQuantityXML(exactLimit); err != nil {
		t.Fatalf("ParseQuantityXML(exact document limit) error = %v", err)
	}

	oversize := []byte(
		`<quantity><value>` +
			strings.Repeat("9", measurement.MaxSerializedBytes) +
			`</value><unit>m</unit></quantity>`,
	)
	if _, err := measurement.ParseQuantityXML(oversize); !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("ParseQuantityXML(oversize) error = %v, want ErrInvalidQuantity", err)
	}
	oversizeScalar := []byte(
		`<quantity><value>` +
			strings.Repeat("9", measurement.MaxTextBytes+1) +
			`</value><unit>m</unit></quantity>`,
	)
	if _, err := measurement.ParseQuantityXML(oversizeScalar); !errors.Is(err, measurement.ErrInvalidQuantity) || !strings.Contains(err.Error(), "field exceeds byte limit") {
		t.Fatalf("ParseQuantityXML(oversize scalar) error = %v", err)
	}
	exactScalar := []byte(
		`<quantity><value>` +
			strings.Repeat("9", measurement.MaxTextBytes) +
			`</value><unit>m</unit></quantity>`,
	)
	if _, err := measurement.ParseQuantityXML(exactScalar); err != nil {
		t.Fatalf("ParseQuantityXML(exact scalar limit) error = %v", err)
	}

	tests := []struct {
		name       string
		payload    string
		wantReason string
	}{
		{
			name:    "nested scalar",
			payload: `<quantity><value><nested><deeper>1</deeper></nested></value><unit>m</unit></quantity>`,
		},
		{
			name:       "attributes",
			payload:    `<quantity private="metadata"><value>1</value><unit>m</unit></quantity>`,
			wantReason: "invalid root element",
		},
		{
			name:       "namespace",
			payload:    `<quantity xmlns="urn:attacker"><value>1</value><unit>m</unit></quantity>`,
			wantReason: "invalid root element",
		},
		{
			name:       "wrong root",
			payload:    `<other><value>1</value><unit>m</unit></other>`,
			wantReason: "invalid root element",
		},
		{
			name:       "field attribute",
			payload:    `<quantity><value private="metadata">1</value><unit>m</unit></quantity>`,
			wantReason: "invalid quantity field",
		},
		{
			name:       "field namespace",
			payload:    `<quantity><value xmlns="urn:attacker">1</value><unit>m</unit></quantity>`,
			wantReason: "invalid quantity field",
		},
		{
			name: "entity declaration",
			payload: `<!DOCTYPE quantity [<!ENTITY x "9999999999">]>` +
				`<quantity><value>&x;</value><unit>m</unit></quantity>`,
		},
		{
			name:    "processing instruction",
			payload: `<quantity><?target data?><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name:    "prologue processing instruction",
			payload: `<?target data?><quantity><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name:    "invalid XML pseudo declaration",
			payload: `<?xml private-marker?><quantity><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name: "duplicate XML declaration",
			payload: `<?xml version="1.0"?><?xml version="1.0"?>` +
				`<quantity><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name:    "prologue text",
			payload: `text<quantity><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name:    "comment",
			payload: `<quantity><!-- repeated token --><value>1</value><unit>m</unit></quantity>`,
		},
		{
			name:       "leading reference marker",
			payload:    `&<quantity><value>1</value><unit>m</unit></quantity>`,
			wantReason: "unsupported XML construct",
		},
		{
			name:       "prologue comment",
			payload:    `<!--comment--><quantity><value>1</value><unit>m</unit></quantity>`,
			wantReason: "unsupported XML construct",
		},
		{
			name:    "truncated quantity",
			payload: `<quantity>`,
		},
		{
			name: "trailing document",
			payload: `<quantity><value>1</value><unit>m</unit></quantity>` +
				`<quantity><value>2</value><unit>m</unit></quantity>`,
		},
		{
			name:    "malformed trailing content",
			payload: `<quantity><value>1</value><unit>m</unit></quantity></other>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := measurement.ParseQuantityXML([]byte(test.payload))
			if !errors.Is(err, measurement.ErrInvalidQuantity) {
				t.Fatalf("ParseQuantityXML() error = %v, want ErrInvalidQuantity", err)
			}
			if test.wantReason != "" && !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("ParseQuantityXML() error = %v, want reason %q", err, test.wantReason)
			}
		})
	}
}

func TestXMLScalarLimitsDistinguishBoundaryFromSemanticFailure(t *testing.T) {
	t.Parallel()

	unitDocument := func(unit string) []byte {
		return []byte(`<quantity><value>1</value><unit>` + unit + `</unit></quantity>`)
	}
	if _, err := measurement.ParseQuantityXML(unitDocument(strings.Repeat("x", measurement.MaxAliasBytes))); !errors.Is(err, measurement.ErrUnknownUnit) {
		t.Fatalf("exact-limit unknown unit error = %v, want ErrUnknownUnit", err)
	}
	if _, err := measurement.ParseQuantityXML(unitDocument(strings.Repeat("x", measurement.MaxAliasBytes+1))); !errors.Is(err, measurement.ErrInvalidQuantity) || !strings.Contains(err.Error(), "field exceeds byte limit") {
		t.Fatalf("over-limit unit error = %v, want byte-limit rejection", err)
	}

	dimensionsDocument := func(count string) []byte {
		return []byte(`<dimensions>` +
			`<length><value>1</value><unit>m</unit></length>` +
			`<width><value>1</value><unit>m</unit></width>` +
			`<height><value>1</value><unit>m</unit></height>` +
			`<quantity>` + count + `</quantity></dimensions>`)
	}
	exactCount := strings.Repeat("0", 19) + "1"
	dimensions, err := measurement.ParseDimensionsXML(dimensionsDocument(exactCount))
	if err != nil {
		t.Fatalf("exact 20-byte package count error = %v", err)
	}
	if dimensions.Quantity() != 1 {
		t.Fatalf("exact 20-byte package count = %d, want 1", dimensions.Quantity())
	}
	if _, err := measurement.ParseDimensionsXML(dimensionsDocument("0" + exactCount)); !errors.Is(err, measurement.ErrInvalidQuantity) || !strings.Contains(err.Error(), "field exceeds byte limit") {
		t.Fatalf("over-limit package count error = %v, want byte-limit rejection", err)
	}
}

func TestBoundedDimensionsXMLRejectsInvalidValuesAndTrailingWork(t *testing.T) {
	t.Parallel()

	validSide := `<value>1</value><unit>m</unit>`
	valid := `<dimensions><length>` + validSide + `</length><width>` + validSide +
		`</width><height>` + validSide + `</height><quantity>1</quantity></dimensions>`
	tests := []struct {
		name       string
		payload    string
		kind       error
		wantReason string
	}{
		{name: "truncated root", payload: `<dimensions>`, kind: measurement.ErrInvalidQuantity},
		{name: "trailing document", payload: valid + `<dimensions/>`, kind: measurement.ErrInvalidQuantity},
		{
			name: "unknown side unit",
			payload: `<dimensions><length><value>1</value><unit>unknown</unit></length>` +
				`<width>` + validSide + `</width><height>` + validSide +
				`</height><quantity>1</quantity></dimensions>`,
			kind: measurement.ErrUnknownUnit,
		},
		{
			name: "wrong side dimension",
			payload: `<dimensions><length><value>1</value><unit>kg</unit></length>` +
				`<width>` + validSide + `</width><height>` + validSide +
				`</height><quantity>1</quantity></dimensions>`,
			kind: measurement.ErrDimensionMismatch,
		},
		{
			name:       "unknown field",
			payload:    `<dimensions><other/></dimensions>`,
			kind:       measurement.ErrInvalidQuantity,
			wantReason: "invalid dimensions field",
		},
		{
			name:       "field attribute",
			payload:    `<dimensions><quantity private="metadata">1</quantity></dimensions>`,
			kind:       measurement.ErrInvalidQuantity,
			wantReason: "invalid dimensions field",
		},
		{
			name:       "field namespace",
			payload:    `<dimensions><quantity xmlns="urn:attacker">1</quantity></dimensions>`,
			kind:       measurement.ErrInvalidQuantity,
			wantReason: "invalid dimensions field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := measurement.ParseDimensionsXML([]byte(test.payload))
			if !errors.Is(err, test.kind) {
				t.Fatalf("ParseDimensionsXML() error = %v, want %v", err, test.kind)
			}
			if test.wantReason != "" && !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("ParseDimensionsXML() error = %v, want reason %q", err, test.wantReason)
			}
		})
	}
}

func TestBoundedQuantityXMLClassifiesUnknownUnitWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	const controlled = "attacker-controlled-unit"
	_, err := measurement.ParseQuantityXML([]byte(
		`<quantity><value>1</value><unit>` + controlled + `</unit></quantity>`,
	))
	if !errors.Is(err, measurement.ErrUnknownUnit) {
		t.Fatalf("ParseQuantityXML() error = %v, want ErrUnknownUnit", err)
	}
	if strings.Contains(err.Error(), controlled) {
		t.Fatalf("ParseQuantityXML() error echoed attacker-controlled unit: %v", err)
	}
}

func TestBoundedXMLDiagnosticsDoNotEchoInput(t *testing.T) {
	t.Parallel()

	const controlled = "attacker-controlled-field"
	_, err := measurement.ParseQuantityXML([]byte(
		`<quantity><` + controlled + `>value</` + controlled + `></quantity>`,
	))
	if !errors.Is(err, measurement.ErrInvalidQuantity) {
		t.Fatalf("ParseQuantityXML() error = %v, want ErrInvalidQuantity", err)
	}
	if strings.Contains(err.Error(), controlled) {
		t.Fatalf("ParseQuantityXML() error echoed attacker-controlled XML name: %v", err)
	}
}
