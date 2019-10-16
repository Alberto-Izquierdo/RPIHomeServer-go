package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc"
)

var exit = false

func main() {
	setupKeyboardSignal()
	config, err := loadConfiguration()
	if err != nil {
		fmt.Println("There was an error parsing the configuration file: ", err)
		return
	}
	err = gpio_manager.Setup(config.PinsActive)
	if err != nil {
		fmt.Println("There was an error setting up the GPIO manager: ", err)
		return
	}
	fmt.Println("Configuration loaded, connecting to gRPC server")
	err = grpc.ConnectToGrpcServer(config)
	if err != nil {
		fmt.Println("There was an error connecting to gRPC server: ", err)
		return
	}
	fmt.Println("Waiting for messages")
	for !exit {
		// TODO: handle messages
	}
	gpio_manager.ClearAllPins()
	fmt.Println("Done!")
}

func loadConfiguration() (config configuration_loader.InitialConfiguration, err error) {
	filepath := flag.String("path", "", "path to the configuration file")
	flag.Parse()
	if *filepath == "" {
		return config, errors.New("\"path\" argument can not be empty")
	}
	fmt.Println("Reading configuration from file")
	return configuration_loader.LoadConfigurationFromPath(*filepath)
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
