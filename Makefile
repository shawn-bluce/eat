NAME      := eat
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -buildid= \
	-X eat/cmd/version.Version=$(patsubst v%,%,$(VERSION)) \
	-X eat/cmd/version.BuildHash=$(shell git rev-parse HEAD 2>/dev/null) \
	-X eat/cmd/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_BUILD := CGO_ENABLED=0 go build -trimpath -ldflags '$(LDFLAGS)'

.PHONY: build all test clean $(PLATFORMS)

# Build for the current system: dist/eat
build:
	$(GO_BUILD) -o dist/$(NAME) .

# Build for every supported platform: dist/eat-<os>-<arch>[.exe]
all: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) \
		$(GO_BUILD) -o dist/$(NAME)-$@$(if $(findstring windows,$@),.exe) .

test:
	go vet ./...
	go test ./...

clean:
	rm -rf dist
