// Package measurementwire adapts Quantity to bounded wire JSON and XML codecs
// while preserving decimal and unit metadata. Start with Encode/Decode and an
// explicitly selected Wire v3 format. This package does not infer a format, own
// transport I/O or authenticate documents. Calls are stateless and can run
// concurrently; payloads must not be mutated during decoding and are not
// retained. No cgo, external
// service or background lifecycle is introduced.
package measurementwire

import (
	"encoding/json"
	"errors"
	"fmt"

	measurement "github.com/faustbrian/go-measurement/v3"
	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/jsonwire"
	"github.com/faustbrian/go-wire/v3/xmlwire"
)

// Options bounds encoded and decoded documents. Zero values retain the
// selected wire codec's defaults. Core decoding independently enforces
// measurement's document ceiling.
type Options struct {
	// MaxBytes is the selected codec's byte ceiling; zero uses its defaults and
	// negative values are invalid. Core decoding still enforces MaxSerializedBytes
	// independently, so a larger codec ceiling does not widen the core schema.
	MaxBytes int64
}

// Encode serializes quantity as JSON or XML, preserving decimal text and unit.
// Other formats return an error matching wire.ErrUnsupportedFormat. The byte
// ceiling governs encoding through the selected codec; returned bytes are
// caller-owned. Errors retain Wire v3 classification.
func Encode(quantity measurement.Quantity, format wire.Format, options Options) ([]byte, error) {
	switch format {
	case wire.FormatJSON:
		return jsonwire.Encode(quantity, jsonwire.EncodeOptions{MaxBytes: options.MaxBytes})
	case wire.FormatXML:
		return xmlwire.Encode(quantity, xmlwire.EncodeOptions{MaxBytes: options.MaxBytes})
	case wire.FormatSOAP, wire.FormatYAML, wire.FormatTOML, wire.FormatMessagePack,
		wire.FormatCBOR, wire.FormatBSON:
		return nil, unsupported("encode", format)
	default:
		return nil, unsupported("encode", format)
	}
}

// Decode parses exactly one bounded quantity document in an explicitly
// selected JSON/XML format. The returned Quantity does not retain payload.
// Unsupported formats match wire.ErrUnsupportedFormat. Codec size failures
// match wire.ErrSizeLimit; XML core syntax errors match wire.ErrParse and core
// schema/value errors match wire.ErrValidation. JSON follows the selected codec
// classification, with redacted causes and sanitized type-error metadata. All
// errors return the zero Quantity. Bound transport input before materializing it.
func Decode(payload []byte, format wire.Format, options Options) (measurement.Quantity, error) {
	var quantity measurement.Quantity
	var err error
	switch format {
	case wire.FormatJSON:
		err = sanitizeJSONWireError(jsonwire.Decode(payload, &quantity, jsonwire.DecodeOptions{
			MaxBytes:              options.MaxBytes,
			DisallowUnknownFields: true,
		}))
	case wire.FormatXML:
		if err = validateXMLSize(payload, options.MaxBytes); err == nil {
			quantity, err = measurement.ParseQuantityXML(payload)
			if err != nil {
				kind := wire.ErrorKindValidation
				if errors.Is(err, measurement.ErrMalformedXML) {
					kind = wire.ErrorKindParse
				}
				err = &wire.Error{
					Kind:   kind,
					Format: wire.FormatXML,
					Op:     "decode",
					Err:    err,
				}
			}
		}
	case wire.FormatSOAP, wire.FormatYAML, wire.FormatTOML, wire.FormatMessagePack,
		wire.FormatCBOR, wire.FormatBSON:
		err = unsupported("decode", format)
	default:
		err = unsupported("decode", format)
	}
	if err != nil {
		return measurement.Quantity{}, err
	}

	return quantity, nil
}

// Preserve sentinel classification without Unwrap: raw codec diagnostics can
// disclose payload text. As exposes only a copy of supported type-error metadata
// with the Value field redacted; syntax errors are deliberately not forwarded.
type redactedJSONCause struct {
	cause error
}

// Error renders a fixed classification without the original JSON diagnostic.
func (e redactedJSONCause) Error() string {
	return "invalid JSON payload"
}

// Is retains cause sentinel identity without exposing it through Unwrap.
func (e redactedJSONCause) Is(target error) bool {
	return errors.Is(e.cause, target)
}

// As returns a sanitized copy only for json.UnmarshalTypeError; original
// syntax errors and unsupported typed causes are deliberately not forwarded.
func (e redactedJSONCause) As(target any) bool {
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

func sanitizeJSONWireError(err error) error {
	if err == nil {
		return nil
	}
	var codecError *wire.Error
	if !errors.As(err, &codecError) {
		return &wire.Error{
			Kind:   wire.ErrorKindParse,
			Format: wire.FormatJSON,
			Op:     "decode",
			Err:    redactedJSONCause{cause: err},
		}
	}

	return &wire.Error{
		Kind:   codecError.Kind,
		Format: codecError.Format,
		Op:     codecError.Op,
		Err:    redactedJSONCause{cause: codecError.Err},
	}
}

func validateXMLSize(payload []byte, configured int64) error {
	if configured < 0 {
		return &wire.Error{
			Kind:   wire.ErrorKindValidation,
			Format: wire.FormatXML,
			Op:     "decode options",
			Err:    errors.New("max bytes must not be negative"),
		}
	}
	maxBytes := configured
	if maxBytes == 0 {
		maxBytes = xmlwire.DefaultMaxBytes
	}
	if int64(len(payload)) > maxBytes {
		return &wire.Error{
			Kind:   wire.ErrorKindSizeLimit,
			Format: wire.FormatXML,
			Op:     "decode",
			Err:    xmlwire.ErrPayloadTooLarge,
		}
	}

	return nil
}

func unsupported(operation string, format wire.Format) error {
	return &wire.Error{
		Kind:   wire.ErrorKindUnsupported,
		Format: format,
		Op:     operation,
		Err:    fmt.Errorf("measurement adapter supports JSON and XML"),
	}
}
