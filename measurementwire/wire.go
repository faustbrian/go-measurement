// Package measurementwire is the compatibility facade for the bounded
// measurement wire adapter.
//
// Deprecated: use github.com/faustbrian/go-measurement/v2/adapters/wire,
// introduced in v1.1.0 with the same Options, Encode, and Decode usage. The
// earliest permitted removal is a future major release and never before
// 2027-03-08.
package measurementwire

import (
	measurement "github.com/faustbrian/go-measurement/v2"
	adapter "github.com/faustbrian/go-measurement/v2/adapters/wire"
	"github.com/faustbrian/go-wire"
)

// Options bounds encoded and decoded documents. XML decoding also enforces
// measurement's smaller core document ceiling. Options remains a distinct
// named compatibility type so reflection-based consumers retain the legacy
// package identity.
type Options struct {
	MaxBytes int64
}

// Encode serializes quantity in an explicitly selected supported format.
func Encode(quantity measurement.Quantity, format wire.Format, options Options) ([]byte, error) {
	return adapter.Encode(quantity, format, adapter.Options{MaxBytes: options.MaxBytes})
}

// Decode parses exactly one bounded quantity document in an explicitly
// selected supported format.
func Decode(payload []byte, format wire.Format, options Options) (measurement.Quantity, error) {
	return adapter.Decode(payload, format, adapter.Options{MaxBytes: options.MaxBytes})
}
