package dashen_test

import (
	"testing"
	"time"

	"github.com/minodisk/dashen"
)

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
