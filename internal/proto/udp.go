package proto

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/packet"
	"github.com/yanzhen74/ping-x/internal/stats"
)

// UDPSender UDP 发送端
type UDPSender struct{}

// UDPReceiver UDP 接收端
type UDPReceiver struct{}

// Send 实现 Sender 接口，发送 UDP 探测包
func (s *UDPSender) Send(ctx context.Context, cfg *config.TestConfig, stat *stats.Collector) error {
	target := fmt.Sprintf("%s:%d", cfg.Target, cfg.Port)
	timeout := cfg.GetTimeout()

	dstAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return fmt.Errorf("UDP resolve addr %s failed: %w", target, err)
	}

	conn, err := net.DialUDP("udp", nil, dstAddr)
	if err != nil {
		return fmt.Errorf("UDP dial %s failed: %w", target, err)
	}
	defer conn.Close()

	var seq uint32 = 1
	count := cfg.Count

	for {
		// 检查 ctx
		select {
		case <-ctx.Done():
			fmt.Print(stat.Report("UDP", target))
			return nil
		default:
		}

		// 创建并发送探测包
		probe := packet.NewProbe(seq, cfg.Size)
		data, err := probe.Marshal()
		if err != nil {
			stat.RecordSend()
			stat.RecordLoss()
			fmt.Printf("[%s] UDP #%d -> %s  FAIL marshal error\n",
				time.Now().Format("15:04:05"), seq, target)
		} else {
			stat.RecordSend()
			_, writeErr := conn.Write(data)
			if writeErr != nil {
				stat.RecordLoss()
				fmt.Printf("[%s] UDP #%d -> %s  FAIL send error\n",
					time.Now().Format("15:04:05"), seq, target)
			} else {
				// 等待 Ack
				conn.SetReadDeadline(time.Now().Add(timeout))
				buf := make([]byte, 65535)
				n, readErr := conn.Read(buf)
				if readErr != nil {
					stat.RecordLoss()
					fmt.Printf("[%s] UDP #%d -> %s  FAIL timeout\n",
						time.Now().Format("15:04:05"), seq, target)
				} else {
					ack, unmarshalErr := packet.Unmarshal(buf[:n])
					if unmarshalErr != nil {
						stat.RecordLoss()
						fmt.Printf("[%s] UDP #%d -> %s  FAIL invalid ack\n",
							time.Now().Format("15:04:05"), seq, target)
					} else {
						rtt := ack.RTT()
						stat.RecordReceive(rtt)
						fmt.Printf("[%s] UDP #%d -> %s  OK  rtt=%s\n",
							time.Now().Format("15:04:05"), seq, target,
							fmt.Sprintf("%.2fms", float64(rtt.Nanoseconds())/1e6))
					}
				}
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
			fmt.Print(stat.Report("UDP", target))
			return nil
		case <-time.After(cfg.GetInterval()):
		}
	}

	fmt.Print(stat.Report("UDP", target))
	return nil
}

// Receive 实现 Receiver 接口，接收 UDP 探测包并回复 Ack
func (r *UDPReceiver) Receive(ctx context.Context, cfg *config.TestConfig) error {
	bind := cfg.Bind
	if bind == "" {
		bind = "0.0.0.0"
	}

	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", bind, cfg.Port))
	if err != nil {
		return fmt.Errorf("UDP resolve listen addr failed: %w", err)
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("UDP listen on %s:%d failed: %w", bind, cfg.Port, err)
	}
	defer conn.Close()

	fmt.Printf("Listening on udp://%s:%d ...\n", bind, cfg.Port)

	// ctx 取消时关闭连接
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("UDP read error: %w", err)
			}
		}

		probe, err := packet.Unmarshal(buf[:n])
		if err != nil {
			continue
		}

		fmt.Printf("[%s] UDP recv #%d from %s\n",
			time.Now().Format("15:04:05"), probe.Seq, remoteAddr.String())

		ack := probe.MakeAck()
		ackData, err := ack.Marshal()
		if err != nil {
			continue
		}

		conn.WriteToUDP(ackData, remoteAddr)
	}
}
