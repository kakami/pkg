package rmet

import (
    "net"
    "net/netip"
)

type MUDPConn struct {
    *net.UDPConn
    *RMet
}

func NewUDPConn(conn *net.UDPConn, n int) *MUDPConn {
    return &MUDPConn{
        UDPConn: conn,
        RMet:    New(n),
    }
}

func (c *MUDPConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
    n, addr, err := c.UDPConn.ReadFromUDP(b)
    c.AddBytesRecv(int64(n))
    return n, addr, err
}

func (c *MUDPConn) ReadFrom(b []byte) (int, net.Addr, error) {
    n, addr, err := c.UDPConn.ReadFrom(b)
    c.AddBytesRecv(int64(n))
    return n, addr, err
}

func (c *MUDPConn) ReadFromUDPAddrPort(b []byte) (int, netip.AddrPort, error) {
    n, addr, err := c.UDPConn.ReadFromUDPAddrPort(b)
    c.AddBytesRecv(int64(n))
    return n, addr, err
}

func (c *MUDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
    n, oobn, flags, addr, err = c.UDPConn.ReadMsgUDP(b, oob)
    c.AddBytesRecv(int64(n))
    return
}

func (c *MUDPConn) ReadMsgUDPAddrPort(b, oob []byte) (n, oobn, flags int, addr netip.AddrPort, err error) {
    n, oobn, flags, addr, err = c.UDPConn.ReadMsgUDPAddrPort(b, oob)
    c.AddBytesRecv(int64(n))
    return
}

func (c *MUDPConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
    c.AddBytesSent(int64(len(b)))
    return c.UDPConn.WriteToUDP(b, addr)
}

func (c *MUDPConn) WriteToUDPAddrPort(b []byte, addr netip.AddrPort) (int, error) {
    c.AddBytesSent(int64(len(b)))
    return c.UDPConn.WriteToUDPAddrPort(b, addr)
}

func (c *MUDPConn) WriteTo(b []byte, addr net.Addr) (int, error) {
    c.AddBytesSent(int64(len(b)))
    return c.UDPConn.WriteTo(b, addr)
}

func (c *MUDPConn) WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error) {
    n, oobn, err = c.UDPConn.WriteMsgUDP(b, oob, addr)
    c.AddBytesSent(int64(n))
    return
}

func (c *MUDPConn) WriteMsgUDPAddrPort(b, oob []byte, addr netip.AddrPort) (n, oobn int, err error) {
    n, oobn, err = c.UDPConn.WriteMsgUDPAddrPort(b, oob, addr)
    c.AddBytesSent(int64(n))
    return
}

func (c *MUDPConn) Write(b []byte) (int, error) {
    n, err := c.UDPConn.Write(b)
    c.AddBytesSent(int64(n))
    return n, err
}

func (c *MUDPConn) Read(b []byte) (int, error) {
    n, err := c.UDPConn.Read(b)
    c.AddBytesRecv(int64(n))
    return n, err
}
