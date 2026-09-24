package measurementwire

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire"
)

func TestSanitizedJSONWireErrorPreservesClassWithoutTypedPayload(t *testing.T) {
	t.Parallel()

	const controlled = "private-marker"
	cause := &json.UnmarshalTypeError{
		Value:  controlled,
		Type:   reflect.TypeOf(""),
		Offset: 7,
	}
	err := sanitizeJSONWireError(&wire.Error{
		Kind:   wire.ErrorKindParse,
		Format: wire.FormatJSON,
		Op:     "decode",
		Err:    cause,
	})
	if !errors.Is(err, wire.ErrParse) {
		t.Fatalf("error = %v, want wire.ErrParse", err)
	}

	var typed *json.UnmarshalTypeError
	if !errors.As(err, &typed) {
		t.Fatalf("error = %v, want sanitized UnmarshalTypeError", err)
	}
	if typed.Value != "invalid value" || strings.Contains(typed.Error(), controlled) {
		t.Fatalf("typed error exposed controlled input: %#v", typed)
	}
	if cause.Value != controlled {
		t.Fatal("sanitization mutated the source error")
	}

	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		t.Fatalf("error = %v, want no SyntaxError", err)
	}
	if errors.As(sanitizeJSONWireError(errors.New(controlled)), &typed) {
		t.Fatal("non-type error became an UnmarshalTypeError")
	}
}

func TestBareJSONCodecErrorIsClassifiedAndRedacted(t *testing.T) {
	t.Parallel()

	const controlled = "private-marker"
	cause := errors.New(controlled)
	err := sanitizeJSONWireError(cause)
	if !errors.Is(err, wire.ErrParse) || !errors.Is(err, cause) {
		t.Fatalf("error = %v, want parse class and original cause identity", err)
	}
	if strings.Contains(err.Error(), controlled) {
		t.Fatalf("error echoed controlled input: %v", err)
	}
}
