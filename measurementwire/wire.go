// Package measurementwire is the compatibility facade for the bounded
// measurement wire adapter. Encode and Decode delegate to adapters/wire without
// owning transport, resources or background work. Calls can run concurrently;
// payload buffers are not retained. No cgo or service dependency is introduced.
//
// Deprecated: use github.com/faustbrian/go-measurement/v3/adapters/wire,
// introduced in v1.1.0 with the same Options, Encode, and Decode usage. The
// earliest permitted removal is a future major release and never before
// 2027-03-08.
package measurementwire

import (
	measurement "github.com/faustbrian/go-measurement/v3"
	adapter "github.com/faustbrian/go-measurement/v3/adapters/wire"
	"github.com/faustbrian/go-wire/v3"
)

// Options bounds encoded and decoded documents. Core decoding independently
// enforces measurement's document ceiling. Options remains a distinct
// named compatibility type so reflection-based consumers retain the legacy
// package identity.
type Options struct {
	// MaxBytes selects the codec byte ceiling. Zero uses its defaults; negative
	// values are invalid. Core decoding keeps its own stricter document ceiling.
	MaxBytes int64
}

// Encode delegates JSON/XML serialization to adapters/wire.Encode, including
// its byte budget and Wire v3 error classification. Returned bytes are caller-owned.
func Encode(quantity measurement.Quantity, format wire.Format, options Options) ([]byte, error) {
	return adapter.Encode(quantity, format, adapter.Options{MaxBytes: options.MaxBytes})
}

// Decode parses exactly one bounded quantity document in an explicitly
// selected JSON/XML format, preserving the successor's error and byte-limit
// semantics. Failure returns the zero Quantity; payload is not retained.
func Decode(payload []byte, format wire.Format, options Options) (measurement.Quantity, error) {
	return adapter.Decode(payload, format, adapter.Options{MaxBytes: options.MaxBytes})
}
