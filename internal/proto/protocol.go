package proto

import (
	"context"
	"fmt"

	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/stats"
)

// Sender 发送端接口
type Sender interface {
	Send(ctx context.Context, cfg *config.TestConfig, stat *stats.Collector) error
}

// Receiver 接收端接口
type Receiver interface {
	Receive(ctx context.Context, cfg *config.TestConfig) error
}

// NewSender 根据协议类型创建 Sender
func NewSender(proto string) (Sender, error) {
	switch proto {
	case "tcp":
		return &TCPSender{}, nil
	case "udp":
		return &UDPSender{}, nil
	case "multicast":
		return &MulticastSender{}, nil
	case "ssm":
		return &SSMSender{}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

// NewReceiver 根据协议类型创建 Receiver
func NewReceiver(proto string) (Receiver, error) {
	switch proto {
	case "tcp":
		return &TCPReceiver{}, nil
	case "udp":
		return &UDPReceiver{}, nil
	case "multicast":
		return &MulticastReceiver{}, nil
	case "ssm":
		return &SSMReceiver{}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}
