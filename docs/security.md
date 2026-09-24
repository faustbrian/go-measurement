# Security threat model

Version: 1.0 (2026-09-13). Owner: go-measurement maintainers.

## Assets, attackers, and trust boundaries

The protected assets are exact quantities, unit identity, dimensional
invariants, process availability, and operational metadata. Callers may supply
hostile decimal text, unit aliases, JSON, XML, SQL values, package counts,
conversion scales, and formula constants. The package crosses decimal-parser,
serializer, SQL-driver, and optional wire-codec boundaries. It has no network,
filesystem, process, environment, plugin, credential, randomness, background
goroutine, or global mutable-registry access.

Attackers may attempt oversized coefficients, exponents, documents, fields,
aliases, or diagnostics; deep or high-token XML; entity/reference expansion;
duplicate or ambiguous fields; invalid dimensions; or expensive division.
Authentication, authorization, SSRF, redirects, path traversal, SQL statement
construction, replay, ordering, and secret storage are outside this module's
owned behavior.

## Enforced controls

- `MaxTextBytes`, `MaxAliasBytes`, and `MaxProfileAliases` bound text and unit
  profile work. `math.Limits` bounds coefficients, exponents, intermediates,
  and outputs.
- `MaxSerializedBytes` bounds direct JSON, SQL, and the supported XML parse
  entrypoints. XML additionally uses fixed field counts, `MaxXMLDepth`, and
  `MaxXMLTokens`; it rejects attributes, namespaces, comments, directives,
  processing instructions, CDATA, references, duplicates, nested scalar
  content, and trailing documents.
- `Quantity.UnmarshalXML` and `Dimensions.UnmarshalXML` fail closed with
  `ErrUnboundedXML`. The `encoding/xml` callback is reached only after its
  caller-owned decoder has parsed the start token, so it cannot impose the
  required pre-token byte limit. Use `ParseQuantityXML`,
  `ParseDimensionsXML`, or `adapters/wire` instead.
- XML diagnostics expose fixed failure classes and reasons, not element names,
  values, or decoder syntax details. Callers should log error classes only.
- JSON, SQL, profile, and unit diagnostics preserve classification while using
  fixed reasons rather than echoing attacker-controlled field names, aliases,
  or unit values.
- The standard library may reject malformed JSON before dispatching a custom
  `UnmarshalJSON` method. Use the bounded package or wire entrypoint directly
  for untrusted bytes; an outer `json.Unmarshal` error is caller-owned and must
  not be logged without redaction.
- Unit and dimensional validation happens before arithmetic. Nonpositive
  dimensions, truck widths, stacking factors, divisors, and indexes fail
  closed. Conversion contexts cannot infer rounding policy.
- Parsing operates on already-resident byte slices and performs no blocking
  external operation, so context cancellation is not applicable. Arithmetic
  APIs that can compose material work provide context-aware variants.

## Deployment guidance

Use tighter arithmetic limits where an application's public contract permits
them. Treat unit profiles, carrier divisors, and accepted payload limits as
versioned caller configuration. Apply an HTTP or transport body limit before
reading a payload into memory; the package limit begins at the supplied byte
slice. Do not create an `xml.Decoder` over an untrusted stream and decode
directly into `Quantity` or `Dimensions`.

## Accepted risks

| Severity | Owner | Risk and rationale | Mitigation | Review condition |
| --- | --- | --- | --- | --- |
| Medium | Integrating application owner | `encoding/xml` parses the root start token before dispatching `UnmarshalXML`; this library cannot prevent allocation performed by that caller-owned decoder. Retaining the methods provides an explicit migration error instead of silently decoding through an unbounded callback. | Direct callbacks fail closed. Bound transport reads first and use `ParseQuantityXML`, `ParseDimensionsXML`, or `adapters/wire`. | Revisit if Go exposes a pre-token decoder limit, the callback methods can be removed in a future major version, or a supported adapter bypasses the bounded entrypoints. |
| Medium | Integrating application owner | A SQL driver materializes a `string` or `[]byte` before invoking `Scan`, outside this package's control. | Apply database and driver response limits. `Scan` rejects oversized values before making its own copy or parsing JSON. | Revisit if supported drivers expose streaming or pre-materialization limits, or an adapter bypasses `Scan`. |
| Medium | Integrating application owner | `encoding/json.Unmarshal` can reject malformed JSON before invoking `Quantity.UnmarshalJSON` or `Dimensions.UnmarshalJSON`, so the standard-library error is outside this package's redaction boundary. | Apply a transport byte limit and call the package's bounded `UnmarshalJSON` method or `adapters/wire.Decode` for untrusted payloads; log fixed error classes rather than raw outer decoder errors. | Revisit if Go adds a pre-validation redaction hook or a supported adapter returns an unsanitized outer error. |
