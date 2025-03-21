NAME=eat
BIN_DIR=$(shell mkdir -p dist; echo 'dist')
BRANCH=$(shell git branch --show-current)
ifeq ($(BRANCH),master) # master production branch
	BUILD_TAG=master-$(shell git rev-parse HEAD)
else ifeq ($(BRANCH),) # checkout to specific tag
	BUILD_TAG=$(shell git describe --tags)
else # other branch
	BUILD_TAG=$(shell git rev-parse HEAD)
endif
CURRENT_OS=$(shell go env GOOS)
CURRENT_ARCH=$(shell go env GOARCH)
CURRENT_TARGET=$(CURRENT_OS)-$(CURRENT_ARCH)

BUILD_TIME=$(shell date -Iseconds --utc)
GO_BUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-X "eat/cmd/version.BuildHash=$(BUILD_TAG)" \
													-X "eat/cmd/version.BuildTime=$(BUILD_TIME)" \
													-w -s -buildid='

.PHONY: all clean default darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 freebsd-amd64 freebsd-arm64 windows-amd64 windows-arm64

default:
	@echo "Building for current system: $(CURRENT_TARGET)"
	@make $(CURRENT_TARGET)

all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

darwin-amd64:
	GOARCH=amd64 GOOS=darwin GOAMD64=v3 $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux GOAMD64=v3 $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd GOAMD64=v3 $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows GOAMD64=v3 $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@.exe

windows-arm64:
	GOARCH=arm64 GOOS=windows $(GO_BUILD) -o $(BIN_DIR)/$(NAME)-$@.exe

clean:
	rm -f $(BIN_DIR)/*
