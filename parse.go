package measurement

import (
	"fmt"
	"strings"

	"github.com/faustbrian/go-math/decimal"
)

const (
	// MaxTextBytes bounds canonical and profile-based text parsing.
	MaxTextBytes = 4096
	// MaxAliasBytes bounds unit tokens resolved through a profile.
	MaxAliasBytes = 128
	// MaxProfileAliases bounds caller-defined unit-token catalogs.
	MaxProfileAliases = 1024
)

// Profile is an immutable explicit mapping of accepted unit tokens, safe for
// concurrent resolution. Its zero value resolves no aliases; no default or
// locale policy is inferred.
type Profile struct {
	aliases map[string]Unit
}

// NewProfile validates and defensively copies an explicit alias-to-unit policy.
// Nil/empty input creates an empty profile. Alias count must not exceed
// MaxProfileAliases; each alias must have 1..MaxAliasBytes bytes and identify a
// supported unit. Alias text is kept exactly, without case/whitespace normalization.
// Callers must not mutate the input map during construction.
func NewProfile(aliases map[string]Unit) (Profile, error) {
	if len(aliases) > MaxProfileAliases {
		return Profile{}, fmt.Errorf("%w: profile exceeds %d aliases", ErrInvalidQuantity, MaxProfileAliases)
	}
	copyOfAliases := make(map[string]Unit, len(aliases))
	for alias, unit := range aliases {
		if len(alias) == 0 || len(alias) > MaxAliasBytes {
			return Profile{}, fmt.Errorf("%w: unit alias length", ErrUnknownUnit)
		}
		if _, err := definitionFor(unit); err != nil {
			return Profile{}, err
		}
		copyOfAliases[alias] = unit
	}

	return Profile{aliases: copyOfAliases}, nil
}

// SymbolProfile accepts only canonical unit symbols.
func SymbolProfile() Profile {
	aliases := make(map[string]Unit, len(unitDefinitions))
	for unit := range unitDefinitions {
		aliases[string(unit)] = unit
	}

	return Profile{aliases: aliases}
}

// Resolve resolves exactly one alias without case or locale inference.
func (p Profile) Resolve(alias string) (Unit, error) {
	if len(alias) == 0 {
		return "", fmt.Errorf("%w: unit alias length", ErrUnknownUnit)
	}
	if len(alias) > MaxAliasBytes {
		return "", fmt.Errorf("%w: unit alias length", ErrUnknownUnit)
	}
	unit, ok := p.aliases[alias]
	if !ok {
		return "", fmt.Errorf("%w: unsupported unit alias", ErrUnknownUnit)
	}
	if _, err := definitionFor(unit); err != nil {
		return "", err
	}

	return unit, nil
}

// Parse parses decimal amount, ASCII space and exact alias text under profile.
// It rejects surrounding whitespace and text over MaxTextBytes; profiles do not
// infer missing units or canonical aliases. Decimal syntax/resource errors wrap
// ErrInvalidQuantity and the go-math cause; unsupported aliases match ErrUnknownUnit.
// Signed and zero amounts are accepted. Failure returns the zero Quantity.
func Parse(input string, profile Profile) (Quantity, error) {
	if len(input) > MaxTextBytes {
		return Quantity{}, fmt.Errorf("%w: text exceeds %d bytes", ErrInvalidQuantity, MaxTextBytes)
	}
	if strings.TrimSpace(input) != input {
		return Quantity{}, ErrInvalidQuantity
	}
	separator := strings.LastIndexByte(input, ' ')
	if separator == -1 {
		return Quantity{}, ErrInvalidQuantity
	}
	amount, err := decimal.Parse(input[:separator])
	if err != nil {
		return Quantity{}, fmt.Errorf("%w: %w", ErrInvalidQuantity, err)
	}
	unit, err := profile.Resolve(input[separator+1:])
	if err != nil {
		return Quantity{}, err
	}

	return New(amount, unit)
}
