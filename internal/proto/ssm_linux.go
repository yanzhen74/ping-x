//go:build linux
// +build linux

package proto

import (
	"net"

	"golang.org/x/net/ipv4"
)

func ssmJoinGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr) error {
	p := ipv4.NewPacketConn(conn)
	return p.JoinSourceSpecificGroup(iface, group, source)
}

func ssmLeaveGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr) error {
	p := ipv4.NewPacketConn(conn)
	return p.LeaveSourceSpecificGroup(iface, group, source)
}
