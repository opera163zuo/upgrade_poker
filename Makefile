.PHONY: build build-all build-win build-linux build-mac run run-gui test stress clean release

BINARY=upgrade
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || date +%Y%m%d)
RELEASE_DIR=release

build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BINARY) .

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/upgrade_$(VERSION)_win64.exe .

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/upgrade_$(VERSION)_linux64 .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/upgrade_$(VERSION)_mac64 .

build-all: build-win build-linux build-mac
	@echo "=== 三平台全部编译完成 ==="
	@ls -lh $(RELEASE_DIR)/

run: build
	./$(BINARY)

run-gui: build
	./$(BINARY) -ui=gui

test:
	go test -v -timeout 30s ./...

stress:
	go test -v -run TestAIPlayNoCrash -count=50 -timeout 120s ./...

clean:
	rm -f $(BINARY)
	rm -rf $(RELEASE_DIR)/*

release: build-all
	@echo "=== 发布包已就绪 ==="
