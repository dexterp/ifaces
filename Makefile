SHELL = /bin/bash

MAKEFLAGS += --no-print-directory

# Project
NAME := $(shell cat NAME)
GOPATH := $(shell go env GOPATH)
VERSION ?= $(shell cat VERSION)
HASCMD := $(shell test -d cmd && echo "true")
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ARCH = $(shell uname -m)

ISRELEASED := $(shell git show-ref v$$(cat VERSION) 2>&1 > /dev/null && echo "true")

# Utilities
# Default environment variables.
# Any variables already set will override the values in this file(s).
DOTENV := godotenv -f $(HOME)/.env,.env

# Variables
ROOT = $(shell pwd)

# Go
GOMODOPTS = GO111MODULE=on
GOGETOPTS = GO111MODULE=off
GOFILES := $(shell find cmd pkg internal src -name '*.go' 2> /dev/null)
GODIRS = $(shell find . -maxdepth 1 -mindepth 1 -type d | egrep 'cmd|internal|pkg|api')

#
# End user targets
#

### HELP

.PHONY: help
help: ## Print Help
	@maker --menu Makefile

### DEVELOPMENT

.PHONY: _build
_build: ## Build binary
	@test -d .cache || go fmt ./...
ifeq ($(HASCMD),true)
ifeq ($(XCOMPILE),true)
	GOOS=linux GOARCH=amd64 $(MAKE) dist/$(NAME)_linux_amd64/$(NAME)
	GOOS=darwin GOARCH=amd64 $(MAKE) dist/$(NAME)_darwin_amd64/$(NAME)
	GOOS=windows GOARCH=amd64 $(MAKE) dist/$(NAME)_windows_amd64/$(NAME).exe
endif
	@$(MAKE) $(NAME)
endif

.PHONY: _install
_install: $(GOPATH)/bin/$(NAME) ## Install to $(GOPATH)/bin

.PHONY: clean
clean: ## Reset project to original state
	rm -rf .cache $(NAME) dist reports tmp vendor

.PHONY: test
test: ## Test
	$(MAKE) lint
	$(MAKE) _test
	@# Combined the return codes of all the tests
	@echo "Exit codes, unit tests: $$(cat reports/exitcode-unit.txt), golangci-lint: $$(cat reports/exitcode-golangci-lint.txt), staticcheck: $$(cat reports/exitcode-staticcheck.txt), vet: $$(cat reports/exitcode-vet.txt)"
	@exit $$(( $$(cat reports/exitcode-unit.txt) + $$(cat reports/exitcode-golangci-lint.txt) + $$(cat reports/exitcode-staticcheck.txt) + $$(cat reports/exitcode-vet.txt) ))

# Test without feature flags
.PHONY: _test
_test:
	$(MAKE) unit
	$(MAKE) cx
	$(MAKE) cc

.PHONY: _unit
_unit:
	### Unit Tests
	gotestsum --jsonfile reports/unit.json --junitfile reports/junit.xml -- -timeout 60s -covermode atomic -coverprofile=./reports/coverage.out -v ./...; echo $$? > reports/exitcode-unit.txt
	@go-test-report -t "$(NAME) unit tests" -o reports/html/unit.html < reports/unit.json > /dev/null

.PHONY: _cc
_cc:
	### Code Coverage
	@go-acc -o ./reports/coverage.out ./... > /dev/null
	@go tool cover -func=./reports/coverage.out | tee reports/coverage.txt
	@go tool cover -html=reports/coverage.out -o reports/html/coverage.html

.PHONY: _cx
_cx:
	### Cyclomatix Complexity Report
	@gocyclo -avg $(GODIRS) | grep -v _test.go | tee reports/cyclomaticcomplexity.txt
	@contents=$$(cat reports/cyclomaticcomplexity.txt); echo "<html><title>cyclomatic complexity</title><body><pre>$${contents}</pre></body><html>" > reports/html/cyclomaticcomplexity.html

.PHONY: _release
_release: ## Trigger a release by creating a tag and pushing to the upstream repository
	@echo "### Releasing v$(VERSION)"
	@$(MAKE) _isreleased 2> /dev/null
	git fetch --tags
	git tag v$(VERSION)
	git push --tags

.PHONY: lint
lint: internal/resources/version/version.go
	golangci-lint run --enable=gocyclo; echo $$? > reports/exitcode-golangci-lint.txt
	go vet ./...; echo $$? > reports/exitcode-vet.txt
	staticcheck ./...; echo $$? > reports/exitcode-staticcheck.txt

.PHONY: gomod
gomod: go.mod ## Install go.mod modules
	$(GOMODOPTS) go mod tidy
	$(GOMODOPTS) go mod download

.PHONY: goinstall
goinstall: ## Install required command line tools by running go install ...
	cat .installs.txt | egrep -v '^#' | xargs -I{} -t -n1 go install {}
	@$(MAKE) tmp/install

.PHONY: report
report: ## Open reports generated by "make test" in a browser
	@$(MAKE) $(REPORTS)

### VERSION INCREMENT

.PHONY: bumpmajor
bumpmajor: ## Increment VERSION file ${major}.0.0 - major bump
	git fetch --tags
	versionbump --checktags major VERSION

.PHONY: bumpminor
bumpminor: ## Increment VERSION file 0.${minor}.0 - minor bump
	git fetch --tags
	versionbump --checktags minor VERSION

.PHONY: bumppatch
bumppatch: ## Increment VERSION file 0.0.${patch} - patch bump
	git fetch --tags
	versionbump --checktags patch VERSION

.PHONY: getversion
getversion:
	VERSION=$(VERSION) bash -c 'echo $$VERSION'

#
# Helper targets
#

# Open html reports
REPORTS = reports/html/unit.html reports/html/coverage.html reports/html/cyclomaticcomplexity.html
.PHONY: $(REPORTS)
$(REPORTS):
ifeq ($(GOOS),darwin)
	@test -f $@ && open $@
else ifeq ($(GOOS),linux)
	@test -f $@ && xdg-open $@
endif

# Check versionbump
.PHONY: _isreleased
_isreleased:
ifeq ($(ISRELEASED),true)
	@echo "Version $(VERSION) has been released."
	@echo "Please bump with 'make bump(minor|patch|major)' depending on breaking changes."
	@exit 1
endif

#
# File targets
#
$(GOPATH)/bin/$(NAME): $(NAME)
	install -m 755 $(NAME) $(GOPATH)/bin/$(NAME)

$(NAME): dist/$(NAME)_$(GOOS)_$(GOARCH)/$(NAME)
	install -m 755 $< $@

dist/$(NAME)_$(GOOS)_$(GOARCH)/$(NAME) dist/$(NAME)_$(GOOS)_$(GOARCH)/$(NAME).exe: $(GOFILES) internal/resources/version/version.go
	@mkdir -p $$(dirname $@)
	go build -o $@ ./cmd/$(NAME)

internal/resources/version/version.go: internal/resources/version/version.go.in VERSION
	@VERSION=$(VERSION) $(DOTENV) envsubst < $< > $@

tmp/install: .installs.txt
	cat .installs.txt | egrep -v '^#' | xargs -I{} -t -n1 go install {}
	@mkdir -p tmp
	@touch tmp/installs

#
# make wrapper - Execute any target target prefixed with a underscore.
# EG 'make task' will result in the execution of 'make _task' 
#
%:
	@maker $@
