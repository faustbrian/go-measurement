package measurement

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRedactedJSONErrorKeepsClassificationWithoutTypedPayload(t *testing.T) {
	t.Parallel()

	const controlled = "private-marker"
	cause := &json.UnmarshalTypeError{
		Value:  controlled,
		Type:   reflect.TypeOf(""),
		Offset: 7,
	}
	err := invalidJSONError(cause)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("error = %v, want ErrInvalidQuantity", err)
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
	if errors.As(invalidJSONError(errors.New(controlled)), &typed) {
		t.Fatal("non-type error became an UnmarshalTypeError")
	}
}
