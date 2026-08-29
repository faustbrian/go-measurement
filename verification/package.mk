.PHONY: architecture docs security

architecture:
	./scripts/check-architecture.sh

docs:
	./scripts/check-docs.sh

security:
	./scripts/check-security.sh
