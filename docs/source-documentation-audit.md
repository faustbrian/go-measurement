# Source documentation audit

## Scope and plan

Baseline: `d16d64970497863ce1c07d9b63f4c38d2d884f8d` (main).
Owner: go-measurement source-documentation pass. Tier A: comments, their active
rendered API snapshot and directly affected prose only. Preserve declarations,
executable tokens, dependency and pinned specification inputs, tests and examples.
The module is `github.com/faustbrian/go-measurement/v3`; packages are root
measurement, adapters/wire (named measurementwire) and the deprecated
measurementwire facade. Direct consumers recorded in modules.json include
Knapsack and Rule Engine's Measurement adapter; unchanged signatures and runtime
do not warrant their execution or a module release.

Named documentation risks: unknown unit admission, zero/signed quantity policy,
exact versus rounded ratios, affine offsets, per-operation rounding in composed
formulas, context ownership, mutable decode receivers and parser budgets.
Smallest sufficient checks: comment-aware executable equivalence, Go formatting,
objective AST comments, rendered documentation/API snapshot stability, native
documentation examples, architecture/security structure, changed local links,
marker/generator/deprecation/secret scans and one ordinary complete-diff review.
Shared persistent documentation enforcement remains coordinator/tooling-owned.

## Reproducible file and package inventory

`git ls-files '*go.mod' '*.go'` identifies one module and 31 tracked Go files.
Every file and declaration was read, including unexported helpers, fields and tests.
Package clauses and modules.json identify three packages; external test package
clauses are not new public packages. Scan special source markers with
`rg -n 'TO[D]O|FI[X]ME|HA[C]K|NO[T]E|WAR[N]ING|Deprecated:|Code generated|go:generate|go:build' --glob '*.go'`.
No build/platform/cgo variants, generated source, generation directives or debt
markers exist. The facade's package deprecation has valid paragraph syntax,
a successor and retained earliest-removal date (2027-03-08).

| File | Package clause | Role | Exported types/functions/methods inspected |
| --- | --- | --- | --- |
| `adapter_successor_test.go` | `measurement_test` | tests, examples or assurance | `TestWireAdaptersExposeCurrentWireErrorIdentity`, `TestWireSuccessorPreservesBoundedSerializationContract`, `TestLegacyWireRemainsDistinctDelegatingCompatibilityFacade` |
| `adapters/wire/diagnostics_internal_test.go` | `measurementwire` | tests, examples or assurance | `TestSanitizedJSONWireErrorPreservesClassWithoutTypedPayload`, `TestBareJSONCodecErrorIsClassifiedAndRedacted` |
| `adapters/wire/wire.go` | `measurementwire` | API and implementation | `Options`, `Encode`, `Decode`, `Error`, `Is`, `As` |
| `adapters/wire/wire_test.go` | `measurementwire_test` | tests, examples or assurance | `TestUnwrappedJSONCauseRemainsRedacted`, `TestJSONAndXMLRoundTripsPreserveUnitMetadata`, `TestAdapterRejectsUnsupportedFormatsAndOversizePayloads`, `TestXMLDecodeAcceptsExactConfiguredLimit`, `TestXMLDecodeClassifiesSyntaxAndValueErrors`, `TestDecodeDoesNotRetainCallerPayload`, `TestXMLDecodeUsesBoundedDefault` |
| `benchmark_test.go` | `measurement_test` | tests, examples or assurance | `BenchmarkConvertMetresToFeet`, `BenchmarkDimensionsLoadingMetres` |
| `catalog_test.go` | `measurement_test` | tests, examples or assurance | `TestUnitCatalogIsStableSortedAndAliasSafe`, `TestProfileDefensivelyCopiesAndResolvesAliases`, `TestProfileConstructionRejectsUnboundedOrInvalidCatalogs`, `TestQuantityEqualityConvertsCompatibleUnits` |
| `context_test.go` | `measurement_test` | tests, examples or assurance | `TestContextFirstSurfaceCompiles`, `TestContextFirstMethodsCheckContextBeforeInvalidArguments`, `TestContextFirstMethodsMatchLegacyResults`, `TestConvertContextPreservesCancellationIdentityAtQuotient`, `TestContextCancellationStopsComposedAndPureDecimalWork`, `TestContextIsNotRetainedAcrossConcurrentCalls`, `Deadline`, `Done`, `Value`, `Err` |
| `dimension.go` | `measurement` | API and implementation | `Dimension`, `String` |
| `dimensions.go` | `measurement` | API and implementation | `Dimensions`, `TruckWidth`, `StackingFactor`, `VolumetricDivisor`, `VolumetricIndex`, `NewDimensions`, `Length`, `Width`, `Height`, `Quantity`, `FloorArea`, `FloorAreaContext`, `CubicVolume`, `CubicVolumeContext`, `TotalVolume`, `TotalVolumeContext`, `NewTruckWidth`, `Quantity`, `NewStackingFactor`, `Decimal`, `LoadingMetres`, `LoadingMetresContext`, `NewVolumetricDivisor`, `VolumePerKilogram`, `VolumeUnit`, `Weight`, `WeightContext`, `NewVolumetricIndex`, `Density`, `Weight`, `WeightContext` |
| `dimensions_test.go` | `measurement_test` | tests, examples or assurance | `TestDimensionsCalculatePackageAndShipmentVolume`, `TestDimensionsCalculateLoadingMetresWithStacking`, `TestVolumetricDivisorUsesExplicitVolumeUnit`, `TestLogisticsValuesRejectInvalidDimensions` |
| `doc.go` | `measurement` | API and implementation |  |
| `encoding.go` | `measurement` | API and implementation | `FormatOptions`, `MarshalJSON`, `UnmarshalJSON`, `MarshalXML`, `UnmarshalXML`, `MarshalText`, `UnmarshalText`, `Value`, `Scan`, `Format`, `FormatContext`, `Error`, `Is`, `As`, `ParseQuantityXML`, `MarshalJSON`, `UnmarshalJSON`, `MarshalXML`, `UnmarshalXML`, `ParseDimensionsXML` |
| `encoding_diagnostics_internal_test.go` | `measurement` | tests, examples or assurance | `TestRedactedJSONErrorKeepsClassificationWithoutTypedPayload` |
| `encoding_test.go` | `measurement_test` | tests, examples or assurance | `TestQuantityJSONPreservesDecimalAndUnitMetadata`, `TestQuantityCodecsRejectAmbiguousFields`, `TestQuantityXMLPreservesDecimalAndUnitMetadata`, `TestQuantitySQLValueAndScannerRoundTrip`, `TestQuantitySQLScannerRejectsOversizeBeforeCopy`, `TestQuantitySQLScannerAcceptsExactSerializedByteLimit`, `TestDirectDecodersRejectOversizeDocumentsBeforeParsing`, `TestFailedJSONAndSQLDecodesPreserveQuantityForRetry`, `TestFailedDimensionsJSONDecodePreservesReceiverForRetry`, `TestFormatRequiresExplicitTargetAndRounding`, `TestDimensionsJSONAndXMLRoundTripAllUnits`, `TestDimensionsCodecsRejectAmbiguousFields` |
| `example_test.go` | `measurement_test` | tests, examples or assurance | `ExampleQuantity_Convert`, `ExampleQuantity_ConvertContext`, `ExampleDimensions_LoadingMetres`, `ExampleVolumetricDivisor_Weight` |
| `fuzz_test.go` | `measurement_test` | tests, examples or assurance | `FuzzParseAndTextRoundTrip`, `FuzzQuantityJSON`, `FuzzQuantityXMLAndSQL` |
| `hardening_internal_test.go` | `measurement` | tests, examples or assurance | `TestInternalDimensionAndUnitFailurePaths`, `TestInternalContextAndQuantityFailurePaths`, `TestInternalQuantityArithmeticResourcePaths`, `TestInternalDimensionTripleFailurePaths`, `TestDecodeJSONObjectRejectsUnexpectedEOF`, `TestInternalEncodingFailurePaths`, `TestInternalStrictCodecTokenPaths`, `TestXMLParseBudgetRejectsExcessDepthAndTokens`, `TestInternalProfileBoundaryPaths` |
| `logistics_fixtures_test.go` | `measurement_test` | tests, examples or assurance | `TestDBSchenkerEuroPalletLoadingMetreFixture`, `TestCarrierVolumetricWeightFixtures`, `TestDSVEuroPalletVolumetricIndexFixture` |
| `measurementwire/wire.go` | `measurementwire` | API and implementation | `Options`, `Encode`, `Decode` |
| `measurementwire/wire_test.go` | `measurementwire_test` | tests, examples or assurance | `TestUnwrappedJSONCauseRemainsRedacted`, `TestJSONAndXMLRoundTripsPreserveUnitMetadata`, `TestAdapterRejectsUnsupportedFormatsAndOversizePayloads`, `TestLegacyXMLDecodeMatchesExactConfiguredLimit`, `TestLegacyXMLDecodeClassifiesSyntaxAndValueErrors`, `TestDecodeErrorsDoNotEchoPayload` |
| `mutation_boundaries_internal_test.go` | `measurement` | tests, examples or assurance | `TestDimensionClosedBoundary`, `TestPackageCountAcceptsExactMaximum`, `TestSerializedInputsAcceptExactMaximum`, `TestFormatOptionBoundaries`, `TestStrictJSONBoundaryDiagnostics`, `TestDimensionsXMLStartElementBoundaries`, `TestProfileAndParseExactLimits`, `TestRoundedContextAcceptsExactExponentLimits` |
| `operations_test.go` | `measurement_test` | tests, examples or assurance | `TestOfficialExactUnitDefinitions`, `TestCompatibleArithmeticComparisonClampAndCount`, `TestDivideMassByVolumeAndMultiplyDensityByVolume`, `TestVolumetricIndexCalculatesWeight`, `TestCountAndDerivedDimensionBounds`, `TestAbsoluteTemperaturesRejectAffineArithmetic`, `TestZeroConversionContextDoesNotInferRoundingPolicy`, `TestTextAndSerializationInputsAreBounded`, `TestConstructorsRejectDecimalsBeyondDefaultResourceLimits` |
| `parse.go` | `measurement` | API and implementation | `Profile`, `NewProfile`, `SymbolProfile`, `Resolve`, `Parse` |
| `property_test.go` | `measurement_test` | tests, examples or assurance | `TestConversionRoundTripsAcrossEveryCompatibleUnitPair`, `TestDerivedDimensionIdentitiesAndCommutativity` |
| `quantity.go` | `measurement` | API and implementation | `ConversionContext`, `Quantity`, `ExactConversion`, `RoundedConversion`, `WithLimits`, `New`, `MustNew`, `Amount`, `Unit`, `Dimension`, `String`, `Convert`, `ConvertContext`, `Add`, `AddContext`, `Subtract`, `SubtractContext`, `Compare`, `CompareContext`, `Equal`, `EqualContext`, `Multiply`, `MultiplyContext`, `Divide`, `DivideContext`, `Round`, `RoundContext`, `Times`, `TimesContext`, `Clamp`, `ClampContext` |
| `quantity_test.go` | `measurement_test` | tests, examples or assurance | `TestExactLengthConversionRetainsOperands`, `TestIncompatibleDimensionsCannotBeCombined`, `TestDerivedDimensionsSelectCanonicalUnits`, `TestTemperatureConversionRequiresExplicitRounding`, `TestParsingUsesOnlyExplicitProfileAliases`, `TestRoundedConversionRoundsOnlyTheFinalRatio` |
| `race_test.go` | `measurement_test` | tests, examples or assurance | `TestQuantitiesContextsAndProfilesAreSafeForConcurrentReuse` |
| `security_diagnostics_test.go` | `measurement_test` | tests, examples or assurance | `TestPublicJSONDiagnosticsSanitizeTypedErrors`, `TestInputDiagnosticsDoNotEchoControlledValues`, `TestJSONDiagnosticsDoNotEchoNumericPayloads` |
| `security_xml_test.go` | `measurement_test` | tests, examples or assurance | `TestDirectXMLUnmarshalFailsClosed`, `TestBoundedXMLParsersPreserveMetadata`, `TestBoundedXMLParsersRejectHostileWork`, `TestXMLScalarLimitsDistinguishBoundaryFromSemanticFailure`, `TestBoundedDimensionsXMLRejectsInvalidValuesAndTrailingWork`, `TestBoundedQuantityXMLClassifiesUnknownUnitWithoutEchoingIt`, `TestBoundedXMLDiagnosticsDoNotEchoInput` |
| `unit.go` | `measurement` | API and implementation | `Unit`, `String`, `Dimension`, `Units` |
| `unit_definitions_test.go` | `measurement_test` | tests, examples or assurance | `TestEveryFiniteUnitRatioAgainstCanonicalUnit`, `TestFahrenheitUsesExactDefinedRatioBeforeRounding`, `TestEveryUnitHasTheAuditedSymbolAndDimension` |

The inventory also covers all unit/dimension constants and error sentinels in
unit.go/dimension.go, all bounds in parse.go/encoding.go/dimensions.go,
FormatOptions' five fields and both Options.MaxBytes fields. No public interface
exists; SQL/JSON/XML/text obligations are expressed through receiver methods.

## Initial gaps

- Existing package comments need entry points, non-goals, concurrency/receiver
  mutation and infrastructure boundaries. Do not imply the wire facade and
  adapters share nominal Options identity.
- Missing idiomatic comments: seven dimension constants; most unit constants;
  Dimension.String, Unit.String and Quantity.String; five error sentinels;
  five FormatOptions fields and Options.MaxBytes in both adapter packages.
- MustNew incorrectly mentions only unknown units although any New error panics.
  Quantity permits numeric zero/negative values and does not enforce physically
  valid absolute temperature. Constructors for logistics inputs add positivity.
- ExactConversion/RoundedConversion need zero-context, delayed validation,
  same-unit no-op and target affine-offset caveats. Derived/composed operations
  round at each conversion/division boundary, not once across the entire formula.
- Context methods need non-nil admission, cancellation-before-validation and no
  retained context/no partial result. No forced interruption claim is justified.
- Decoder methods need nil/SQL NULL/ownership/failure-preservation contracts;
  raw XML callbacks reject all decoding and direct JSON redaction does not govern
  errors produced before callback dispatch by encoding/json.
- JSON/XML duplicate checks, redacted error introspection and exact affine ratio
  placement deserve adjacent rationale. Existing XML pre-token resource rationale
  is useful and remains.
- Fuzz inputs/invariants, benchmark workloads/setup and deterministic
  cancellation helper rationale need concise comments. Existing examples and
  clear regression/property assertions need no narration.

## Investigated rationale versus unresolved questions

Source and unit_definitions_test.go/operations_test.go bind unit factors and
conversion placement to the existing specification decisions: BIPM SI Brochure
9th edition v4.01, NIST SP 811 (2008), and carrier snapshots dated 2026-08-30.
The finite supported catalog, ASCII identifiers and logistics formula policy
are implementation decisions, not an unsupported completeness claim.
See [specification decisions](specification-decisions.md) and
[security model](security.md) for authoritative boundaries and residual risks.

quantity.go validates context and conversion policy before the same-unit return;
that return does not quantize or reapply custom amount limits. RoundContext and
TimesContext manipulate amounts without checking unit admission. Rounding of a
conversion quotient occurs before subtracting the target affine offset, so the
requested quotient scale need not equal final amount scale. Composed logistics
operations reuse that policy at each boundary. These facts are documented,
not redesigned in this prose pass.

JSON/XML/SQL and profile tests establish bounded parsing, sentinel classification,
receiver preservation and copies. Context tests establish admission order and
cancellation identity; race tests exercise reuse without making an unsupported
primitive-preemption promise. Carrier fixtures test raw formulas, not current
commercial tariff policy. No new performance result is asserted.

No unresolved rationale requires a guessed guarantee. Whether to alter same-unit
limit checking or introduce a single-final-rounding formula API is future runtime
policy outside this task; current behavior is the documented contract.

## Completion evidence

All initial documentation gaps above were addressed or clarified from source;
existing good API comments, examples and regression assertions were preserved.
The audit includes all 31 files, 3 fuzz targets, 2 benchmarks, 4 examples,
context helpers, unit authority and carrier fixtures. No generated/platform
source was excluded. Ordinary complete-final-diff inspection found no remaining
documentation finding; the documented runtime-policy boundaries remain unchanged.

Selected local evidence (Go 1.27.2, modules.json/.go-version):

- `gofmt -l` on all changed Go files: no output; `git diff --check`: passed.
- Comment-aware lexical comparison: executable content/string literals unchanged
  in all 12 changed Go files. Go parser/scanner comparison additionally parsed
  every tracked Go file and found identical non-comment executable tokens.
- Objective AST comment scan: 164 exported declarations/members have valid
  name-prefixed documentation; one canonical comment per public package path.
- `make -f verification/package.mk architecture security docs`: passed. The
  documentation target executes only `go test ./... -run '^Example'`: four root
  examples pass; the two adapter packages compile but have no examples selected.
- `go doc .`, `go doc ./adapters/wire`, `go doc ./measurementwire`,
  `go doc Quantity.ConvertContext` and `go doc Quantity.FormatContext`: passed;
  rendered contracts were inspected for useful entry points and caveats.
- Active `api/v3.txt` regenerated with the native configured recipe:
  concatenate `go doc -all` output for `.`, `./adapters/wire`, `./measurementwire`
  in that order, trim final newline bytes, append one newline. A second
  generation is byte-identical. Historical v1/v2 snapshots are unchanged.
- Changed local links and unfinished-marker checks: passed. No new build or
  generator markers; existing facade deprecation remains valid. Changed-content
  private-key/cloud/access-token/chat-token indicators: absent.
- All Go runs used private build/module/temporary caches in one bounded session;
  those caches and the temporary AST/token checker were removed after capture.

No broad runtime tests, race/fuzz/mutation/benchmark campaign, online source
monitoring, public consumer or release qualification was selected for this Tier A
batch. This source/documentation evidence does not certify remote main or hosted
CI; the coordinator owns delivery and shared persistent tooling enforcement.

## Objective enforcement adoption

The local `source-docs-check` target is required by ordinary/full checks and CI.
Shared CI uses `source_documentation: check` with the immutable Tools v2.4.0
workflow and matching authenticated binary/checksum declaration. Structural
checks do not replace this maintained semantic audit or native docs/examples.
