package main

import (
	"log"

	"github.com/ioben/btops/config"
	"github.com/ioben/btops/handlers"
	"github.com/ioben/btops/ipc"
	"github.com/ioben/btops/monitors"
)

func main() {
	for {
		listen()
	}
}

func listen() {
	c, err := config.GetConfig()
	if err != nil {
		log.Fatal("Unable to get config", err)
	}

	log.Println("Config: ", c)

	handlers := handlers.NewHandlers(c)

	sub, err := ipc.NewSubscriber()
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()

	for !c.ConfigChanged() && sub.Scanner.Scan() {
		monitors, err := monitors.GetMonitors()
		if err != nil {
			log.Println("Unable to obtain monitors:", err)
		}

		handlers.Handle(monitors)
	}
}
