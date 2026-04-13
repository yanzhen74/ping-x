package proto

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/packet"
	"github.com/yanzhen74/ping-x/internal/stats"
)

// TCPSender TCP 发送端
type TCPSender struct{}

// TCPReceiver TCP 接收端
type TCPReceiver struct{}

// tcpWriteFrame 写入 4 字节长度头 + 数据
func tcpWriteFrame(conn net.Conn, data []byte) error {
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint32(hdr, uint32(len(data)))
	if _, err := conn.Write(hdr); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

// tcpReadFrame 读取 4 字节长度头，再读取对应字节数据
func tcpReadFrame(conn net.Conn) ([]byte, error) {
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(hdr)
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	_, err := io.ReadFull(conn, buf)
	return buf, err
}

// Send 实现 Sender 接口，发送 TCP 探测包
func (s *TCPSender) Send(ctx context.Context, cfg *config.TestConfig, stat *stats.Collector) error {
	target := fmt.Sprintf("%s:%d", cfg.Target, cfg.Port)
	timeout := cfg.GetTimeout()

	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return fmt.Errorf("TCP connect to %s failed: %w", target, err)
	}
	defer conn.Close()

	var seq uint32 = 1
	count := cfg.Count

	for {
		// 检查 ctx
		select {
		case <-ctx.Done():
			fmt.Print(stat.Report("TCP", target))
			return nil
		default:
		}

		// 创建并发送探测包
		probe := packet.NewProbe(seq, cfg.Size)
		data, err := probe.Marshal()
		if err != nil {
			stat.RecordSend()
			stat.RecordLoss()
			fmt.Printf("[%s] TCP #%d -> %s  FAIL marshal error\n",
				time.Now().Format("15:04:05"), seq, target)
		} else {
			stat.RecordSend()
			writeErr := tcpWriteFrame(conn, data)
			if writeErr != nil {
				stat.RecordLoss()
				fmt.Printf("[%s] TCP #%d -> %s  FAIL send error\n",
					time.Now().Format("15:04:05"), seq, target)
			} else {
				// 等待 Ack
				conn.SetReadDeadline(time.Now().Add(timeout))
				ackData, readErr := tcpReadFrame(conn)
				if readErr != nil {
					stat.RecordLoss()
					fmt.Printf("[%s] TCP #%d -> %s  FAIL timeout\n",
						time.Now().Format("15:04:05"), seq, target)
				} else {
					ack, unmarshalErr := packet.Unmarshal(ackData)
					if unmarshalErr != nil {
						stat.RecordLoss()
						fmt.Printf("[%s] TCP #%d -> %s  FAIL invalid ack\n",
							time.Now().Format("15:04:05"), seq, target)
					} else {
						rtt := ack.RTT()
						stat.RecordReceive(rtt)
						fmt.Printf("[%s] TCP #%d -> %s  OK  rtt=%s\n",
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
			fmt.Print(stat.Report("TCP", target))
			return nil
		case <-time.After(cfg.GetInterval()):
		}
	}

	fmt.Print(stat.Report("TCP", target))
	return nil
}

// Receive 实现 Receiver 接口，接收 TCP 探测包并回复 Ack
func (r *TCPReceiver) Receive(ctx context.Context, cfg *config.TestConfig) error {
	bind := cfg.Bind
	if bind == "" {
		bind = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", bind, cfg.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP listen on %s failed: %w", addr, err)
	}

	fmt.Printf("Listening on tcp://%s:%d ...\n", bind, cfg.Port)

	// ctx 取消时关闭 listener
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("TCP accept error: %w", err)
			}
		}
		go tcpHandleConn(conn)
	}
}

// tcpHandleConn 处理单个 TCP 连接
func tcpHandleConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	for {
		data, err := tcpReadFrame(conn)
		if err != nil {
			// 连接断开，正常退出
			return
		}

		probe, err := packet.Unmarshal(data)
		if err != nil {
			// 无效数据，忽略
			continue
		}

		fmt.Printf("[%s] TCP recv #%d from %s\n",
			time.Now().Format("15:04:05"), probe.Seq, remoteAddr)

		ack := probe.MakeAck()
		ackData, err := ack.Marshal()
		if err != nil {
			continue
		}

		if err := tcpWriteFrame(conn, ackData); err != nil {
			return
		}
	}
}
