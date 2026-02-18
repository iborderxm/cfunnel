//go:build windows

package tun

import (
	"fmt"

	"golang.zx2c4.com/wireguard/tun"

	"github.com/fmnx/cftun/client/tun/core/device"
)

const (
	offset     = 0
	defaultMTU = 0 /* auto */
)

type TUN struct {
	device tun.Device
	mtu    uint32
	name   string
}

func openPlatform(name string, mtu uint32) (device.Device, error) {
	t := &TUN{
		name: name,
		mtu:  uint32(defaultMTU),
	}

	forcedMTU := defaultMTU
	if mtu > 0 {
		forcedMTU = int(mtu)
		t.mtu = mtu
	}

	nt, err := createTUN(t.name, forcedMTU)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}
	t.device = nt

	tunMTU, err := nt.MTU()
	if err != nil {
		return nil, fmt.Errorf("get mtu: %w", err)
	}
	t.mtu = uint32(tunMTU)

	return t, nil
}

func createTUN(name string, mtu int) (tun.Device, error) {
	return tun.CreateTUN(name, mtu)
}

func (t *TUN) Name() string {
	name, _ := t.device.Name()
	return name
}

func (t *TUN) Type() string {
	return Driver
}

func (t *TUN) Read(buf []byte) (int, error) {
	return t.device.Read(buf, offset)
}

func (t *TUN) Write(buf []byte) (int, error) {
	return t.device.Write(buf, offset)
}

func (t *TUN) Close() error {
	return t.device.Close()
}

func (t *TUN) MTU() int {
	return int(t.mtu)
}

func (t *TUN) Capabilities() []string {
	return []string{}
}

func (t *TUN) Attach(dispatcher device.LinkEndpointDispatcher) {
}

func (t *TUN) IsAttached() bool {
	return false
}

var _ device.Device = (*TUN)(nil)
