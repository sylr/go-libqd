DEBUG                ?= 0
VERBOSE              ?= 0

ifneq ($(DEBUG),0)
GO_TEST_FLAGS        += -count=1
endif
ifneq ($(VERBOSE),0)
GO_TEST_FLAGS        += -v
GO_TEST_BENCH_FLAGS  += -v
endif

GO_TOOLS_GOLANGCI_LINT ?= $(shell go env GOPATH)/bin/golangci-lint

# -- test ----------------------------------------------------------------------

.PHONY: test bench
.ONESHELL: test bench lint

test:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR)/$$dir; \
		go test $(GO_TEST_FLAGS) ./...; \
	done

bench:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR)/$$dir; \
		go test $(GO_TEST_FLAGS) -bench=.* ./...; \
	done

lint: $(GO_TOOLS_GOLANGCI_LINT)
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		cd $(CURDIR)/$$dir; \
		$(GO_TOOLS_GOLANGCI_LINT) run; \
	done

# -- tools ---------------------------------------------------------------------

.PHONY: tools

tools: $(GO_TOOLS_GOLANGCI_LINT)

$(GO_TOOLS_GOLANGCI_LINT):
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0


# -- go mod --------------------------------------------------------------------

.PHONY: go-mod-verify go-mod-tidy
.ONESHELL: go-mod-verify go-mod-tidy

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

# -- release -------------------------------------------------------------------

.PHONY: tags
.ONESHELL: tags

tags:
	@for dir in $$(find . -name go.mod ! -path \*/example/\* -exec dirname {} \;); do \
		PRETTY_DIR=$$(sed -e "s#^./##" -e "s#/\$$##" <<<"$$dir"); \
		LAST_TAG=$$(cd $$dir; git describe --abbrev=0 --match="$$PRETTY_DIR/*" | sed -e "s#^$$PRETTY_DIR/##"); \
		NEXT_TAG=$$(cd $$dir; gorelease -base=$$LAST_TAG | grep 'Suggested version' | sed -e 's#Suggested version: .* (with tag \(.*\))#\1#'); \
		git rev-parse $$NEXT_TAG -- > /dev/null 2>&1; \
		if [ "$$?" -gt "0" ]; then \
			if [ -n "$$TAG" ]; then \
				echo "Tagging $$NEXT_TAG"; \
				git tag -s -m $$NEXT_TAG $$NEXT_TAG; \
			else \
				echo "Would tag $$NEXT_TAG"; \
			fi; \
		else \
			echo "No new tag for $$PRETTY_DIR"; \
		fi; \
	done
