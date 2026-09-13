package measurement

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/faustbrian/go-math/decimal"
)

const (
	// MaxSerializedBytes bounds direct JSON, XML, and SQL decoding.
	MaxSerializedBytes = 64 << 10
	// MaxXMLDepth bounds the fixed measurement XML schema.
	MaxXMLDepth = 3
	// MaxXMLTokens bounds token work in one measurement XML document.
	MaxXMLTokens = 64

	maxPackageQuantityTextBytes = 20
)

type encodedQuantity struct {
	Value string `json:"value" xml:"value"`
	Unit  Unit   `json:"unit" xml:"unit"`
}

type encodedDimensions struct {
	Length   Quantity `json:"length" xml:"length"`
	Width    Quantity `json:"width" xml:"width"`
	Height   Quantity `json:"height" xml:"height"`
	Quantity uint64   `json:"quantity" xml:"quantity"`
}

// MarshalJSON encodes value and unit metadata with the decimal as a string.
func (q Quantity) MarshalJSON() ([]byte, error) {
	if _, err := definitionFor(q.unit); err != nil {
		return nil, err
	}

	return json.Marshal(encodedQuantity{Value: q.amount.String(), Unit: q.unit})
}

// UnmarshalJSON decodes one bounded strict quantity object.
func (q *Quantity) UnmarshalJSON(data []byte) error {
	if q == nil {
		return ErrInvalidQuantity
	}
	if len(data) > MaxSerializedBytes {
		return fmt.Errorf("%w: JSON exceeds %d bytes", ErrInvalidQuantity, MaxSerializedBytes)
	}
	var encoded encodedQuantity
	if err := decodeJSONObject(data, map[string]struct{}{"value": {}, "unit": {}}, &encoded); err != nil {
		return err
	}

	return q.decode(encoded)
}

// MarshalXML encodes value and unit metadata without numeric narrowing.
func (q Quantity) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if _, err := definitionFor(q.unit); err != nil {
		return err
	}
	if start.Name.Local == "" || start.Name.Local == "Quantity" {
		start.Name.Local = "quantity"
	}

	return encoder.EncodeElement(encodedQuantity{Value: q.amount.String(), Unit: q.unit}, start)
}

// UnmarshalXML fails closed because encoding/xml invokes this callback only
// after its caller-owned decoder has parsed and allocated the start token. Use
// ParseQuantityXML for byte-bounded XML input.
func (q *Quantity) UnmarshalXML(_ *xml.Decoder, _ xml.StartElement) error {
	return fmt.Errorf("%w: %w", ErrInvalidQuantity, ErrUnboundedXML)
}

// MarshalText returns canonical amount-space-symbol text.
func (q Quantity) MarshalText() ([]byte, error) {
	if _, err := definitionFor(q.unit); err != nil {
		return nil, err
	}

	return []byte(q.String()), nil
}

// UnmarshalText parses canonical text with SymbolProfile.
func (q *Quantity) UnmarshalText(text []byte) error {
	if q == nil {
		return ErrInvalidQuantity
	}
	parsed, err := Parse(string(text), SymbolProfile())
	if err != nil {
		return err
	}
	*q = parsed

	return nil
}

// Value stores a quantity as lossless JSON text suitable for SQL text or JSON
// columns.
func (q Quantity) Value() (driver.Value, error) {
	data, err := q.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

// Scan reads lossless JSON text from a SQL driver.
func (q *Quantity) Scan(source any) error {
	var data []byte
	switch value := source.(type) {
	case string:
		if len(value) > MaxSerializedBytes {
			return fmt.Errorf("%w: SQL value exceeds %d bytes", ErrInvalidQuantity, MaxSerializedBytes)
		}
		data = []byte(value)
	case []byte:
		if len(value) > MaxSerializedBytes {
			return fmt.Errorf("%w: SQL value exceeds %d bytes", ErrInvalidQuantity, MaxSerializedBytes)
		}
		data = append([]byte(nil), value...)
	default:
		return fmt.Errorf("%w: unsupported SQL value %T", ErrInvalidQuantity, source)
	}

	return q.UnmarshalJSON(data)
}

// FormatOptions contains every conversion and display decision. No locale or
// preferred unit is inferred.
type FormatOptions struct {
	Unit       Unit
	Conversion ConversionContext
	Scale      int32
	Rounding   decimal.RoundingMode
	Separator  string
}

// Format converts, rounds, and formats a quantity without caller cancellation.
func (q Quantity) Format(options FormatOptions) (string, error) {
	return q.FormatContext(context.Background(), options)
}

// FormatContext converts, rounds, and formats a quantity under caller
// cancellation and explicit options.
func (q Quantity) FormatContext(ctx context.Context, options FormatOptions) (string, error) {
	if err := validateOperationContext(ctx); err != nil {
		return "", err
	}
	if options.Unit == "" || len(options.Separator) > 16 || !utf8.ValidString(options.Separator) ||
		strings.ContainsAny(options.Separator, "\r\n\x00") {
		return "", ErrInvalidQuantity
	}
	converted, err := q.ConvertContext(ctx, options.Unit, options.Conversion)
	if err != nil {
		return "", err
	}
	rounded, err := converted.RoundContext(ctx, options.Scale, options.Rounding)
	if err != nil {
		return "", err
	}

	return rounded.amount.String() + options.Separator + string(options.Unit), nil
}

func (q *Quantity) decode(encoded encodedQuantity) error {
	amount, err := decimal.Parse(encoded.Value)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidQuantity, err)
	}
	decoded, err := New(amount, encoded.Unit)
	if err != nil {
		return err
	}
	*q = decoded

	return nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("%w: trailing JSON value", ErrInvalidQuantity)
	}

	return invalidJSONError(err)
}

type redactedJSONError struct {
	cause error
}

func (e redactedJSONError) Error() string {
	return "invalid JSON value"
}

func (e redactedJSONError) Is(target error) bool {
	return errors.Is(e.cause, target)
}

func (e redactedJSONError) As(target any) bool {
	result, ok := target.(**json.UnmarshalTypeError)
	if !ok {
		return false
	}
	var source *json.UnmarshalTypeError
	if !errors.As(e.cause, &source) {
		return false
	}
	sanitized := *source
	sanitized.Value = "invalid value"
	*result = &sanitized

	return true
}

func invalidJSONError(cause error) error {
	return fmt.Errorf("%w: %w", ErrInvalidQuantity, redactedJSONError{cause: cause})
}

func decodeJSONObject(data []byte, allowed map[string]struct{}, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return invalidJSONError(err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return fmt.Errorf("%w: expected JSON object", ErrInvalidQuantity)
	}

	fields := make(map[string]struct{}, len(allowed))
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return invalidJSONError(err)
		}
		name := token.(string)
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("%w: unknown JSON field", ErrInvalidQuantity)
		}
		if _, duplicate := fields[name]; duplicate {
			return fmt.Errorf("%w: duplicate JSON field", ErrInvalidQuantity)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return invalidJSONError(err)
		}
		fields[name] = struct{}{}
	}
	if _, err := decoder.Token(); err != nil {
		return invalidJSONError(err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return invalidJSONError(err)
	}

	return nil
}

// ParseQuantityXML parses one strict, byte-bounded quantity XML document.
func ParseQuantityXML(data []byte) (Quantity, error) {
	decoder, start, budget, err := startXMLDocument(data, "quantity")
	if err != nil {
		return Quantity{}, err
	}
	encoded, err := decodeQuantityXML(decoder, start, budget)
	if err != nil {
		return Quantity{}, err
	}
	if err := finishXMLDocument(decoder, budget); err != nil {
		return Quantity{}, err
	}

	var quantity Quantity
	if err := quantity.decode(encoded); err != nil {
		return Quantity{}, sanitizeXMLValueError(err)
	}

	return quantity, nil
}

type xmlParseBudget struct {
	depth  int
	tokens int
}

func startXMLDocument(data []byte, root string) (*xml.Decoder, xml.StartElement, *xmlParseBudget, error) {
	if len(data) > MaxSerializedBytes {
		return nil, xml.StartElement{}, nil, invalidXML("document exceeds byte limit")
	}
	// Measurement values never require XML references, directives, comments,
	// or CDATA. Reject them before tokenization so reference expansion and
	// directive payloads cannot amplify parser work.
	if bytes.IndexByte(data, '&') >= 0 || bytes.Contains(data, []byte("<!")) {
		return nil, xml.StartElement{}, nil, invalidXML("unsupported XML construct")
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	budget := &xmlParseBudget{}
	declarationAllowed := true
	for {
		token, err := budget.next(decoder)
		if err != nil {
			return nil, xml.StartElement{}, nil, normalizeXMLTokenError(err)
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Space != "" || value.Name.Local != root || len(value.Attr) != 0 {
				return nil, xml.StartElement{}, nil, invalidXML("invalid root element")
			}

			return decoder, value, budget, nil
		case xml.CharData:
			if len(bytes.TrimSpace(value)) != 0 {
				return nil, xml.StartElement{}, nil, invalidXML("unexpected document text")
			}
		case xml.ProcInst:
			if !declarationAllowed || value.Target != "xml" {
				return nil, xml.StartElement{}, nil, invalidXML("unsupported processing instruction")
			}
			declarationAllowed = false
		}
	}
}

func (b *xmlParseBudget) next(decoder *xml.Decoder) (xml.Token, error) {
	if b.tokens >= MaxXMLTokens {
		return nil, invalidXML("document exceeds token limit")
	}
	token, err := decoder.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}

		return nil, invalidXML("malformed XML")
	}
	b.tokens++
	switch token.(type) {
	case xml.StartElement:
		b.depth++
		if b.depth > MaxXMLDepth {
			return nil, invalidXML("document exceeds depth limit")
		}
	case xml.EndElement:
		b.depth--
	}

	return token, nil
}

func decodeQuantityXML(
	decoder *xml.Decoder,
	start xml.StartElement,
	budget *xmlParseBudget,
) (encodedQuantity, error) {
	var encoded encodedQuantity
	seen := make(map[string]struct{}, 2)
	for {
		token, err := budget.next(decoder)
		if err != nil {
			return encodedQuantity{}, normalizeXMLTokenError(err)
		}
		switch value := token.(type) {
		case xml.StartElement:
			name := value.Name.Local
			if value.Name.Space != "" || len(value.Attr) != 0 || (name != "value" && name != "unit") {
				return encodedQuantity{}, invalidXML("invalid quantity field")
			}
			if _, duplicate := seen[name]; duplicate {
				return encodedQuantity{}, invalidXML("duplicate quantity field")
			}
			seen[name] = struct{}{}
			if name == "value" {
				text, err := decodeXMLText(decoder, value, budget, MaxTextBytes)
				if err != nil {
					return encodedQuantity{}, err
				}
				encoded.Value = text
			} else {
				text, err := decodeXMLText(decoder, value, budget, MaxAliasBytes)
				if err != nil {
					return encodedQuantity{}, err
				}
				encoded.Unit = Unit(text)
			}
		case xml.EndElement:
			if value.Name == start.Name {
				return encoded, nil
			}
		case xml.CharData:
			if len(bytes.TrimSpace(value)) != 0 {
				return encodedQuantity{}, invalidXML("unexpected quantity text")
			}
		default:
			return encodedQuantity{}, invalidXML("unsupported quantity token")
		}
	}
}

// MarshalJSON encodes every side's value and unit plus package quantity.
func (d Dimensions) MarshalJSON() ([]byte, error) {
	if _, err := NewDimensions(d.length, d.width, d.height, d.quantity); err != nil {
		return nil, err
	}

	return json.Marshal(encodedDimensions{
		Length: d.length, Width: d.width, Height: d.height, Quantity: d.quantity,
	})
}

// UnmarshalJSON decodes and validates a complete dimension triple.
func (d *Dimensions) UnmarshalJSON(data []byte) error {
	if d == nil {
		return ErrInvalidQuantity
	}
	if len(data) > MaxSerializedBytes {
		return fmt.Errorf("%w: JSON exceeds %d bytes", ErrInvalidQuantity, MaxSerializedBytes)
	}
	var encoded encodedDimensions
	allowed := map[string]struct{}{"length": {}, "width": {}, "height": {}, "quantity": {}}
	if err := decodeJSONObject(data, allowed, &encoded); err != nil {
		return err
	}
	decoded, err := NewDimensions(encoded.Length, encoded.Width, encoded.Height, encoded.Quantity)
	if err != nil {
		return err
	}
	*d = decoded

	return nil
}

// MarshalXML encodes every side's value and unit plus package quantity.
func (d Dimensions) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if _, err := NewDimensions(d.length, d.width, d.height, d.quantity); err != nil {
		return err
	}
	if start.Name.Local == "" || start.Name.Local == "Dimensions" {
		start.Name.Local = "dimensions"
	}

	return encoder.EncodeElement(encodedDimensions{
		Length: d.length, Width: d.width, Height: d.height, Quantity: d.quantity,
	}, start)
}

// UnmarshalXML fails closed because encoding/xml invokes this callback only
// after its caller-owned decoder has parsed and allocated the start token. Use
// ParseDimensionsXML for byte-bounded XML input.
func (d *Dimensions) UnmarshalXML(_ *xml.Decoder, _ xml.StartElement) error {
	return fmt.Errorf("%w: %w", ErrInvalidQuantity, ErrUnboundedXML)
}

// ParseDimensionsXML parses one strict, byte-bounded dimensions XML document.
func ParseDimensionsXML(data []byte) (Dimensions, error) {
	decoder, start, budget, err := startXMLDocument(data, "dimensions")
	if err != nil {
		return Dimensions{}, err
	}
	encoded, err := decodeDimensionsXML(decoder, start, budget)
	if err != nil {
		return Dimensions{}, err
	}
	if err := finishXMLDocument(decoder, budget); err != nil {
		return Dimensions{}, err
	}
	decoded, err := NewDimensions(encoded.Length, encoded.Width, encoded.Height, encoded.Quantity)
	if err != nil {
		return Dimensions{}, sanitizeXMLValueError(err)
	}

	return decoded, nil
}

func decodeDimensionsXML(
	decoder *xml.Decoder,
	start xml.StartElement,
	budget *xmlParseBudget,
) (encodedDimensions, error) {
	var encoded encodedDimensions
	seen := make(map[string]struct{}, 4)
	for {
		token, err := budget.next(decoder)
		if err != nil {
			return encodedDimensions{}, normalizeXMLTokenError(err)
		}
		switch value := token.(type) {
		case xml.StartElement:
			name := value.Name.Local
			if value.Name.Space != "" || len(value.Attr) != 0 ||
				(name != "length" && name != "width" && name != "height" && name != "quantity") {
				return encodedDimensions{}, invalidXML("invalid dimensions field")
			}
			if _, duplicate := seen[name]; duplicate {
				return encodedDimensions{}, invalidXML("duplicate dimensions field")
			}
			seen[name] = struct{}{}
			switch name {
			case "length":
				encoded.Length, err = decodeXMLQuantityValue(decoder, value, budget)
			case "width":
				encoded.Width, err = decodeXMLQuantityValue(decoder, value, budget)
			case "height":
				encoded.Height, err = decodeXMLQuantityValue(decoder, value, budget)
			case "quantity":
				var text string
				text, err = decodeXMLText(decoder, value, budget, maxPackageQuantityTextBytes)
				if err == nil {
					encoded.Quantity, err = strconv.ParseUint(text, 10, 64)
					if err != nil {
						err = invalidXML("invalid package quantity")
					}
				}
			}
			if err != nil {
				return encodedDimensions{}, err
			}
		case xml.EndElement:
			if value.Name == start.Name {
				return encoded, nil
			}
		case xml.CharData:
			if len(bytes.TrimSpace(value)) != 0 {
				return encodedDimensions{}, invalidXML("unexpected dimensions text")
			}
		default:
			return encodedDimensions{}, invalidXML("unsupported dimensions token")
		}
	}
}

func decodeXMLQuantityValue(
	decoder *xml.Decoder,
	start xml.StartElement,
	budget *xmlParseBudget,
) (Quantity, error) {
	encoded, err := decodeQuantityXML(decoder, start, budget)
	if err != nil {
		return Quantity{}, err
	}
	var quantity Quantity
	if err := quantity.decode(encoded); err != nil {
		return Quantity{}, sanitizeXMLValueError(err)
	}

	return quantity, nil
}

func decodeXMLText(
	decoder *xml.Decoder,
	_ xml.StartElement,
	budget *xmlParseBudget,
	maxBytes int,
) (string, error) {
	var text []byte
	for {
		token, err := budget.next(decoder)
		if err != nil {
			return "", normalizeXMLTokenError(err)
		}
		switch value := token.(type) {
		case xml.CharData:
			if len(value) > maxBytes-len(text) {
				return "", invalidXML("XML field exceeds byte limit")
			}
			text = append(text, value...)
		case xml.EndElement:
			return string(text), nil
		default:
			return "", invalidXML("XML field must contain text only")
		}
	}
}

func finishXMLDocument(decoder *xml.Decoder, budget *xmlParseBudget) error {
	for {
		token, err := budget.next(decoder)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return normalizeXMLTokenError(err)
		}
		if text, ok := token.(xml.CharData); ok && len(bytes.TrimSpace(text)) == 0 {
			continue
		}

		return invalidXML("trailing XML content")
	}
}

func sanitizeXMLValueError(err error) error {
	switch {
	case errors.Is(err, ErrUnknownUnit):
		return ErrUnknownUnit
	case errors.Is(err, ErrDimensionMismatch):
		return ErrDimensionMismatch
	default:
		return ErrInvalidQuantity
	}
}

func normalizeXMLTokenError(err error) error {
	if errors.Is(err, io.EOF) {
		return invalidXML("unexpected end of XML document")
	}

	return err
}

func invalidXML(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidQuantity, reason)
}
