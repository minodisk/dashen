package dashen

import (
	"errors"
	"strings"

	"github.com/google/gopacket/pcap"
)

func Validate(iface pcap.Interface) error {
	as := iface.Addresses
	if strings.Contains(iface.Name, "docker") {
		return errors.New("docker")
	}
	if strings.HasPrefix(iface.Name, "br-") {
		return errors.New("bridge")
	}

	ok := false
	for _, a := range as {
		if a.IP[0] == 127 {
			continue
		}
		if a.Netmask[0] != 0xff || a.Netmask[1] != 0xff {
			continue
		}
		ok = true
		break
	}
	if !ok {
		return errors.New("bad address")
	}
	return nil
}
