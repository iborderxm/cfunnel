package native

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/fmnx/cftun/client/tun/core/adapter"
	"github.com/fmnx/cftun/client/tun/log"
	"github.com/fmnx/cftun/client/tun/tunnel"
)

const (
	IPv4HeaderLen = 20
	IPv6HeaderLen = 40
	TCPHeaderLen  = 20
	UDPHeaderLen  = 8
	ICMPHeaderLen = 8

	ProtocolICMP = 1
	ProtocolTCP  = 6
	ProtocolUDP  = 17
)

type NativeStack struct {
	device      Device
	tunnel      *tunnel.Tunnel
	mtu         int
	stopChan    chan struct{}
	wg          sync.WaitGroup
	udpSessions map[string]*udpSession
	udpMu       sync.RWMutex
}

type Device interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	MTU() int
	Name() string
	Type() string
}

type udpSession struct {
	conn      net.Conn
	lastSeen  int64
	timeout   int64
}

func New(device Device, tunnel *tunnel.Tunnel, mtu int) *NativeStack {
	return &NativeStack{
		device:      device,
		tunnel:      tunnel,
		mtu:         mtu,
		stopChan:    make(chan struct{}),
		udpSessions: make(map[string]*udpSession),
	}
}

func (s *NativeStack) Start() error {
	s.wg.Add(1)
	go s.readLoop()
	return nil
}

func (s *NativeStack) Stop() error {
	close(s.stopChan)
	s.wg.Wait()
	return nil
}

func (s *NativeStack) readLoop() {
	defer s.wg.Done()

	buf := make([]byte, s.mtu+128)
	for {
		select {
		case <-s.stopChan:
			return
		default:
			n, err := s.device.Read(buf)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Errorf("[NATIVE] read error: %v", err)
				}
				return
			}

			packet := buf[:n]
			if err := s.handlePacket(packet); err != nil {
				log.Debugf("[NATIVE] handle packet error: %v", err)
			}
		}
	}
}

func (s *NativeStack) handlePacket(packet []byte) error {
	if len(packet) < IPv4HeaderLen {
		return errors.New("packet too short")
	}

	version := packet[0] >> 4
	switch version {
	case 4:
		return s.handleIPv4(packet)
	case 6:
		return s.handleIPv6(packet)
	default:
		return errors.New("unsupported IP version")
	}
}

func (s *NativeStack) handleIPv4(packet []byte) error {
	if len(packet) < IPv4HeaderLen {
		return errors.New("IPv4 packet too short")
	}

	ihl := int(packet[0]&0x0F) * 4
	if len(packet) < ihl {
		return errors.New("invalid IPv4 header length")
	}

	protocol := packet[9]
	srcIP := net.IP(packet[12:16]).To4()
	dstIP := net.IP(packet[16:20]).To4()

	payload := packet[ihl:]
	if len(payload) == 0 {
		return nil
	}

	switch protocol {
	case ProtocolTCP:
		return s.handleTCP(payload, srcIP, dstIP)
	case ProtocolUDP:
		return s.handleUDP(payload, srcIP, dstIP)
	case ProtocolICMP:
		return s.handleICMP(payload, srcIP, dstIP)
	default:
		return nil
	}
}

func (s *NativeStack) handleIPv6(packet []byte) error {
	if len(packet) < IPv6HeaderLen {
		return errors.New("IPv6 packet too short")
	}

	nextHeader := packet[6]
	srcIP := net.IP(packet[8:24])
	dstIP := net.IP(packet[24:40])

	payload := packet[IPv6HeaderLen:]
	if len(payload) == 0 {
		return nil
	}

	switch nextHeader {
	case ProtocolTCP:
		return s.handleTCP(payload, srcIP, dstIP)
	case ProtocolUDP:
		return s.handleUDP(payload, srcIP, dstIP)
	case ProtocolICMP:
		return s.handleICMP(payload, srcIP, dstIP)
	default:
		return nil
	}
}

func (s *NativeStack) handleTCP(payload []byte, srcIP, dstIP net.IP) error {
	if len(payload) < TCPHeaderLen {
		return errors.New("TCP packet too short")
	}

	srcPort := binary.BigEndian.Uint16(payload[0:2])
	dstPort := binary.BigEndian.Uint16(payload[2:4])
	seqNum := binary.BigEndian.Uint32(payload[4:8])
	ackNum := binary.BigEndian.Uint32(payload[8:12])
	flags := payload[13]

	isSyn := flags&0x02 != 0
	isFin := flags&0x01 != 0
	isRst := flags&0x04 != 0

	srcAddr := &net.TCPAddr{IP: srcIP, Port: int(srcPort)}
	dstAddr := &net.TCPAddr{IP: dstIP, Port: int(dstPort)}

	conn := newNativeTCPConn(srcAddr, dstAddr)
	conn.seqNum = seqNum
	conn.ackNum = ackNum
	conn.isSyn = isSyn
	conn.isFin = isFin
	conn.isRst = isRst

	s.tunnel.HandleTCP(conn)
	return nil
}

func (s *NativeStack) handleUDP(payload []byte, srcIP, dstIP net.IP) error {
	if len(payload) < UDPHeaderLen {
		return errors.New("UDP packet too short")
	}

	srcPort := binary.BigEndian.Uint16(payload[0:2])
	dstPort := binary.BigEndian.Uint16(payload[2:4])
	length := binary.BigEndian.Uint16(payload[4:6])
	data := payload[UDPHeaderLen:]

	if int(length) > len(data) {
		data = data[:length]
	}

	srcAddr := &net.UDPAddr{IP: srcIP, Port: int(srcPort)}
	dstAddr := &net.UDPAddr{IP: dstIP, Port: int(dstPort)}

	conn := newNativeUDPConn(srcAddr, dstAddr, data, s)
	s.tunnel.HandleUDP(conn)
	return nil
}

func (s *NativeStack) handleICMP(payload []byte, srcIP, dstIP net.IP) error {
	return nil
}

func (s *NativeStack) cleanupUDPSessions() {
	s.udpMu.Lock()
	defer s.udpMu.Unlock()

	now := time.Now().Unix()
	for key, session := range s.udpSessions {
		if now-session.lastSeen > session.timeout {
			session.conn.Close()
			delete(s.udpSessions, key)
		}
	}
}

type nativeTCPConn struct {
	srcAddr *net.TCPAddr
	dstAddr *net.TCPAddr
	seqNum  uint32
	ackNum  uint32
	isSyn   bool
	isFin   bool
	isRst   bool
	id      *adapter.EndpointID
}

func newNativeTCPConn(srcAddr, dstAddr *net.TCPAddr) *nativeTCPConn {
	return &nativeTCPConn{
		srcAddr: srcAddr,
		dstAddr: dstAddr,
		id: &adapter.EndpointID{
			LocalAddress:  dstAddr.IP.String(),
			RemoteAddress: srcAddr.IP.String(),
			LocalPort:     uint16(dstAddr.Port),
			RemotePort:    uint16(srcAddr.Port),
		},
	}
}

func (c *nativeTCPConn) ID() *adapter.EndpointID {
	return c.id
}

func (c *nativeTCPConn) LocalAddr() net.Addr {
	return c.srcAddr
}

func (c *nativeTCPConn) RemoteAddr() net.Addr {
	return c.dstAddr
}

func (c *nativeTCPConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *nativeTCPConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c *nativeTCPConn) Close() error {
	return nil
}

func (c *nativeTCPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *nativeTCPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *nativeTCPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type nativeUDPConn struct {
	srcAddr *net.UDPAddr
	dstAddr *net.UDPAddr
	data    []byte
	stack   *NativeStack
	id      *adapter.EndpointID
}

func newNativeUDPConn(srcAddr, dstAddr *net.UDPAddr, data []byte, stack *NativeStack) *nativeUDPConn {
	return &nativeUDPConn{
		srcAddr: srcAddr,
		dstAddr: dstAddr,
		data:    data,
		stack:   stack,
		id: &adapter.EndpointID{
			LocalAddress:  dstAddr.IP.String(),
			RemoteAddress: srcAddr.IP.String(),
			LocalPort:     uint16(dstAddr.Port),
			RemotePort:    uint16(srcAddr.Port),
		},
	}
}

func (c *nativeUDPConn) ID() *adapter.EndpointID {
	return c.id
}

func (c *nativeUDPConn) LocalAddr() net.Addr {
	return c.srcAddr
}

func (c *nativeUDPConn) RemoteAddr() net.Addr {
	return c.dstAddr
}

func (c *nativeUDPConn) Read(b []byte) (n int, err error) {
	return copy(b, c.data), nil
}

func (c *nativeUDPConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c *nativeUDPConn) Close() error {
	return nil
}

func (c *nativeUDPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *nativeUDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *nativeUDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
