# -- test ----------------------------------------------------------------------

.PHONY: test bench

test:
	@go test ./...

bench:
	@go test -bench=.* ./...

# -- go mod --------------------------------------------------------------------

.PHONY: go-mod-download go-mod-download go-mod-verify

go-mod-download:
	@go mod download

go-mod-tidy:
	@go mod tidy

go-mod-verify: go-mod-download
	@git diff --quiet go.* || git diff --exit-code go.*
