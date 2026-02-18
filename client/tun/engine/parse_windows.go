package engine

import (
	"net/url"

	"golang.org/x/sys/windows"
	wun "golang.zx2c4.com/wireguard/tun"

	"github.com/fmnx/cftun/client/tun/core/device"
	"github.com/fmnx/cftun/client/tun/core/device/tun"
)

func init() {
	wun.WintunTunnelType = "argotunnel"
}

func parseTUN(u *url.URL, mtu uint32) (device.Device, error) {
	guid := u.Query().Get("guid")
	if guid != "" {
		guidValue, err := windows.GUIDFromString(guid)
		if err != nil {
			return nil, err
		}
		wun.WintunStaticRequestedGUID = &guidValue
	}
	return tun.Open(u.Host, mtu)
}
