# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-measurement/v3

`measurement` is an immutable, exact, unit-safe measurement package for Track, Postal, Location, and logistics services. It uses github.com/faustbrian/go-math/decimal exclusively and never requires binary floating-point conversion.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-measurement/v3`](../doc.go) | Use for immutable quantities, dimensions, exact or explicitly rounded conversions, and logistics formulas. |
| [`github.com/faustbrian/go-measurement/v3/adapters/wire`](../adapters/wire/wire.go) | Use for bounded JSON and XML encoding or decoding through the Golib wire contracts. |
| [`github.com/faustbrian/go-measurement/v3/measurementwire`](../measurementwire/wire.go) | Deprecated compatibility path for bounded wire encoding; use adapters/wire. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
