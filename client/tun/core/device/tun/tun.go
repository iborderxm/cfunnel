package tun

import (
	"github.com/fmnx/cftun/client/tun/core/device"
)

const (
	// Driver is the driver name for TUN devices.
	Driver = "tun"
)

func Open(name string, mtu uint32) (device.Device, error) {
	return openPlatform(name, mtu)
}
