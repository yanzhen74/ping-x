# ping-x

[![Go Version](https://img.shields.io/badge/Go-1.17+-00ADD8?logo=go)](https://golang.org)
[![Platforms](https://img.shields.io/badge/Platforms-Linux%20%7C%20Windows-blue)]()
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A cross-platform CLI tool for testing network connectivity across TCP, UDP, Multicast (ASM), and SSM protocols.

## Features

- **Multi-Protocol Support**: TCP, UDP, Multicast (ASM), and Source-Specific Multicast (SSM)
- **Dual Mode Operation**: Run as sender or receiver for end-to-end connectivity testing
- **Flexible Configuration**: Command-line flags or YAML configuration files
- **Cross-Platform**: Native builds for Linux and Windows
- **Real-time Statistics**: Packet loss, latency metrics, and throughput

## Quick Start

### Download Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/yanzhen74/ping-x/releases):

```bash
# Linux
curl -L -o ping-x.tar.gz https://github.com/yanzhen74/ping-x/releases/latest/download/ping-x_linux_amd64.tar.gz
tar xzf ping-x.tar.gz

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/yanzhen74/ping-x/releases/latest/download/ping-x_windows_amd64.zip -OutFile ping-x.zip
Expand-Archive ping-x.zip
```

### Build from Source

```bash
git clone https://github.com/yanzhen74/ping-x.git
cd ping-x
go build -o ping-x .
```

## Usage

### TCP Test

**Receiver (Host B):**
```bash
ping-x recv -p tcp --port 9000
```

**Sender (Host A):**
```bash
ping-x send -p tcp -t 192.168.1.100 --port 9000 -n 100
```

### UDP Test

**Receiver (Host B):**
```bash
ping-x recv -p udp --port 9001
```

**Sender (Host A):**
```bash
ping-x send -p udp -t 192.168.1.100 --port 9001 -n 50
```

### Multicast (ASM) Test

**Receiver (Host B):**
```bash
ping-x recv -p multicast -g 239.1.1.1 --port 9002 -i eth0
```

**Sender (Host A):**
```bash
ping-x send -p multicast -g 239.1.1.1 --port 9002 -i eth0 -n 20
```

### SSM (Source-Specific Multicast) Test

**Receiver (Host B):**
```bash
ping-x recv -p ssm -g 232.1.1.1 --source 10.0.0.50 --port 9003 -i eth0
```

**Sender (Host A):**
```bash
ping-x send -p ssm -g 232.1.1.1 --port 9003 -i eth0 -b 10.0.0.50 -n 20
```

### YAML Configuration Mode

Generate a sample configuration file:

```bash
ping-x config -o config.yaml
```

Run with configuration file:

```bash
# Sender mode
ping-x send -c config.yaml

# Receiver mode
ping-x recv -c config.yaml
```

## CLI Reference

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Configuration file path |
| `--verbose` | `-v` | Enable verbose output |

### Send Command

```bash
ping-x send [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--proto` | `-p` | `tcp` | Protocol: tcp, udp, multicast, ssm |
| `--target` | `-t` | - | Target address (required for tcp/udp) |
| `--group` | `-g` | - | Multicast group address (required for multicast/ssm) |
| `--port` | - | - | Target port (required) |
| `--bind` | `-b` | `0.0.0.0` | Local bind address |
| `--iface` | `-i` | - | Network interface name (for multicast) |
| `--count` | `-n` | `0` | Number of packets (0 = infinite) |
| `--interval` | - | `1s` | Send interval |
| `--timeout` | - | `3s` | Per-packet timeout |
| `--size` | `-s` | `64` | Packet size in bytes |
| `--ttl` | - | `0` | TTL (0 = auto: unicast 64, multicast 16) |

### Recv Command

```bash
ping-x recv [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--proto` | `-p` | `tcp` | Protocol: tcp, udp, multicast, ssm |
| `--bind` | `-b` | `0.0.0.0` | Listen bind address |
| `--port` | - | - | Listen port (required) |
| `--group` | `-g` | - | Multicast group address (required for multicast/ssm) |
| `--source` | - | - | SSM source address (required for ssm recv) |
| `--iface` | `-i` | - | Network interface name (for multicast) |

### Config Command

```bash
ping-x config [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | stdout | Output file path |

## Configuration File

YAML configuration supports multiple test definitions:

```yaml
tests:
  - name: "tcp-basic"
    proto: tcp
    mode: send
    target: 192.168.1.100
    port: 9000
    count: 100
    interval: 1s
    timeout: 3s
    size: 64

  - name: "tcp-basic-recv"
    proto: tcp
    mode: recv
    port: 9000

  - name: "multicast-test"
    proto: multicast
    mode: send
    group: 239.1.1.1
    port: 9002
    iface: eth0
    count: 20
```

### Configuration Fields

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Test identifier | Yes |
| `proto` | Protocol: tcp, udp, multicast, ssm | Yes |
| `mode` | Mode: send or recv | Yes |
| `target` | Target IP address | tcp/udp send |
| `group` | Multicast group address | multicast/ssm |
| `source` | SSM source address | ssm recv |
| `port` | Port number | Yes |
| `bind` | Local bind address | No |
| `iface` | Network interface | multicast/ssm |
| `count` | Packet count (0 = infinite) | No |
| `interval` | Send interval | No |
| `timeout` | Per-packet timeout | No |
| `size` | Packet size in bytes | No |
| `ttl` | Time-to-live | No |

## Build from Source

### Requirements

- Go 1.17 or later

### Build Commands

```bash
# Build for current platform
go build -o ping-x .

# Build for Linux
make build-linux

# Build for Windows
make build-windows

# Build all release artifacts
make release
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ping-x_linux_amd64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ping-x_windows_amd64.exe .
```

## License

MIT License - see [LICENSE](LICENSE) for details.
