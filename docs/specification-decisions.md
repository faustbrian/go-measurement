# Specification decisions

This register records the source-backed choices that define observable measurement behavior. The machine-readable authority is [`specification/decisions.json`](../specification/decisions.json); evidence bindings are in [`specification/conformance.json`](../specification/conformance.json). An empty differential lane means no maintained software peer was assessed, not that agreement was proved.

## MEASUREMENT-DEC-001: Exact unit catalog and accepted SI units

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `interoperability policy`; `application-policy`.
- **Authority:** BIPM SI Brochure 9th edition version 4.01; BIPM SI Brochure 9th edition version 4.01 (2026); `bipm-si-brochure-source`; [source](https://www.bipm.org/documents/20126/41483022/SI-Brochure-9.pdf), section 2.3, 2.4, 2.5, 4, and 5; requirement strength `not specified`.
- **Issue:** The SI Brochure defines SI units, prefixes, coherent derived units, and accepted units, while the package must also choose a finite catalog and exact decimal representation for international customary units.
- **Interpretations:** `Expose only coherent SI base units`, `Expose every named unit found in source tables`, `Expose a bounded logistics-oriented catalog with exact source-defined ratios`
- **Peer behavior:** Maintained unit libraries expose different catalogs, aliases, numeric representations, and exactness guarantees; no peer comparison is executed by this repository.
- **Selected behavior:** Use metre, square metre, cubic metre, kilogram, kelvin, kilogram per cubic metre, and one as canonical units; encode decimal SI prefixes exactly; map litre to 0.001 cubic metre and tonne to 1000 kilograms; and encode the international inch, foot, yard, ounce, and pound with exact NIST factors.
- **Rationale:** A closed audited catalog keeps every conversion ratio finite, attributable, and stable without binary floating-point conversion.
- **Security consequences:** Unknown units fail closed and cannot introduce caller-controlled conversion ratios.
- **Resource consequences:** The fixed catalog bounds lookup and conversion-factor size.
- **Compatibility consequences:** Adding or changing a unit, symbol, dimension, or ratio is a reviewed compatibility event.
- **Wire consequences:** Serialized quantities retain the stable catalog symbol and exact decimal amount.
- **Executable evidence:** `TestEveryFiniteUnitRatioAgainstCanonicalUnit`, `TestOfficialExactUnitDefinitions`, `TestEveryUnitHasTheAuditedSymbolAndDimension`
- **Fixture evidence:** `unit_definitions_test.go`
- **Fuzz evidence:** `FuzzParseAndTextRoundTrip`
- **Interoperability evidence:** None.
- **Differential evidence:** None.
- **Public APIs:** `Unit`, `Units`, `Quantity.Convert`, `SymbolProfile`
- **Documentation:** `docs/specification-decisions.md`, `docs/sources.md`, `docs/exactness.md`
- **Upstream status:** The BIPM source is current at version 4.01; NIST SP 811 remains the pinned authority for the exact international customary conversion factors used here.
- **Reconsider when:** The BIPM SI Brochure, NIST conversion factors, or the supported unit catalog changes.
- Additional authoritative source: `{"id":"nist-sp811-source","version":"NIST SP 811 2008 edition","url":"https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication811e2008.pdf","specifications":["NIST SP 811 2008 edition"]}`

## MEASUREMENT-DEC-002: Exact affine temperature conversion

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `interoperability policy`; `application-policy`.
- **Authority:** NIST SP 811 2008 edition; NIST SP 811 2008 edition; `nist-sp811-source`; [source](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication811e2008.pdf), section Appendix B.8 temperature conversion factors; requirement strength `not specified`.
- **Issue:** Celsius and Fahrenheit are affine scales, so a ratio-only conversion model would produce incorrect absolute temperatures and intermediate rounding would lose exact source semantics.
- **Interpretations:** `Treat every temperature unit as a ratio scale`, `Store pre-offset affine definitions and apply one final conversion policy`, `Convert through binary floating point`
- **Peer behavior:** Unit libraries vary between separate absolute and interval types, affine conversion callbacks, and floating-point formulas; no peer comparison is executed by this repository.
- **Selected behavior:** Use K = C + 273.15 and K = (F + 459.67) * 5 / 9 with exact decimal operands, then apply the caller-selected exact or rounded conversion context only to the combined quotient.
- **Rationale:** The affine definitions preserve the official scale relationships and make unavoidable non-terminating decimal results explicit.
- **Security consequences:** Temperature input cannot bypass dimension checks or select an implicit rounding policy.
- **Resource consequences:** Affine arithmetic uses the same bounded decimal limits as every other conversion.
- **Compatibility consequences:** Changing offsets, ratios, or rounding placement changes observable values and requires a new reviewed decision digest.
- **Wire consequences:** Wire values preserve the selected temperature symbol and decimal result; conversion is never implicit during serialization.
- **Executable evidence:** `TestFahrenheitUsesExactDefinedRatioBeforeRounding`, `TestOfficialExactUnitDefinitions`, `TestTemperatureConversionRequiresExplicitRounding`
- **Fixture evidence:** `unit_definitions_test.go`
- **Fuzz evidence:** None.
- **Interoperability evidence:** None.
- **Differential evidence:** None.
- **Public APIs:** `Celsius`, `Fahrenheit`, `Kelvin`, `Quantity.Convert`
- **Documentation:** `docs/specification-decisions.md`, `docs/sources.md`, `docs/exactness.md`
- **Upstream status:** NIST publishes the exact Celsius, kelvin, and Fahrenheit relationships used by the package.
- **Reconsider when:** The authoritative temperature guidance changes or temperature intervals become a supported model.
- Additional authoritative source: `{"id":"bipm-si-brochure-source","version":"BIPM SI Brochure 9th edition version 4.01 (2026)","url":"https://www.bipm.org/documents/20126/41483022/SI-Brochure-9.pdf","specifications":["BIPM SI Brochure 9th edition version 4.01"]}`

## MEASUREMENT-DEC-003: Explicit exact or single-rounding conversion

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `implementation-defined behavior`; `application-policy`.
- **Authority:** NIST SP 811 2008 edition; NIST SP 811 2008 edition; `nist-sp811-source`; [source](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication811e2008.pdf), section Appendix B conversion-factor exactness and section 7.9 rounding; requirement strength `not specified`.
- **Issue:** Source factors distinguish exact values but do not choose a decimal library policy for non-terminating quotients, rounding scale, rounding mode, or intermediate operations.
- **Interpretations:** `Round every intermediate operation`, `Silently apply a package default`, `Require exact division or one explicit final rounding operation`
- **Peer behavior:** Numeric and unit libraries choose different default precision and rounding behavior; no peer comparison is executed by this repository.
- **Selected behavior:** ExactConversion rejects a non-terminating base-10 quotient; RoundedConversion combines the full ratio and rounds once at the requested scale and mode; the zero ConversionContext is invalid.
- **Rationale:** An explicit context prevents hidden precision loss and separates source exactness from caller-owned output policy.
- **Security consequences:** Callers cannot trigger ambient precision or rounding defaults through a zero-value context.
- **Resource consequences:** All exact and rounded arithmetic remains subject to caller-selected or default decimal resource limits.
- **Compatibility consequences:** Existing exact failures and once-rounded results are stable public behavior.
- **Wire consequences:** Formatting and serialization do not select conversion precision on behalf of callers.
- **Executable evidence:** `TestRoundedConversionRoundsOnlyTheFinalRatio`, `TestTemperatureConversionRequiresExplicitRounding`, `TestZeroConversionContextDoesNotInferRoundingPolicy`
- **Fixture evidence:** None.
- **Fuzz evidence:** None.
- **Interoperability evidence:** None.
- **Differential evidence:** None.
- **Public APIs:** `ExactConversion`, `RoundedConversion`, `ConversionContext.WithLimits`, `Quantity.Convert`, `Quantity.Format`
- **Documentation:** `docs/specification-decisions.md`, `docs/exactness.md`
- **Upstream status:** NIST identifies exact conversion factors and rounding guidance, while decimal execution policy remains package-owned.
- **Reconsider when:** The decimal dependency changes its exact-division contract or the public conversion context changes.

## MEASUREMENT-DEC-004: Closed dimensions and canonical derived results

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `implementation-defined behavior`; `application-policy`.
- **Authority:** BIPM SI Brochure 9th edition version 4.01; BIPM SI Brochure 9th edition version 4.01 (2026); `bipm-si-brochure-source`; [source](https://www.bipm.org/documents/20126/41483022/SI-Brochure-9.pdf), section 2.3.4, 2.3.5, and 2.3.6; requirement strength `not specified`.
- **Issue:** Dimensional algebra permits many derived quantities, but the public package must choose which dimensions exist and which unit represents each supported arithmetic result.
- **Interpretations:** `Create arbitrary symbolic dimensions`, `Return the left operand unit`, `Use a closed dimension set and one canonical result unit per supported operation`
- **Peer behavior:** Unit libraries vary between symbolic dimension algebra and closed domain-specific catalogs; no peer comparison is executed by this repository.
- **Selected behavior:** Support dimensionless, length, area, volume, mass, absolute temperature, density, and loading metre; return square metre, cubic metre, kilogram per cubic metre, kilogram, or one for supported derived operations; reject every unmodeled result.
- **Rationale:** A closed model prevents accidental ontology growth and makes every derived result predictable.
- **Security consequences:** Unsupported or mismatched dimensions fail before decimal arithmetic.
- **Resource consequences:** Dimension resolution is constant-size and cannot construct unbounded symbolic expressions.
- **Compatibility consequences:** New dimensions or changed canonical result units require explicit compatibility review.
- **Wire consequences:** Derived quantities serialize with one stable canonical unit symbol.
- **Executable evidence:** `TestDerivedDimensionsSelectCanonicalUnits`, `TestDerivedDimensionIdentitiesAndCommutativity`, `TestDivideMassByVolumeAndMultiplyDensityByVolume`
- **Fixture evidence:** None.
- **Fuzz evidence:** None.
- **Interoperability evidence:** None.
- **Differential evidence:** None.
- **Public APIs:** `Dimension`, `Quantity.Multiply`, `Quantity.Divide`, `Quantity.Add`, `Quantity.Subtract`
- **Documentation:** `docs/specification-decisions.md`, `docs/dimensional-analysis.md`
- **Upstream status:** The SI dimensional relationships are stable; the supported subset and canonical result selection are package policy.
- **Reconsider when:** A new public dimension or derived operation is proposed.

## MEASUREMENT-DEC-005: Absolute-temperature arithmetic and symbol profile boundaries

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `optional behavior`; `application-policy`.
- **Authority:** BIPM SI Brochure 9th edition version 4.01; BIPM SI Brochure 9th edition version 4.01 (2026); `bipm-si-brochure-source`; [source](https://www.bipm.org/documents/20126/41483022/SI-Brochure-9.pdf), section 2.3.1, 2.3.2, and 5.4.2; requirement strength `not specified`.
- **Issue:** The source distinguishes thermodynamic temperature from intervals and defines typographic symbols, while this package exposes ASCII wire identifiers and has no temperature-interval type or universal locale alias set.
- **Interpretations:** `Add and subtract absolute temperatures as ordinary scalars`, `Guess aliases and locale preferences globally`, `Reject affine arithmetic and require caller-owned alias profiles`
- **Peer behavior:** Peers differ on temperature-interval types, Unicode symbols, case sensitivity, and locale aliases; no peer comparison is executed by this repository.
- **Selected behavior:** Absolute temperatures can be converted and compared but not added or subtracted; canonical wire symbols are strict ASCII identifiers including m2, m3, degC, and degF; aliases are accepted only through an explicit caller-owned Profile.
- **Rationale:** Separating absolute values from unmodeled intervals prevents invalid arithmetic, while explicit profiles prevent locale and spelling guesses from changing parsing.
- **Security consequences:** Ambiguous aliases and affine arithmetic fail closed rather than being silently reinterpreted.
- **Resource consequences:** Profiles are bounded, validated, and defensively copied.
- **Compatibility consequences:** Canonical symbols remain stable; callers own migrations for their explicit alias profiles.
- **Wire consequences:** JSON, XML, SQL, and text output use canonical ASCII identifiers rather than localized typography.
- **Executable evidence:** `TestAbsoluteTemperaturesRejectAffineArithmetic`, `TestParsingUsesOnlyExplicitProfileAliases`, `TestUnitCatalogIsStableSortedAndAliasSafe`
- **Fixture evidence:** None.
- **Fuzz evidence:** `FuzzParseAndTextRoundTrip`, `FuzzQuantityJSON`
- **Interoperability evidence:** None.
- **Differential evidence:** None.
- **Public APIs:** `Quantity.Add`, `Quantity.Subtract`, `SymbolProfile`, `NewProfile`, `Parse`
- **Documentation:** `docs/specification-decisions.md`, `docs/dimensional-analysis.md`, `docs/serialization.md`
- **Upstream status:** The source distinction remains stable; ASCII symbols and alias ownership are deliberate package compatibility policy.
- **Reconsider when:** A temperature-interval type, localized formatter, or versioned alias profile is introduced.

## MEASUREMENT-DEC-006: Caller-owned logistics formulas and carrier fixtures

- **Status / owner:** `resolved`; go-measurement maintainers.
- **Classification / scope:** `interoperability policy`; `application-policy`.
- **Authority:** Carrier logistics references snapshot 2026-08-30; Carrier logistics references snapshot 2026-08-30; `dsv-logistics-source`; [source](https://www.dsv.com/it-it/sostegno/faq/calcolatore-di-peso-volumetrico), section International road volumetric-weight and loading-metre examples; requirement strength `not specified`.
- **Issue:** Carrier publications provide concrete loading-metre and dimensional-weight examples but do not define universal vehicle width, stacking, divisor, index, dimension-rounding, or chargeable-weight policy.
- **Interpretations:** `Embed one carrier tariff as a universal constant`, `Implement only raw formulas with explicit caller inputs`, `Infer carrier and effective date from dimensions`
- **Peer behavior:** Carrier calculators and unit libraries differ by route, product, divisor, density index, stacking, and tariff rounding; no maintained software peer comparison is executed by this repository.
- **Selected behavior:** Model loading metre as a distinct semantic dimension; calculate floor area divided by explicit truck width and stacking factor; calculate volumetric weight from an explicit volume-per-kilogram divisor or density index; leave orientation, tariff rounding, actual-versus-dimensional selection, carrier, and effective date to callers.
- **Rationale:** Raw typed formulas are reusable across carriers without misrepresenting commercial policy as a physical constant.
- **Security consequences:** Positive dimensions, counts, truck widths, stacking factors, divisors, and indexes are validated before formula execution.
- **Resource consequences:** Package quantity and decimal arithmetic are bounded; formulas perform fixed per-package work.
- **Compatibility consequences:** Carrier fixtures validate formula interoperability but do not promise that a tariff remains current or that the package computes a final billable weight.
- **Wire consequences:** Formula outputs contain raw kilograms or loading metres and no carrier or tariff metadata.
- **Executable evidence:** `TestDBSchenkerEuroPalletLoadingMetreFixture`, `TestCarrierVolumetricWeightFixtures`, `TestDSVEuroPalletVolumetricIndexFixture`, `TestDimensionsCalculateLoadingMetresWithStacking`, `TestVolumetricDivisorUsesExplicitVolumeUnit`
- **Fixture evidence:** `logistics_fixtures_test.go`
- **Fuzz evidence:** None.
- **Interoperability evidence:** `logistics_fixtures_test.go`
- **Differential evidence:** None.
- **Public APIs:** `Dimensions.LoadingMetres`, `NewTruckWidth`, `NewStackingFactor`, `NewVolumetricDivisor`, `NewVolumetricIndex`
- **Documentation:** `docs/specification-decisions.md`, `docs/logistics-formulas.md`, `docs/sources.md`
- **Upstream status:** The carrier sources are attributable snapshots; callers must select current contract values for their route and product.
- **Reconsider when:** A carrier policy is embedded, formula inputs become implicit, or any cited provider changes the checked examples.
- Additional authoritative source: `{"id":"dbschenker-logistics-source","version":"DB Schenker international road terms effective 2026-05-04","url":"https://www.dbschenker.com/resource/blob/2479672/3b7520da0b6a5db3138cbd14b70f3320/terms-of-provision-of-services-in-international-road-forwarding-from-4-05-2026-data.pdf","specifications":["Carrier logistics references snapshot 2026-08-30"]}`
- Additional authoritative source: `{"id":"dhl-logistics-source","version":"DHL DCT help snapshot 2026-08-30","url":"https://dct.dhl.com/help","specifications":["Carrier logistics references snapshot 2026-08-30"]}`
