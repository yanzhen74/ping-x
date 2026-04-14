//go:build windows
// +build windows

package proto

/*
#cgo windows LDFLAGS: -lws2_32

#include <winsock2.h>
#include <ws2tcpip.h>

struct my_ip_mreq_source {
    struct in_addr imr_multiaddr;
    struct in_addr imr_sourceaddr;
    struct in_addr imr_interface;
};

static int ssm_join(SOCKET fd, const char* group_ip, const char* source_ip, const char* iface_ip) {
    struct my_ip_mreq_source mreq;
    memset(&mreq, 0, sizeof(mreq));
    mreq.imr_multiaddr.s_addr  = inet_addr(group_ip);
    mreq.imr_sourceaddr.s_addr = inet_addr(source_ip);
    if (iface_ip != NULL && iface_ip[0] != '\0') {
        mreq.imr_interface.s_addr = inet_addr(iface_ip);
    } else {
        mreq.imr_interface.s_addr = INADDR_ANY;
    }
    return setsockopt(fd, IPPROTO_IP, IP_ADD_SOURCE_MEMBERSHIP,
                      (const char*)&mreq, sizeof(mreq));
}

static int ssm_leave(SOCKET fd, const char* group_ip, const char* source_ip, const char* iface_ip) {
    struct my_ip_mreq_source mreq;
    memset(&mreq, 0, sizeof(mreq));
    mreq.imr_multiaddr.s_addr  = inet_addr(group_ip);
    mreq.imr_sourceaddr.s_addr = inet_addr(source_ip);
    if (iface_ip != NULL && iface_ip[0] != '\0') {
        mreq.imr_interface.s_addr = inet_addr(iface_ip);
    } else {
        mreq.imr_interface.s_addr = INADDR_ANY;
    }
    return setsockopt(fd, IPPROTO_IP, IP_DROP_SOURCE_MEMBERSHIP,
                      (const char*)&mreq, sizeof(mreq));
}

static int get_last_wsa_error() {
    return WSAGetLastError();
}
*/
import "C"

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

func ssmJoinGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr, bindIP string) error {
	return cgoSSMOpt(conn, iface, group, source, bindIP, true)
}

func ssmLeaveGroup(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr, bindIP string) error {
	return cgoSSMOpt(conn, iface, group, source, bindIP, false)
}

func cgoSSMOpt(conn net.PacketConn, iface *net.Interface, group, source *net.UDPAddr, bindIP string, join bool) error {
	groupStr := group.IP.To4().String()
	sourceStr := source.IP.To4().String()

	// 获取接口IPv4地址，优先使用bindIP
	ifaceStr := bindIP
	if ifaceStr == "" {
		ifaceStr = getIfaceIPv4(iface)
	}

	cGroup := C.CString(groupStr)
	cSource := C.CString(sourceStr)
	cIface := C.CString(ifaceStr)
	defer C.free(unsafe.Pointer(cGroup))
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cIface))

	// 获取底层 RawConn
	type syscallConn interface {
		SyscallConn() (syscall.RawConn, error)
	}
	sc, ok := conn.(syscallConn)
	if !ok {
		return fmt.Errorf("connection does not support SyscallConn")
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return fmt.Errorf("failed to get raw conn: %w", err)
	}

	var retErr error
	err = rc.Control(func(fd uintptr) {
		var ret C.int
		if join {
			ret = C.ssm_join(C.SOCKET(fd), cGroup, cSource, cIface)
		} else {
			ret = C.ssm_leave(C.SOCKET(fd), cGroup, cSource, cIface)
		}
		if ret != 0 {
			wsaErr := C.get_last_wsa_error()
			action := "join"
			if !join {
				action = "leave"
			}
			retErr = fmt.Errorf("SSM %s setsockopt failed, WSA error: %d", action, int(wsaErr))
		}
	})
	if err != nil {
		return err
	}
	return retErr
}

// getIfaceIPv4 获取网络接口的第一个 IPv4 地址字符串
func getIfaceIPv4(iface *net.Interface) string {
	if iface == nil {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}
