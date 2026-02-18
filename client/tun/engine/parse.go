package engine

import (
	"fmt"
	"net/netip"
	"net/url"
	"runtime"
	"strings"

	"github.com/fmnx/cftun/client/tun/core/device"
	"github.com/fmnx/cftun/client/tun/core/device/tun"
)

func parseDevice(s string, mtu uint32) (device.Device, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", tun.Driver /* default driver */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	driver := strings.ToLower(u.Scheme)

	switch driver {
	case tun.Driver:
		return parseTUN(u, mtu)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func parseTUN(u *url.URL, mtu uint32) (device.Device, error) {
	return tun.Open(u.Host, mtu)
}

func parseMulticastGroups(s string) (multicastGroups []netip.Addr, _ error) {
	for _, ip := range strings.Split(s, ",") {
		if ip = strings.TrimSpace(ip); ip == "" {
			continue
		}
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, err
		}
		if !addr.IsMulticast() {
			return nil, fmt.Errorf("invalid multicast IP: %s", addr)
		}
		multicastGroups = append(multicastGroups, addr)
	}
	return
}
