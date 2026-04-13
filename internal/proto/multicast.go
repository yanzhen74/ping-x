package proto

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/ipv4"

	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/packet"
	"github.com/yanzhen74/ping-x/internal/stats"
)

// MulticastSender ASM 多播发送端
type MulticastSender struct{}

// MulticastReceiver ASM 多播接收端
type MulticastReceiver struct{}

// Send 实现 Sender 接口，发送 ASM 多播探测包（单向，不等 Ack）
func (s *MulticastSender) Send(ctx context.Context, cfg *config.TestConfig, stat *stats.Collector) error {
	target := fmt.Sprintf("%s:%d", cfg.Group, cfg.Port)

	dstAddr, err := net.ResolveUDPAddr("udp4", target)
	if err != nil {
		return fmt.Errorf("MULTICAST resolve addr %s failed: %w", target, err)
	}

	conn, err := net.DialUDP("udp4", nil, dstAddr)
	if err != nil {
		return fmt.Errorf("MULTICAST dial %s failed: %w", target, err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)

	// 设置 MulticastTTL
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 16
	}
	p.SetMulticastTTL(ttl)

	// 设置多播接口
	if cfg.Iface != "" {
		iface, err := net.InterfaceByName(cfg.Iface)
		if err == nil {
			p.SetMulticastInterface(iface)
		}
	}

	var seq uint32 = 1
	count := cfg.Count

	for {
		// 检查 ctx
		select {
		case <-ctx.Done():
			fmt.Print(stat.Report("MULTICAST", target))
			return nil
		default:
		}

		probe := packet.NewProbe(seq, cfg.Size)
		data, err := probe.Marshal()
		if err != nil {
			stat.RecordSend()
			fmt.Printf("[%s] MULTICAST #%d -> %s  FAIL marshal error\n",
				time.Now().Format("15:04:05"), seq, target)
		} else {
			stat.RecordSend()
			_, writeErr := conn.Write(data)
			if writeErr != nil {
				fmt.Printf("[%s] MULTICAST #%d -> %s  FAIL send error\n",
					time.Now().Format("15:04:05"), seq, target)
			} else {
				fmt.Printf("[%s] MULTICAST #%d -> %s  SENT\n",
					time.Now().Format("15:04:05"), seq, target)
			}
		}

		seq++

		// 检查是否达到发送次数
		if count > 0 && int(seq-1) >= count {
			break
		}

		// 等待间隔，期间检查 ctx
		select {
		case <-ctx.Done():
			fmt.Print(stat.Report("MULTICAST", target))
			return nil
		case <-time.After(cfg.GetInterval()):
		}
	}

	fmt.Print(stat.Report("MULTICAST", target))
	return nil
}

// Receive 实现 Receiver 接口，接收 ASM 多播探测包
func (r *MulticastReceiver) Receive(ctx context.Context, cfg *config.TestConfig) error {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	conn, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		return fmt.Errorf("MULTICAST listen on %s failed: %w", listenAddr, err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)

	var iface *net.Interface
	if cfg.Iface != "" {
		iface, _ = net.InterfaceByName(cfg.Iface)
	}

	group := &net.UDPAddr{IP: net.ParseIP(cfg.Group)}
	if err := p.JoinGroup(iface, group); err != nil {
		return fmt.Errorf("MULTICAST join group %s failed: %w", cfg.Group, err)
	}
	defer p.LeaveGroup(iface, group)

	fmt.Printf("Listening on multicast group %s:%d ...\n", cfg.Group, cfg.Port)

	// ctx 取消时关闭连接
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("MULTICAST read error: %w", err)
			}
		}

		probe, err := packet.Unmarshal(buf[:n])
		if err != nil {
			continue
		}

		fmt.Printf("[%s] MULTICAST recv #%d from %s  group=%s\n",
			time.Now().Format("15:04:05"), probe.Seq, addr.String(), cfg.Group)
	}
}
