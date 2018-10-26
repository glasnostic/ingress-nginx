package store

import (
	"encoding/binary"
	"fmt"
	"net"

	core "k8s.io/api/core/v1"
)

func updateEndpoints(endpoints *core.Endpoints, n, p *net.IPNet, shouldConvert bool) {
	for i, s := range endpoints.Subsets {
		for j, addr := range s.Addresses {
			endpoints.Subsets[i].Addresses[j].IP =
				getPhantomIP(net.ParseIP(addr.IP), n, p, shouldConvert).String()
		}
	}
}

func getPhantomIP(realIP net.IP, net, phantomNet *net.IPNet, shouldConvert bool) net.IP {
	if !shouldConvert || !net.Contains(realIP) {
		return realIP
	}
	if realIP.To4() != nil {
		ip := binary.BigEndian.Uint32(realIP[12:])
		n := binary.BigEndian.Uint32(net.IP)
		p := binary.BigEndian.Uint32(phantomNet.IP)

		phantomIPValue := (ip ^ n) | p
		phantomIP := make([]byte, 4)
		binary.BigEndian.PutUint32(phantomIP, phantomIPValue)
		return phantomIP
	}
	ip := ipv6ToUint128(realIP)
	n := ipv6ToUint128(net.IP)
	p := ipv6ToUint128(phantomNet.IP)

	phantomIPValue := ip.xor(n).or(p)
	phantomIP := make([]byte, 16)
	binary.BigEndian.PutUint64(phantomIP[:8], phantomIPValue.low)
	binary.BigEndian.PutUint64(phantomIP[8:], phantomIPValue.high)
	return phantomIP
}

func ipv6ToUint128(ip net.IP) uint128 {
	return uint128{
		low:  binary.BigEndian.Uint64(ip[:8]),
		high: binary.BigEndian.Uint64(ip[8:]),
	}
}

type uint128 struct {
	low  uint64
	high uint64
}

func (ui *uint128) xor(r uint128) *uint128 {
	ui.high ^= r.high
	ui.low ^= r.low
	return ui
}

func (ui *uint128) or(r uint128) *uint128 {
	ui.high |= r.high
	ui.low |= r.low
	return ui
}

func (ui uint128) String() string {
	return fmt.Sprintf("%064b%064b", ui.low, ui.high)
}
