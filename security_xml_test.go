package measurement_test

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	measurement "github.com/faustbrian/go-measurement/v2"
)

func TestDirectXMLUnmarshalFailsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		target  any
	}{
		{
			name:    "quantity",
			payload: `<quantity><value>12.50</value><unit>cm</unit></quantity>`,
			target:  new(measurement.Quantity),
		},
		{
			name: "dimensions",
			payload: `<dimensions>` +
				`<length><value>1.2</value><unit>m</unit></length>` +
				`<width><value>80</value><unit>cm</unit></width>` +
				`<height><value>600</value><unit>mm</unit></height>` +
				`<quantity>2</quantity>` +
				`</dimensions>`,
			target: new(measurement.Dimensions),
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
