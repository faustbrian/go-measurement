# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's contract and docs.
2. Run ordinary Go tests for the affected package before editing.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable entries in the
[specification decision register](docs/specification-decisions.md), and complete
the Specification Decisions section of the pull request template. An unresolved
interpretation or stale source pin is release-blocking; peer behavior cannot
silently select policy. When no maintained software peer is practical, record
the differential lane as not assessed instead of implying agreement.

Selected mutation gates must finish with zero surviving viable mutants.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use Go 1.27.2 for development and ordinary CI; the public minimum remains
Go 1.27.0. An ordinary standalone checkout is sufficient for tests and a
focused active fuzz smoke, without sibling repositories or goal ledgers:

```bash
go test ./...
go test . -run '^$' -fuzz '^FuzzParseAndTextRoundTrip$' -fuzztime=2s
```

Select the native fuzz target affected by your change. Ordinary tests run
deterministic seeds and saved corpora. For shared checks, install the public
`golib` v2.4.0 release declared in `.golib.yaml` and verify its checksums:

```bash
make check
make ci
```

Both Make routes use bounded local assurance, including one race run of the
named context-isolation assertion. Base-free Make checks run deterministic
seeds/corpora; changed parser inputs receive causal active smoke in CI.
`make ci` also validates the specification register offline. The sole root
module declares no services or test tags; its wire adapters are packages in
this module, not sibling or nested modules.

Optional mutation analysis can select the root package:

```bash
golib mutation --module . --package .
```

For exact-unit, ratio or logistics changes, select the maintained canonical
unit and pinned carrier fixtures rather than inferring policy from a peer:

```bash
go test . -run '^Test(EveryFiniteUnitRatioAgainstCanonicalUnit|OfficialExactUnitDefinitions|DBSchenkerEuroPalletLoadingMetreFixture|CarrierVolumetricWeightFixtures|DSVEuroPalletVolumetricIndexFixture)$' -count=1
```

Use `make check-full` for explicit online source validation and full assurance
with all unchanged 10,000-iteration fuzz campaigns, architecture/security,
API baseline, canonical units, exact ratios, carrier fixtures, docs and
benchmarks. Preserve all pinned fixture/source authority. Select directly
affected public consumers and relevant release qualification, including
`golib release check` and `golib release dry-run --all`, when their contract
changes. Full campaigns remain explicit manual or release assurance;
ordinary scheduled CI stays bounded and separately monitors sources online.
Report unavailable or failing checks without claiming release
readiness. CI/configuration-only changes do not require a library release.

## Selected Assurance

No periodic fuzz or mutation exploration is configured. Deterministic parser
seeds and changed-causal PR smoke cover the affected text/JSON/XML-SQL inputs;
focused manual exploration remains available through the commands above.
Canonical-unit, exact-ratio and carrier fixtures address the numerical contract,
while the named context-isolation race addresses concurrent calls. A rotating
parser campaign would not strengthen those authority or concurrency proofs.
Select relevant fixtures, consumers and explicit full/release checks for their
material risk rather than repeating every campaign after unchanged inputs.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.

## Source documentation

Use the authenticated Tools v2.4.0 release declared in `.golib.yaml`.
`make source-docs-check` checks all tracked Go documentation without compilation.
It is required by `make check`, `make check-full` and `make ci`; shared CI selects
the same strict check and cannot pass its required result if the check fails.
Semantic contract review and the native documentation/examples remain necessary.
