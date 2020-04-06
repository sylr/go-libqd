DEBUG                ?= 0
VERBOSE              ?= 0

ifneq ($(DEBUG),0)
GO_TEST_FLAGS        += -count=1
endif
ifneq ($(VERBOSE),0)
GO_TEST_FLAGS        += -v
GO_TEST_BENCH_FLAGS  += -v
endif

# -- test ----------------------------------------------------------------------

.PHONY: test bench

test:
	go test $(GO_TEST_FLAGS) ./...

bench:
	@go test $(GO_TEST_FLAGS) -bench=.* ./...

# -- go mod --------------------------------------------------------------------

.PHONY: go-mod-download go-mod-download go-mod-verify

go-mod-download:
	@go mod download

go-mod-tidy:
	@go mod tidy

go-mod-verify: go-mod-download
	@git diff --quiet go.* || git diff --exit-code go.*
