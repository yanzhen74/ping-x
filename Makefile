.PHONY: build build-linux build-windows release clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -s -w"
GOFLAGS = -mod=vendor
DIST_DIR = dist

build:
	go build $(GOFLAGS) $(LDFLAGS) -o ping-x .

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_linux_amd64 .

build-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_windows_amd64.exe .

release: clean build-linux build-windows
	cd $(DIST_DIR) && tar czf ping-x_$(VERSION)_linux_amd64.tar.gz ping-x_$(VERSION)_linux_amd64
	cd $(DIST_DIR) && zip ping-x_$(VERSION)_windows_amd64.zip ping-x_$(VERSION)_windows_amd64.exe

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
