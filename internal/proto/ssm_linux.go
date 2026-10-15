//go:build linux
// +build linux

package proto

import (
	"net"
	"syscall"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

func ssmJoinGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr, bindIP string) error {
	p := ipv4.NewPacketConn(conn)
	return p.JoinSourceSpecificGroup(iface, group, source)
}

func ssmLeaveGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr, bindIP string) error {
	p := ipv4.NewPacketConn(conn)
	return p.LeaveSourceSpecificGroup(iface, group, source)
}

// setSocketReuse 设置 SO_REUSEADDR 和 SO_REUSEPORT，允许多播端口重用
func setSocketReuse(fd uintptr) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return err
	}
	return nil
}
