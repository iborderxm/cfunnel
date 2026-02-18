package engine

import (
	"github.com/fmnx/cftun/client/tun/buffer"
	"github.com/fmnx/cftun/client/tun/dialer"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/fmnx/cftun/client/tun/core/device"
	"github.com/fmnx/cftun/client/tun/log"
	"github.com/fmnx/cftun/client/tun/native"
	"github.com/fmnx/cftun/client/tun/proxy"
	"github.com/fmnx/cftun/client/tun/tunnel"
)

var (
	Mu sync.Mutex

	// Device holds the default device for the engine.
	Device device.Device

	// NativeStack holds the native network stack for the engine.
	NativeStack *native.NativeStack

	ArgoProxy *proxy.Argo
)

// Stop shuts the default engine down.
func Stop() {
	if err := stop(); err != nil {
		log.Fatalf("[ENGINE] failed to stop: %v", err)
	}
}

func stop() (err error) {
	Mu.Lock()
	if Device != nil {
		Device.Close()
	}
	if ArgoProxy != nil {
		go ArgoProxy.Close()
	}
	if NativeStack != nil {
		NativeStack.Stop()
	}
	Mu.Unlock()
	return nil
}

func HandleNetStack(argoProxy *proxy.Argo, device, interfaceName, logLevel string, mtu int) (err error) {
	ArgoProxy = argoProxy
	buffer.RelayBufferSize = mtu
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logger, err := log.NewLeveled(level)
	if err != nil {
		return err
	}
	log.SetLogger(logger)

	if interfaceName != "" {
		iface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return err
		}
		dialer.DefaultInterfaceName = iface.Name
		atomic.StoreInt32(&dialer.DefaultInterfaceIndex, int32(iface.Index))
		log.Infof("[DIALER] bind to interface: %s", interfaceName)
	}

	transport := tunnel.New(argoProxy)
	transport.ProcessAsync()

	if Device, err = parseDevice(device, uint32(mtu)); err != nil {
		log.Fatalf(err.Error(), "\n")
		return
	}

	NativeStack = native.New(Device, transport, mtu)
	if err := NativeStack.Start(); err != nil {
		log.Fatalf("[NATIVE] failed to start native stack: %v", err)
		return err
	}

	log.Infof(
		"[NATIVE] %s://%s <-> %s -> %s",
		Device.Type(), Device.Name(),
		argoProxy.Host(), argoProxy.Addr(),
	)
	return nil
}
