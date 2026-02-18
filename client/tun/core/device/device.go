package device

// LinkEndpoint is the interface for network link layer endpoints.
type LinkEndpoint interface {
	// Write writes a packet to the endpoint.
	Write([]byte) (int, error)

	// Read reads a packet from the endpoint.
	Read([]byte) (int, error)

	// Close closes the endpoint.
	Close() error

	// MTU returns the maximum transmission unit.
	MTU() int

	// Capabilities returns the endpoint capabilities.
	Capabilities() []string

	// Attach attaches a dispatcher to the endpoint.
	Attach(LinkEndpointDispatcher)

	// IsAttached returns whether a dispatcher is attached.
	IsAttached() bool
}

// LinkEndpointDispatcher is the interface for dispatching packets.
type LinkEndpointDispatcher interface {
	DeliverNetworkPacket(LinkEndpoint, []byte)
}

// Device is the interface that implemented by network layer devices (e.g. tun),
// and easy to use as LinkEndpoint.
type Device interface {
	LinkEndpoint

	// Name returns the current name of the device.
	Name() string

	// Type returns the driver type of the device.
	Type() string
}
