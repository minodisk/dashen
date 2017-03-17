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
