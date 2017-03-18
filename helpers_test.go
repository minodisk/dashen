package dashen_test

import (
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
			&dashen.Skip{"docker"},
		},
		{
			pcap.Interface{
				Name: "br-XXXXXX",
			},
			&dashen.Skip{"bridge"},
		},
		{
			pcap.Interface{
				Name: "good-name",
			},
			&dashen.Skip{"bad address"},
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP: net.IPv4(127, 0, 0, 1),
					},
				},
			},
			&dashen.Skip{"bad address"},
		},
		{
			pcap.Interface{
				Name: "good-name",
				Addresses: []pcap.InterfaceAddress{
					{
						IP: net.IPv4(10, 0, 0, 1),
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
