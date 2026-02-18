//go:build linux

package tun

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/fmnx/cftun/client/tun/core/device"
)

type TUN struct {
	fd   *os.File
	mtu  uint32
	name string
}

func openPlatform(name string, mtu uint32) (device.Device, error) {
	t := &TUN{name: name, mtu: mtu}

	if len(t.name) >= unix.IFNAMSIZ {
		return nil, fmt.Errorf("interface name too long: %s", t.name)
	}

	fd, err := openNativeTun(t.name)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}
	t.fd = fd

	if t.mtu > 0 {
		if err := setMTU(t.name, t.mtu); err != nil {
			t.fd.Close()
			return nil, fmt.Errorf("set mtu: %w", err)
		}
	}

	return t, nil
}

func openNativeTun(name string) (*os.File, error) {
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/net/tun: %w", err)
	}

	var ifr [unix.IFNAMSIZ]byte
	copy(ifr[:], name)
	ifr[unix.IFNAMSIZ-1] = 0

	flags := uint16(unix.IFF_TUN | unix.IFF_NO_PI)
	ifr[unix.IFNAMSIZ-2] = byte(flags)

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr[0])),
	)
	if errno != 0 {
		unix.Close(fd)
		return nil, fmt.Errorf("TUNSETIFF: %w", errno)
	}

	return os.NewFile(uintptr(fd), "/dev/net/tun"), nil
}

func setMTU(name string, n uint32) error {
	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)
	if err != nil {
		return err
	}

	defer unix.Close(fd)

	ifr, err := unix.NewIfreq(name)
	if err != nil {
		return err
	}
	ifr.SetUint32(n)
	return unix.IoctlIfreq(fd, unix.SIOCSIFMTU, ifr)
}

func (t *TUN) Name() string {
	return t.name
}

func (t *TUN) Type() string {
	return Driver
}

func (t *TUN) Read(buf []byte) (int, error) {
	return t.fd.Read(buf)
}

func (t *TUN) Write(buf []byte) (int, error) {
	return t.fd.Write(buf)
}

func (t *TUN) Close() error {
	return t.fd.Close()
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
