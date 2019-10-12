package main

import (
	"fmt"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"os"
	"os/signal"
)

var exit = false

func main() {
	setupSignal()
	fmt.Println("Reading configuration from file")
	// TODO: Read configuration from file for now it is empty
	pins := []gpio_manager.PairNamePin{}
	gpio_manager.Setup(pins)
	fmt.Println("Configuration loaded, connecting to gRPC server")
	// TODO: connect to gRPC server
	fmt.Println("Waiting for messages")
	for !exit {
		// TODO: handle messages
	}
	gpio_manager.ClearAllPins()
	fmt.Println("Done!")
}

func setupSignal() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		fmt.Println("Ctrl+C captured, cleaning up")
		exit = true
	}()
}
