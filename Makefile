.PHONY: build build-win package-win run run-gui test stress clean release

BINARY=upgrade
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || date +%Y%m%d)
RELEASE_DIR=release

build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BINARY) .

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/upgrade_$(VERSION)_win64.exe .

package-win: build-win
	powershell -NoProfile -Command "Compress-Archive -LiteralPath '$(RELEASE_DIR)/upgrade_$(VERSION)_win64.exe' -DestinationPath '$(RELEASE_DIR)/upgrade_$(VERSION)_win64.zip' -Force"
	@echo "=== Windows 发布产物已生成 ==="
	@ls -lh $(RELEASE_DIR)/upgrade_$(VERSION)_win64.*

run: build
	./$(BINARY)

run-gui: build
	./$(BINARY)

test:
	go test -v -timeout 30s ./...

stress:
	go test -v -run TestAIPlayNoCrash -count=50 -timeout 120s ./...

clean:
	rm -f $(BINARY)
	rm -rf $(RELEASE_DIR)/*

release: package-win
	@echo "=== Windows 发布包已就绪 ==="
