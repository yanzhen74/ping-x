.PHONY: build build-linux build-windows release clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -s -w"
DIST_DIR = dist

build:
	go build $(LDFLAGS) -o ping-x .

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_linux_amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/ping-x_$(VERSION)_windows_amd64.exe .

release: clean build-linux build-windows
	cd $(DIST_DIR) && tar czf ping-x_$(VERSION)_linux_amd64.tar.gz ping-x_$(VERSION)_linux_amd64
	cd $(DIST_DIR) && zip ping-x_$(VERSION)_windows_amd64.zip ping-x_$(VERSION)_windows_amd64.exe

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
