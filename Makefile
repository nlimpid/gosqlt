REPO_ROOT := $(abspath .)
COVERDIR ?= coverage
GO_TEST_FLAGS ?=
GO_COVERPKG ?= github.com/nlimpid/gosqlt/scanner

.PHONY: test
test:
	go test ./... $(GO_TEST_FLAGS)
	@if [ -d tests ]; then \
		( cd tests && go test ./... $(GO_TEST_FLAGS) ); \
	fi

.PHONY: test-coverage
test-coverage:
	COVERDIR="$(COVERDIR)" GO_COVERPKG="$(GO_COVERPKG)" ./scripts/test.sh $(GO_TEST_FLAGS)

.PHONY: upload-coverage-report
upload-coverage-report: test-coverage
	COVERDIR="$(COVERDIR)" ./scripts/codecov_upload.sh

.PHONY: clean
clean:
	rm -rf "$(COVERDIR)"
	rm -f bin/codecov.sh
