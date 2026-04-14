.PHONY: build build-linux build-windows release clean vendor-build vendor-build-linux vendor-build-windows

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -s -w"
DIST_DIR = dist

# 默认使用 go mod 方式
GOFLAGS =

# 标准构建（go mod 方式）
build:
	go build $(GOFLAGS) $(LDFLAGS) -o ping-x .

build-linux:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_linux_amd64 .

build-windows:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_windows_amd64.exe .

release: clean build-linux build-windows
	cd $(DIST_DIR) && tar czf ping-x_$(VERSION)_linux_amd64.tar.gz ping-x_$(VERSION)_linux_amd64
	cd $(DIST_DIR) && zip ping-x_$(VERSION)_windows_amd64.zip ping-x_$(VERSION)_windows_amd64.exe

# ============================================
# 内网离线构建（vendor 方式）
# 使用方法：make vendor-build
# ============================================
VENDOR_FLAGS = -mod=vendor

vendor-build:
	go build $(VENDOR_FLAGS) $(LDFLAGS) -o ping-x .

vendor-build-linux:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(VENDOR_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_linux_amd64 .

vendor-build-windows:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build $(VENDOR_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_windows_amd64.exe .

vendor-release: clean vendor-build-linux vendor-build-windows
	cd $(DIST_DIR) && tar czf ping-x_$(VERSION)_linux_amd64.tar.gz ping-x_$(VERSION)_linux_amd64
	cd $(DIST_DIR) && zip ping-x_$(VERSION)_windows_amd64.zip ping-x_$(VERSION)_windows_amd64.exe

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
