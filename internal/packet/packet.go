package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	// Magic 魔数标识 "PXNG"
	Magic uint32 = 0x50584E47
	// HeaderSize 固定头部大小: Magic(4) + Seq(4) + Timestamp(8) + Size(2) = 18 bytes
	HeaderSize = 18
)

// Probe 探测包结构
type Probe struct {
	Magic     uint32
	Seq       uint32
	Timestamp int64  // UnixNano
	Size      uint16 // payload 大小
	Payload   []byte
}

// NewProbe 创建新的探测包
func NewProbe(seq uint32, payloadSize int) *Probe {
	payload := make([]byte, payloadSize)
	// 填充简单模式数据用于校验
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	return &Probe{
		Magic:     Magic,
		Seq:       seq,
		Timestamp: time.Now().UnixNano(),
		Size:      uint16(payloadSize),
		Payload:   payload,
	}
}

// Marshal 序列化探测包为字节流
func (p *Probe) Marshal() ([]byte, error) {
	buf := make([]byte, HeaderSize+len(p.Payload))
	binary.BigEndian.PutUint32(buf[0:4], p.Magic)
	binary.BigEndian.PutUint32(buf[4:8], p.Seq)
	binary.BigEndian.PutUint64(buf[8:16], uint64(p.Timestamp))
	binary.BigEndian.PutUint16(buf[16:18], p.Size)
	copy(buf[HeaderSize:], p.Payload)
	return buf, nil
}

// Unmarshal 从字节流反序列化探测包
func Unmarshal(data []byte) (*Probe, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("data too short for probe header")
	}
	p := &Probe{
		Magic:     binary.BigEndian.Uint32(data[0:4]),
		Seq:       binary.BigEndian.Uint32(data[4:8]),
		Timestamp: int64(binary.BigEndian.Uint64(data[8:16])),
		Size:      binary.BigEndian.Uint16(data[16:18]),
	}
	if p.Magic != Magic {
		return nil, fmt.Errorf("invalid magic: 0x%X (expected 0x%X)", p.Magic, Magic)
	}
	if len(data) > HeaderSize {
		p.Payload = make([]byte, len(data)-HeaderSize)
		copy(p.Payload, data[HeaderSize:])
	}
	return p, nil
}

// RTT 计算从发送时间戳到现在的往返时间
func (p *Probe) RTT() time.Duration {
	return time.Duration(time.Now().UnixNano() - p.Timestamp)
}

// MakeAck 基于收到的 Probe 创建确认包（保留原始 Seq 和 Timestamp）
func (p *Probe) MakeAck() *Probe {
	return &Probe{
		Magic:     Magic,
		Seq:       p.Seq,
		Timestamp: p.Timestamp,
		Size:      0,
		Payload:   nil,
	}
}
