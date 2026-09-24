# Security Policy

Report vulnerabilities through the repository's
[private GitHub advisory form](https://github.com/faustbrian/go-measurement/security/advisories/new).
Do not put exploit details, live payloads, or reporter data in public issues.

The published v1 major is supported; its latest release is v1.1.0. The v2
source on `main` is planned, non-releasable, and has no supported published
version. Assess each report against its affected module path and released
versions. A confirmed vulnerability requires a focused regression, an affected-
version assessment, a private fix, an advisory, release notes, and upgrade
guidance where a released version is affected. A fix for v1 need not publish
unrelated v2 source.

Maintainers will privately acknowledge and triage reports, assign severity by
exploitability and impact, and coordinate remediation, any embargo, and public
disclosure with the reporter. No fixed response or release deadline is promised.

The threat model, bounds, and deployment guidance are in
[docs/security.md](docs/security.md).
