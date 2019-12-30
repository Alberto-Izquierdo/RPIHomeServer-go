package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/message_generator"
)

var mainExitChannel = make(chan bool)

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
	defer gpio_manager.ClearAllPins()
	/*
		fmt.Println("Configuration loaded, connecting to gRPC server")
		client, connection, err := grpc.ConnectToGrpcServer(config)
		if err != nil {
			fmt.Println("There was an error connecting to the gRPC server: ", err)
			return
		}
		defer connection.Close()
		err = grpc.RegisterPinsToGRPCServer(client, config)
		if err != nil {
			fmt.Println("There was an error with the gRPC registration: ", err)
			return
		}
	*/
	actionsChannel := make(chan configuration_loader.Action)
	var exitChannels []chan bool
	if len(config.AutomaticMessages) > 0 {
		exitChannels = append(exitChannels, make(chan bool))
		go message_generator.Run(config.AutomaticMessages, actionsChannel, exitChannels[len(exitChannels)-1])
	}
	fmt.Println("Waiting for messages")
	var exit = false
	for !exit {
		select {
		case action := <-actionsChannel:
			gpio_manager.SetPinState(action.Pin, action.State)
		case exit = <-mainExitChannel:
			for _, exitChannel := range exitChannels {
				exitChannel <- true
			}
		}
	}
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
		mainExitChannel <- true
	}()
}
