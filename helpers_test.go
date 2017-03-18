package dashen_test

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/google/gopacket/pcap"
	"github.com/minodisk/dashen"
)

func TestValidate(t *testing.T) {
	for _, c := range []struct {
		iface pcap.Interface
		err   error
	}{
		{
			pcap.Interface{
				Name: "docker",
			},
			errors.New("docker"),
		},
		{
			pcap.Interface{
				Name: "br-XXXXXX",
			},
			errors.New("bridge"),
		},
		{
			pcap.Interface{
				Name: "good-name",
			},
			errors.New("bad address"),
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP: net.IP{127},
					},
				},
			},
			errors.New("bad address"),
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP:      net.IP{192},
						Netmask: net.IPMask{0x00, 0x00},
					},
				},
			},
			errors.New("bad address"),
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP:      net.IP{192},
						Netmask: net.IPMask{0xff, 0x00},
					},
				},
			},
			errors.New("bad address"),
		},
		{pcap.Interface{
			Name: "good-name",
			Addresses: []pcap.InterfaceAddress{
				{
					IP:      net.IP{192},
					Netmask: net.IPMask{0x00, 0xff},
				},
			},
		},
			errors.New("bad address"),
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP:      net.IP{192},
						Netmask: net.IPMask{0xff, 0xff},
					},
				},
			},
			nil,
		},
	} {
		t.Run(c.iface.Name, func(t *testing.T) {
			err := dashen.Filter(c.iface)
			if !reflect.DeepEqual(err, c.err) {
				t.Errorf("got %v, want %v", err, c.err)
			}
		})
	}
}
