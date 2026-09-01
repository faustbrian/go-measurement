# Specification conformance matrix

The [specification decision register](../docs/specification-decisions.md) records rationale and consequences. Machine-readable evidence bindings live in `conformance.json`; authority monitoring and source pins live beside it.

| Decision | Observable contract | Evidence disposition |
| --- | --- | --- |
| MEASUREMENT-DEC-001 | Exact finite unit catalog and canonical units | Unit-definition tests and parse fuzzing; software differential not assessed |
| MEASUREMENT-DEC-002 | Exact affine temperature conversion | Temperature definition and rounding tests; software differential not assessed |
| MEASUREMENT-DEC-003 | Explicit exact or once-rounded conversion | Conversion-context regression tests; software differential not assessed |
| MEASUREMENT-DEC-004 | Closed dimensions and canonical derived results | Dimensional identity and arithmetic tests; software differential not assessed |
| MEASUREMENT-DEC-005 | Absolute-temperature and alias boundaries | Arithmetic, profile, catalog, and fuzz evidence; software differential not assessed |
| MEASUREMENT-DEC-006 | Explicit raw logistics formulas | DSV, DB Schenker, and DHL provider fixtures plus a historical FedEx arithmetic vector; software differential not assessed |

The offline specification gate validates every binding and fails for missing, contradictory, stale, unresolved, or untested decisions. The online form revalidates the pinned public authority bodies and change authorities. Carrier fixtures establish agreement only with the cited reviewed examples; they do not prove a current tariff, route, or customer contract.
