package main

import (
	"log"

	"github.com/minodisk/dashen"
)

func main() {
	d := dashen.New()
	d.Subscribe("f0:d5:bf:66:7f:47", func() {
		log.Println("detected")
	})
	if err := d.Listen(); err != nil {
		panic(err)
	}
}
