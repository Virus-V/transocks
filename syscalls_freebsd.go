//go:build !linux
// +build !linux

package main

import (
	"fmt"
	"net"
	"syscall"
)

// GetOriginalDST retrieves the original destination address from
// NATed connection.  Currently, only Linux iptables using DNAT/REDIRECT
// is supported.  For other operating systems, this will just return
// conn.LocalAddr().
//
// Note that this function only works when nf_conntrack_ipv4 and/or
// nf_conntrack_ipv6 is loaded in the kernel.
func GetOriginalDST(conn *net.TCPConn) (*net.TCPAddr, error) {
	return conn.LocalAddr().(*net.TCPAddr), nil
}

// enable bind any
func SetTransparentListener(sysConn syscall.RawConn) error {
	var err, sockErr error

	/* enable ip bindany */
	err = sysConn.Control(func(fd uintptr) {
		fmt.Println("set ip bind any")
		sockErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BINDANY, 1)
	})

	if err != nil {
		return err
	}

	return sockErr
}
