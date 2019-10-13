package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

var exit = false

func main() {
	setupKeyboardSignal()
	fmt.Println("Reading configuration from file")
	filepath := flag.String("path", "", "path to the configuration file")
	flag.Parse()
	if *filepath == "" {
		fmt.Println("Path argument can not be empty")
		return
	}
	config, err := configuration_loader.LoadConfigurationFromPath(*filepath)
	if err != nil {
		fmt.Println("There was an error parsing the configuration file: ", err)
		return
	}
	err = gpio_manager.Setup(config.PinsActive)
	if err != nil {
		fmt.Println("There was an error setting up the gpio manager: ", err)
		return
	}
	fmt.Println("Configuration loaded, connecting to gRPC server")
	// TODO: connect to gRPC server
	fmt.Println("Waiting for messages")
	for !exit {
		// TODO: handle messages
	}
	gpio_manager.ClearAllPins()
	fmt.Println("Done!")
}

func setupKeyboardSignal() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		fmt.Println("Ctrl+C captured, cleaning up")
		exit = true
	}()
}