package adapter

import (
	"net"
)

type EndpointID struct {
	LocalAddress  string
	RemoteAddress string
	LocalPort     uint16
	RemotePort    uint16
}

// TCPConn implements the net.Conn interface.
type TCPConn interface {
	net.Conn

	// ID returns the transport endpoint id of TCPConn.
	ID() *EndpointID
}

// UDPConn implements net.Conn and net.PacketConn.
type UDPConn interface {
	net.Conn
	net.PacketConn

	// ID returns the transport endpoint id of UDPConn.
	ID() *EndpointID
}
