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
.ONESHELL: test bench

test:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR); \
		cd $$dir; \
		go test $(GO_TEST_FLAGS) ./...; \
	done

bench:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR); \
		cd $$dir; \
		go test $(GO_TEST_FLAGS) -bench=.* ./...; \
	done

# -- go mod --------------------------------------------------------------------

.PHONY: go-mod-verify go-mod-tidy
.ONESHELL: go-mod-verify go-mod-tidy
.SHELLFLAGS: -e

go-mod-verify:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR); \
		cd $$dir; \
		go mod download; \
		git diff --quiet go.* || git diff --exit-code go.* || exit 1; \
	done

go-mod-tidy:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR); \
		cd $$dir; \
		go mod download; \
		go mod tidy; \
	done
