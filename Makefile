
# pre-commit
lint:
	golangci-lint run --verbose --max-issues-per-linter=0 --max-same-issues=0

lint-fix:
	golangci-lint run --verbose --fix
