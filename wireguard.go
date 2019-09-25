package main

import (
	"errors"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	defaultPersistentKeepaliveInterval = 25 * time.Second
	defaultWireguardDeviceName         = "wg0"
)

func newPeerConfig(publicKey string, presharedKey string, endpoint string, allowedIPs []string) (*wgtypes.PeerConfig, error) {
	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return nil, err
	}
	t := defaultPersistentKeepaliveInterval
	peer := &wgtypes.PeerConfig{PublicKey: key, PersistentKeepaliveInterval: &t}
	if presharedKey != "" {
		key, err := wgtypes.ParseKey(presharedKey)
		if err != nil {
			return nil, err
		}
		peer.PresharedKey = &key
	}
	if endpoint != "" {
		addr, err := net.ResolveUDPAddr("udp4", endpoint)
		if err != nil {
			return nil, err
		}
		peer.Endpoint = addr
	}
	for _, ai := range allowedIPs {
		_, network, err := net.ParseCIDR(ai)
		if err != nil {
			return nil, err
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *network)
	}
	return peer, nil
}

func setPeers(deviceName string, peers []wgtypes.PeerConfig) error {
	wg, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer func() {
		if err := wg.Close(); err != nil {
			log.Printf("Failed to close wireguard client: %v", err)
		}
	}()
	if deviceName == "" {
		deviceName = defaultWireguardDeviceName
	}
	device, err := wg.Device(deviceName)
	if err != nil {
		return err
	}
	for _, ep := range device.Peers {
		found := false
		for _, np := range peers {
			if ep.PublicKey.String() == np.PublicKey.String() {
				found = true
				break
			}
		}
		if !found {
			peers = append(peers, wgtypes.PeerConfig{PublicKey: ep.PublicKey, Remove: true})
		}
	}
	return wg.ConfigureDevice(deviceName, wgtypes.Config{Peers: peers})
}

func addNetlinkRoute() error {
	link, err := netlink.LinkByName(defaultWireguardDeviceName)
	if err != nil {
		return err
	}
	err = netlink.RouteAdd(&netlink.Route{LinkIndex: link.Attrs().Index, Dst: userPeerSubnet})
	if errors.Is(err, syscall.EEXIST) {
		log.Printf("Could not add route: %v", err)
		return nil
	}
	return err
}
