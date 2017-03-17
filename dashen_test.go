package dashen_test

import (
	"testing"
	"time"

	"github.com/minodisk/dashen"
)

func TestListen(t *testing.T) {
	d := dashen.New()

	errCh1 := make(chan error, 1)
	go func(d *dashen.Dashen) {
		if err := d.Listen(); err != nil {
			errCh1 <- err
		}
	}(d)

	errCh2 := make(chan error, 1)
	go func(d *dashen.Dashen) {
		time.Sleep(time.Second)
		if err := d.Listen(); err != nil {
			errCh2 <- err
		}
	}(d)

	select {
	case err := <-errCh1:
		t.Fatal("1st time Listen shouldn't return error:", err)
	case err := <-errCh2:
		if err != dashen.AlreadyListening {
			t.Errorf("got %v, want %v", err, dashen.AlreadyListening)
		}
	}
}

func TestClose(t *testing.T) {
	t.Run("close unlistening", func(t *testing.T) {
		d := dashen.New()

		errCh := make(chan error, 1)
		go func(d *dashen.Dashen) {
			time.Sleep(time.Second * 3)
			if err := d.Close(); err != nil {
				errCh <- err
				return
			}
		}(d)
		if err := <-errCh; err == nil {
			t.Error("shouldn't be able to close before listening")
		}
	})

	t.Run("close listening", func(t *testing.T) {
		d := dashen.New()
		errCh := make(chan error, 1)
		doneCh := make(chan struct{})
		go func(d *dashen.Dashen) {
			if err := d.Listen(); err != nil {
				errCh <- err
				return
			}
			doneCh <- struct{}{}
		}(d)
		go func(d *dashen.Dashen) {
			time.Sleep(time.Second * 3)
			if err := d.Close(); err != nil {
				errCh <- err
				return
			}
			doneCh <- struct{}{}
		}(d)
		select {
		case err := <-errCh:
			t.Errorf("should be able to close after listening: %v", err)
		case <-doneCh:
		}
	})
}

func cb1() {}
func cb2() {}

func TestSubscribe(t *testing.T) {
	mac1 := "11:11:11:11:11:11"
	mac2 := "22:22:22:22:22:22"
	d := dashen.New()

	t.Run("can subscribe unique func", func(t *testing.T) {
		d.Subscribe(mac1, cb1)
		cbs, ok := d.MACCallbacksMap[mac1]
		if !ok {
			t.Errorf("not registered MAC: %s", mac1)
		}

		got := len(cbs)
		want := 1
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})

	t.Run("not subscribe same func", func(t *testing.T) {
		d.Subscribe(mac1, cb1)
		cbs, ok := d.MACCallbacksMap[mac1]
		if !ok {
			t.Errorf("not registered MAC: %s", mac1)
		}

		got := len(cbs)
		want := 1
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})

	t.Run("can subscribe different func", func(t *testing.T) {
		d.Subscribe(mac1, cb2)
		cbs, ok := d.MACCallbacksMap[mac1]
		if !ok {
			t.Errorf("not registered MAC: %s", mac1)
		}

		got := len(cbs)
		want := 2
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})

	t.Run("can subscribe different MAC", func(t *testing.T) {
		d.Subscribe(mac2, cb1)
		d.Subscribe(mac2, cb2)
		cbs, ok := d.MACCallbacksMap[mac2]
		if !ok {
			t.Errorf("not registered MAC: %s", mac2)
		}

		got := len(cbs)
		want := 2
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})
}

func TestUnsubscribe(t *testing.T) {
	mac1 := "11:11:11:11:11:11"
	mac2 := "22:22:22:22:22:22"
	d := dashen.New()

	d.Subscribe(mac1, cb1)
	d.Subscribe(mac1, cb2)
	d.Subscribe(mac2, cb1)
	d.Subscribe(mac2, cb2)

	t.Run("can't unsubscribe unknown MAC", func(t *testing.T) {
		d.Unsubscribe("33:33:33:33:33:33", cb1)
		got := len(d.MACCallbacksMap[mac1])
		want := 2
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})

	t.Run("can unsubscribe", func(t *testing.T) {
		d.Unsubscribe(mac1, cb1)
		got := len(d.MACCallbacksMap[mac1])
		want := 1
		if got != want {
			t.Errorf("number of callbacks: got %d, want %d", got, want)
		}
	})
}
