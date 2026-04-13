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

// SSMSender SSM 指定源多播发送端
type SSMSender struct{}

// SSMReceiver SSM 指定源多播接收端
type SSMReceiver struct{}

// Send 实现 Sender 接口，发送 SSM 多播探测包（单向，不等 Ack）
func (s *SSMSender) Send(ctx context.Context, cfg *config.TestConfig, stat *stats.Collector) error {
	target := fmt.Sprintf("%s:%d", cfg.Group, cfg.Port)

	dstAddr, err := net.ResolveUDPAddr("udp4", target)
	if err != nil {
		return fmt.Errorf("SSM resolve addr %s failed: %w", target, err)
	}

	// 如果指定了绑定源地址，使用该地址作为本地地址
	var laddr *net.UDPAddr
	if cfg.Bind != "" && cfg.Bind != "0.0.0.0" {
		laddr, err = net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:0", cfg.Bind))
		if err != nil {
			return fmt.Errorf("SSM resolve bind addr %s failed: %w", cfg.Bind, err)
		}
	}

	conn, err := net.DialUDP("udp4", laddr, dstAddr)
	if err != nil {
		return fmt.Errorf("SSM dial %s failed: %w", target, err)
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
			fmt.Print(stat.Report("SSM", target))
			return nil
		default:
		}

		probe := packet.NewProbe(seq, cfg.Size)
		data, err := probe.Marshal()
		if err != nil {
			stat.RecordSend()
			fmt.Printf("[%s] SSM #%d -> %s  FAIL marshal error\n",
				time.Now().Format("15:04:05"), seq, target)
		} else {
			stat.RecordSend()
			_, writeErr := conn.Write(data)
			if writeErr != nil {
				fmt.Printf("[%s] SSM #%d -> %s  FAIL send error\n",
					time.Now().Format("15:04:05"), seq, target)
			} else {
				fmt.Printf("[%s] SSM #%d -> %s  SENT  (source=%s)\n",
					time.Now().Format("15:04:05"), seq, target, cfg.Bind)
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
			fmt.Print(stat.Report("SSM", target))
			return nil
		case <-time.After(cfg.GetInterval()):
		}
	}

	fmt.Print(stat.Report("SSM", target))
	return nil
}

// Receive 实现 Receiver 接口，接收 SSM 多播探测包
func (r *SSMReceiver) Receive(ctx context.Context, cfg *config.TestConfig) error {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	conn, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		return fmt.Errorf("SSM listen on %s failed: %w", listenAddr, err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)

	var iface *net.Interface
	if cfg.Iface != "" {
		iface, _ = net.InterfaceByName(cfg.Iface)
	}

	group := &net.UDPAddr{IP: net.ParseIP(cfg.Group)}
	source := &net.UDPAddr{IP: net.ParseIP(cfg.Source)}

	if err := p.JoinSourceSpecificGroup(iface, group, source); err != nil {
		return fmt.Errorf("SSM join source specific group %s source %s failed: %w",
			cfg.Group, cfg.Source, err)
	}
	defer p.LeaveSourceSpecificGroup(iface, group, source)

	fmt.Printf("Listening on SSM group %s:%d source=%s ...\n", cfg.Group, cfg.Port, cfg.Source)

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
				return fmt.Errorf("SSM read error: %w", err)
			}
		}

		probe, err := packet.Unmarshal(buf[:n])
		if err != nil {
			continue
		}

		fmt.Printf("[%s] SSM recv #%d from %s  group=%s source=%s\n",
			time.Now().Format("15:04:05"), probe.Seq, addr.String(), cfg.Group, cfg.Source)
	}
}
