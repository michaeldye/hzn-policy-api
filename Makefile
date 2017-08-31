SHELL := /bin/bash
# N.B. this is for compat only, we want to use uname -m instead
ARCH = $(shell tools/arch-tag)
VERSION = $(shell cat VERSION)

EXECUTABLE = $(shell basename $$PWD)
PKGS=$(shell go list ./... | gawk '$$1 !~ /vendor\// {print $$1}')

DOCKER_IMAGE = "summit.hovitos.engineering/$(ARCH)/$(EXECUTABLE)"
DOCKER_TAG = "$(VERSION)"

COMPILE_ARGS := CGO_ENABLED=0
# TODO: handle other ARM architectures on build boxes too
ifeq ($(ARCH),armv7l)
	COMPILE_ARGS +=  GOARCH=arm GOARM=7
endif

all: $(EXECUTABLE)

# will always run b/c deps target is PHONY
$(EXECUTABLE): $(shell find . -name '*.go' -not -path './vendor/*')
	$(COMPILE_ARGS) go build -o $(EXECUTABLE)

clean:
	find ./vendor -maxdepth 1 -not -path ./vendor -and -not -iname "vendor.json" -print0 | xargs -0 rm -Rf
	rm -f $(EXECUTABLE)

deps: $(GOPATH)/bin/govendor
	govendor sync

# bail if there are uncommitted changes (note: this doesn't know about or check untracked, uncommitted files)
dirty:
	@echo "Checking if your local repository or index have uncommitted changes..."
	git diff-index --quiet HEAD

$(GOPATH)/bin/govendor:
	go get -u github.com/kardianos/govendor

lint:
	-golint ./... | grep -v "vendor/"
	-go vet ./... 2>&1 | grep -vP "exit\ status|vendor/"

# only unit tests
test: all
	go test -v -cover $(PKGS)

test-integration: all
	go test -v -cover -tags=integration $(PKGS)

.PHONY: clean deps lint publish test test-integration
