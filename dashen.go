package dashen

import (
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
)

type Dashen struct {
	Logger          Logger
	MACCallbacksMap MACCallbacksMap
	doneCh          chan struct{}
	mutex           *sync.Mutex
	listening       bool
}

func New() *Dashen {
	return &Dashen{
		MACCallbacksMap: MACCallbacksMap{},
		mutex:           new(sync.Mutex),
		doneCh:          make(chan struct{}),
	}
}

func (d *Dashen) Listen() error {
	d.mutex.Lock()
	if d.listening {
		d.mutex.Unlock()
		return errors.New("already listening")
	}
	d.listening = true
	d.mutex.Unlock()

	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func(errCh chan error) {
		for {
			select {
			case <-d.doneCh:
				return
			case err := <-errCh:
				d.Println(err)
			}
		}
	}(errCh)

	var wg sync.WaitGroup
	for _, iface := range ifaces {
		wg.Add(1)
		go func(iface pcap.Interface) {
			defer func() {
				wg.Done()
			}()
			if err := validate(iface); err != nil {
				errCh <- errors.Wrap(err, "skip to open interface")
				return
			}
			if err := d.open(iface, errCh); err != nil {
				errCh <- errors.Wrap(err, "fail to open interface with error")
				return
			}
		}(iface)
	}
	wg.Wait()

	return nil
}

func (d *Dashen) open(iface pcap.Interface, errCh chan error) error {
	d.Println("open interface:", iface.Name)
	handle, err := pcap.OpenLive(iface.Name, 1024, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	// handle.Close() blocks this goroutine
	// https://github.com/google/gopacket/issues/253
	// defer handle.Close()

	d.Println("listen:", handle.LinkType().String(), iface.Name)
	src := gopacket.NewPacketSource(handle, handle.LinkType())
	in := src.Packets()
	for {
		select {
		case <-d.doneCh:
			return nil
		case packets := <-in:
			layer := packets.Layer(layers.LayerTypeEthernet)
			if layer == nil {
				continue
			}
			ethernet, ok := layer.(*layers.Ethernet)
			if !ok {
				continue
			}
			srcMac := ethernet.SrcMAC.String()
			for mac, cbs := range d.MACCallbacksMap {
				if srcMac != mac {
					continue
				}
				for _, cb := range cbs {
					cb()
				}
			}
		}
	}

	return nil
}

func (d *Dashen) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.listening {
		return errors.New("not listening")
	}
	close(d.doneCh)
	d.listening = false
	return nil
}

func (d *Dashen) Subscribe(mac string, cb Callback) {
	cs, ok := d.MACCallbacksMap[mac]
	if !ok {
		d.MACCallbacksMap[mac] = []Callback{cb}
		return
	}
	for _, c := range cs {
		if &c == &cb {
			return
		}
	}
	d.MACCallbacksMap[mac] = append(cs, cb)
}

func (d *Dashen) Unsubscribe(mac string, cb Callback) {
	cs, ok := d.MACCallbacksMap[mac]
	if !ok {
		return
	}
	cbs := []Callback{}
	for _, c := range cs {
		if &c == &cb {
			continue
		}
		cbs = append(cbs, c)
	}
	d.MACCallbacksMap[mac] = cbs
}

func (d *Dashen) Println(v ...interface{}) {
	if d.Logger == nil {
		return
	}
	d.Logger.Println(v...)
}
